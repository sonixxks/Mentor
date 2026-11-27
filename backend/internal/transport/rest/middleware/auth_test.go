package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/platform/authctx"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTokens(t *testing.T) *auth.TokenManager {
	t.Helper()
	tm, err := auth.NewTokenManager(testSecret, time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestRequire(t *testing.T) {
	tokens := newTokens(t)
	mw := NewAuth(tokens)

	userID := uuid.New()
	studentToken, _, _ := tokens.IssueAccessToken(userID, user.Student)

	var gotPrincipal authctx.Principal
	handler := mw.Require(func(w http.ResponseWriter, r *http.Request) {
		gotPrincipal, _ = authctx.PrincipalFrom(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}, user.Student, user.Mentor)

	curatorOnly := mw.Require(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, user.Curator)

	tests := []struct {
		name    string
		handler http.Handler
		header  string
		want    int
	}{
		{"no header", handler, "", http.StatusUnauthorized},
		{"wrong scheme", handler, "Basic " + studentToken, http.StatusUnauthorized},
		{"garbage token", handler, "Bearer abc", http.StatusUnauthorized},
		{"valid token", handler, "Bearer " + studentToken, http.StatusNoContent},
		{"wrong role", curatorOnly, "Bearer " + studentToken, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			tt.handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}

	if gotPrincipal.UserID != userID || gotPrincipal.Role != user.Student {
		t.Fatalf("principal = %+v, want %s student", gotPrincipal, userID)
	}
}

func TestOptional(t *testing.T) {
	mw := NewAuth(newTokens(t))

	var authed bool
	handler := mw.Optional(func(w http.ResponseWriter, r *http.Request) {
		_, authed = PrincipalFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || authed {
		t.Fatalf("anonymous: status = %d, authed = %v", rec.Code, authed)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer broken")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("broken token: status = %d, want 401", rec.Code)
	}
}
