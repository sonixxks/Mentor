package profilehttp

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
	"github.com/sonixxks/Mentor/backend/internal/platform/storage"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func TestHandler(t *testing.T) {
	pool := pgtest.Pool(t)
	files, err := storage.NewLocal(t.TempDir(), "http://test/uploads")
	if err != nil {
		t.Fatal(err)
	}
	svc := profile.NewService(profile.NewRepository(postgres.NewTransactionManager(pool)), files)

	m := uuid.New()
	fx := pgtest.NewFixtures(t, pool)
	fx.Exec(`INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, 'x', 'mentor')`, m, m.String()+"@t.local")
	fx.Exec(`INSERT INTO profiles (user_id, first_name, last_name) VALUES ($1, 'Иван', 'Смирнов')`, m)

	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(svc, middleware.NewAuth(tokens), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	token, _, _ := tokens.IssueAccessToken(m, user.Mentor)

	do := func(method, path string, body io.Reader, contentType string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(context.Background(), method, path, body)
		req.Header.Set("Authorization", "Bearer "+token)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	rec := do(http.MethodGet, "/profiles/me", nil, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"verification_status":"none"`) ||
		!strings.Contains(rec.Body.String(), `"current_skills":[]`) || !strings.Contains(rec.Body.String(), `"age":null`) {
		t.Fatalf("get: %d %s", rec.Code, rec.Body)
	}

	rec = do(http.MethodPatch, "/profiles/me", strings.NewReader(`{"job_title":"Go-разработчик","hourly_rate":1500}`), "application/json")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"hourly_rate":1500`) {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body)
	}
	rec = do(http.MethodPatch, "/profiles/me", strings.NewReader(`{"onboarding_completed":true}`), "application/json")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"VALIDATION_ERROR"`) || !strings.Contains(rec.Body.String(), "help_topics") {
		t.Fatalf("incomplete onboarding: %d %s", rec.Code, rec.Body)
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("files", "cert.png")
	_, _ = fw.Write([]byte("png"))
	_ = mw.WriteField("links", "https://linkedin.com/in/ivan")
	_ = mw.Close()

	rec = do(http.MethodPost, "/profiles/me/verification-assets", &buf, mw.FormDataContentType())
	if rec.Code != http.StatusOK || strings.Count(rec.Body.String(), `"id"`) != 2 || !strings.Contains(rec.Body.String(), `"type":"file"`) {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}

	if rec := do(http.MethodPost, "/profiles/me/verification-assets", strings.NewReader("{}"), "application/json"); rec.Code != http.StatusBadRequest {
		t.Fatalf("not multipart: %d", rec.Code)
	}
}
