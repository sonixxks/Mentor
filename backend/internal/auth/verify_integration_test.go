package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/platform/memory"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/platform/validation"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type inbox struct{ code string }

func (i *inbox) SendVerificationCode(_ context.Context, _ string, code string) error {
	i.code = code
	return nil
}

func TestVerifyRegistrationHTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("requires migrated TEST_DB_DSN")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx := postgres.NewTransactionManager(pool)
	tokens, err := auth.NewTokenManager(strings.Repeat("x", 32), 15*time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := auth.NewDefaultPasswordHasher()
	if err != nil {
		t.Fatal(err)
	}
	mail := &inbox{}
	registrations := memory.NewRegistrationStore()
	svc := auth.NewService(auth.ServiceDeps{Users: user.NewRepository(tx), Profiles: profile.NewRepository(tx), Transactions: tx, Tokens: tokens, Hasher: hasher, RefreshTokens: memory.NewRefreshTokenStore(), Registrations: registrations, ResendLimiter: memory.NewResendLimiter(), Mailer: mail, Codes: auth.RandomCodeGenerator{}}, auth.Settings{CodeSettings: auth.CodeSettings{TTL: 10 * time.Minute, ResendInterval: time.Minute, MaxAttempts: 3, HashSecret: []byte("test")}, RefreshTokenSettings: auth.NewRefreshTokenSettings(time.Hour, time.Hour)})
	router := rest.NewRouter([]rest.ServiceHandler{auth.NewHandler(validation.NewValidator(), slog.Default(), svc)}, "v1")
	post := func(path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/v1/auth/"+path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	for _, scenario := range []string{"parallel", "rollback", "expired", "limit"} {
		t.Run(scenario, func(t *testing.T) {
			email := uuid.NewString() + "@example.org"
			defer pool.Exec(ctx, "DELETE FROM users WHERE email=$1", email)
			w := post("register", fmt.Sprintf(`{"email":%q,"role":"student","first_name":"Иван","last_name":"Иванов","password":"testpass123","terms_accepted":true}`, strings.ToUpper(email)))
			var ttl auth.CodeSentResponse
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &ttl) != nil || ttl.ExpiresInSeconds != 600 || ttl.ResendAfterSeconds != 60 {
				t.Fatal(w.Code, w.Body)
			}
			count := func(want int) {
				t.Helper()
				var n int
				if err := pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE email=$1", email).Scan(&n); err != nil || n != want {
					t.Fatalf("users=%d err=%v", n, err)
				}
			}
			count(0)
			verify := func(code string) *httptest.ResponseRecorder {
				return post("verify-email", fmt.Sprintf(`{"email":%q,"code":%q}`, email, code))
			}
			switch scenario {
			case "expired", "rollback":
				p, err := registrations.Get(ctx, email)
				if err != nil {
					t.Fatal(err)
				}
				duration := time.Minute
				want := 500
				if scenario == "expired" {
					duration = -time.Second
					want = 400
				} else {
					p.FirstName = strings.Repeat("a", 256)
				}
				if err := registrations.Save(ctx, p, duration); err != nil {
					t.Fatal(err)
				}
				if w := verify(mail.code); w.Code != want {
					t.Fatal(w.Code, w.Body)
				}
				count(0)
			case "limit":
				wrong := "000000"
				if wrong == mail.code {
					wrong = "111111"
				}
				for _, want := range []int{400, 400, 429} {
					if w := verify(wrong); w.Code != want {
						t.Fatal(w.Code, w.Body)
					}
				}
				if w := verify(mail.code); w.Code != 429 {
					t.Fatal(w.Code, w.Body)
				}
				count(0)
			case "parallel":
				results := make(chan *httptest.ResponseRecorder, 2)
				for range 2 {
					go func() { results <- verify(mail.code) }()
				}
				a, b := <-results, <-results
				if b.Code == 201 {
					a, b = b, a
				}
				if a.Code != 201 || b.Code != 400 {
					t.Fatal(a.Code, b.Code)
				}
				var pair auth.TokenResponse
				if err := json.Unmarshal(a.Body.Bytes(), &pair); err != nil || pair.RefreshToken == "" {
					t.Fatal("missing tokens", err)
				}
				principal, err := tokens.ParseAccessToken(pair.AccessToken)
				if err != nil {
					t.Fatal(err)
				}
				protected := middleware.NewAuth(tokens).Require(func(w http.ResponseWriter, r *http.Request) {
					actor, ok := middleware.PrincipalFrom(r.Context())
					if !ok || actor != principal {
						t.Fatal("incorrect request principal")
					}
					w.WriteHeader(http.StatusNoContent)
				}, user.Student)
				request := httptest.NewRequest("GET", "/", nil)
				request.Header.Set("Authorization", "Bearer "+pair.AccessToken)
				response := httptest.NewRecorder()
				protected.ServeHTTP(response, request)
				if response.Code != http.StatusNoContent {
					t.Fatal(response.Code, response.Body)
				}
				var first string
				err = pool.QueryRow(ctx, `SELECT p.first_name FROM profiles p JOIN users u ON p.user_id=u.id WHERE u.id=$1 AND u.email=$2 AND u.consent_at IS NOT NULL`, principal.UserID, email).Scan(&first)
				if err != nil || first != "Иван" {
					t.Fatal(first, err)
				}
				if w := post("login", fmt.Sprintf(`{"email":%q,"password":"testpass123"}`, email)); w.Code != 200 {
					t.Fatal(w.Code, w.Body)
				}
			}
		})
	}
}
