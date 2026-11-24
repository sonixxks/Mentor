package chat

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type fakeService struct {
	before *time.Time
	limit  int
	sent   SendInput
}

func (f *fakeService) History(_ context.Context, _ Actor, _ uuid.UUID, before *time.Time, limit int) ([]Message, error) {
	f.before, f.limit = before, limit
	return nil, nil
}

func (f *fakeService) Send(_ context.Context, _ Actor, _ uuid.UUID, in SendInput) (Message, error) {
	f.sent = in
	text := in.Text
	return Message{ID: uuid.New(), Type: TypeText, Text: &text}, nil
}

func (f *fakeService) MarkRead(context.Context, Actor, uuid.UUID) error { return nil }

func TestHandler(t *testing.T) {
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	svc := &fakeService{}
	mux := http.NewServeMux()
	NewHandler(svc, middleware.NewAuth(tokens), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)

	do := func(role user.Role, method, target, body string) *httptest.ResponseRecorder {
		token, _, _ := tokens.IssueAccessToken(uuid.New(), role)
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	base := "/workspaces/" + uuid.NewString() + "/messages"

	rec := do(user.Student, http.MethodGet, base+"?before=2026-11-20T10:00:00%2B05:00&limit=20", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" || svc.limit != 20 || svc.before == nil {
		t.Fatalf("history: status = %d, body = %s, before = %v", rec.Code, rec.Body, svc.before)
	}
	if rec := do(user.Student, http.MethodGet, base+"?before=yesterday", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad before: status = %d", rec.Code)
	}

	rec = do(user.Mentor, http.MethodPost, base, `{"text":"Привет"}`)
	if rec.Code != http.StatusCreated || svc.sent.Text != "Привет" || !strings.Contains(rec.Body.String(), `"file":null`) {
		t.Fatalf("send: status = %d, body = %s", rec.Code, rec.Body)
	}

	if rec := do(user.Curator, http.MethodGet, base, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("curator reads chat: status = %d", rec.Code)
	}
	if rec := do(user.Student, http.MethodPost, base+"/read", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("read: status = %d", rec.Code)
	}
}
