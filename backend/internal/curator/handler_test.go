package curator

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
	CuratorService
	filter WorkspaceFilter
	target NotifyTarget
	err    error
}

func (f *fakeService) ListWorkspaces(_ context.Context, filter WorkspaceFilter) (WorkspacePage, error) {
	f.filter = filter
	return WorkspacePage{Items: []WorkspaceRow{{Severity: SeverityRed, Signals: []Signal{}}}, Total: 1}, f.err
}

func (f *fakeService) NotifyWorkspace(_ context.Context, _ uuid.UUID, target NotifyTarget, _ string) error {
	f.target = target
	return f.err
}

func (f *fakeService) SetVerification(context.Context, uuid.UUID, VerificationDecision, string) (Profile, error) {
	return Profile{}, f.err
}

func serve(t *testing.T, svc CuratorService, role user.Role, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(svc, middleware.NewAuth(tokens), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)

	token, _, _ := tokens.IssueAccessToken(uuid.New(), role)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHandlerWorkspaces(t *testing.T) {
	svc := &fakeService{}

	rec := serve(t, svc, user.Curator, http.MethodGet, "/curator/workspaces?sort=activity&has_signals=true&q=Анна", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.filter.Sort != SortActivity || svc.filter.HasSignals == nil || !*svc.filter.HasSignals || svc.filter.Query != "Анна" {
		t.Fatalf("filter = %+v", svc.filter)
	}
	for _, field := range []string{`"counts"`, `"severity":"red"`, `"signals":[]`} {
		if !strings.Contains(rec.Body.String(), field) {
			t.Fatalf("body %s has no %s", rec.Body, field)
		}
	}

	if rec := serve(t, svc, user.Mentor, http.MethodGet, "/curator/workspaces", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("mentor: status = %d", rec.Code)
	}

	rec = serve(t, svc, user.Curator, http.MethodPost, "/curator/workspaces/"+uuid.NewString()+"/notify", `{"target":"both","message":"Привет"}`)
	if rec.Code != http.StatusNoContent || svc.target != TargetBoth {
		t.Fatalf("notify: status = %d, target = %s", rec.Code, svc.target)
	}
}

func TestHandlerVerificationErrors(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{ErrNotPending, http.StatusConflict},
		{ErrCommentRequired, http.StatusBadRequest},
		{ErrProfileNotFound, http.StatusNotFound},
	}
	for _, tt := range tests {
		rec := serve(t, &fakeService{err: tt.err}, user.Curator, http.MethodPatch,
			"/curator/profiles/"+uuid.NewString()+"/verification", `{"status":"rejected"}`)
		if rec.Code != tt.status {
			t.Fatalf("%v: status = %d, want %d", tt.err, rec.Code, tt.status)
		}
	}
}
