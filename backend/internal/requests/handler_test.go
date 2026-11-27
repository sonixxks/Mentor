package requests

import (
	"context"
	"encoding/json"
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
	err        error
	gotCreate  CreateInput
	gotChange  ChangeStatusInput
	gotGroup   Group
	gotLimit   int
	gotOffset  int
	gotStatus  *Status
	returnView View
}

func (f *fakeService) Create(_ context.Context, _ Actor, in CreateInput) (View, error) {
	f.gotCreate = in
	return f.returnView, f.err
}

func (f *fakeService) CreateByCurator(context.Context, Actor, CuratorCreateInput) (View, error) {
	return f.returnView, f.err
}

func (f *fakeService) ListMine(_ context.Context, _ Actor, g Group, limit, offset int) ([]View, int, error) {
	f.gotGroup, f.gotLimit, f.gotOffset = g, limit, offset
	return []View{f.returnView}, 11, f.err
}

func (f *fakeService) ListAll(_ context.Context, _ Actor, s *Status, _, _ int) ([]View, int, error) {
	f.gotStatus = s
	return []View{f.returnView}, 1, f.err
}

func (f *fakeService) ChangeStatus(_ context.Context, _ Actor, in ChangeStatusInput) (View, error) {
	f.gotChange = in
	return f.returnView, f.err
}

type testServer struct {
	mux    *http.ServeMux
	tokens *auth.TokenManager
}

func newTestServer(t *testing.T, svc RequestService) testServer {
	t.Helper()
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(svc, middleware.NewAuth(tokens), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return testServer{mux: mux, tokens: tokens}
}

func (s testServer) do(t *testing.T, role user.Role, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	token, _, err := s.tokens.IssueAccessToken(uuid.New(), role)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		ErrorCode string `json:"error_code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}
	return body.ErrorCode
}

func TestHandlerCreate(t *testing.T) {
	svc := &fakeService{returnView: View{Request: Request{ID: requestID, Status: StatusPending}}}
	srv := newTestServer(t, svc)

	mentorID := uuid.New()
	rec := srv.do(t, user.Student, http.MethodPost, "/requests",
		`{"mentor_id":"`+mentorID.String()+`","cover_letter":"Привет"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.gotCreate.MentorID != mentorID || svc.gotCreate.CoverLetter != "Привет" {
		t.Fatalf("service input = %+v", svc.gotCreate)
	}

	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "pending" || resp["id"] != requestID.String() {
		t.Fatalf("response = %v", resp)
	}
	if _, ok := resp["student"].(map[string]any)["last_name"]; ok {
		t.Fatal("student last name must not be exposed in requests")
	}

	if rec := srv.do(t, user.Student, http.MethodPost, "/requests", `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing mentor_id: status = %d", rec.Code)
	}
	if rec := srv.do(t, user.Student, http.MethodPost, "/requests", `{"mentor_id":"x"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad uuid: status = %d", rec.Code)
	}
	if rec := srv.do(t, user.Mentor, http.MethodPost, "/requests", `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("mentor creates: status = %d", rec.Code)
	}
}

func TestHandlerErrorMapping(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{ErrRequestLimit, http.StatusConflict, "REQUEST_LIMIT_REACHED"},
		{ErrDuplicate, http.StatusConflict, "DUPLICATE_REQUEST"},
		{ErrMentorLimit, http.StatusConflict, "MENTOR_LIMIT_REACHED"},
		{ErrInvalidTransition, http.StatusConflict, "INVALID_TRANSITION"},
		{ErrStatusConflict, http.StatusConflict, "CONFLICT"},
		{ErrReasonRequired, http.StatusBadRequest, "VALIDATION_ERROR"},
		{ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{context.DeadlineExceeded, http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			srv := newTestServer(t, &fakeService{err: tt.err})
			rec := srv.do(t, user.Student, http.MethodPatch, "/requests/"+requestID.String()+"/status", `{"status":"cancelled"}`)
			if rec.Code != tt.status || errorCode(t, rec) != tt.code {
				t.Fatalf("got %d %s, want %d %s", rec.Code, errorCode(t, rec), tt.status, tt.code)
			}
		})
	}
}

func TestHandlerListMine(t *testing.T) {
	svc := &fakeService{}
	srv := newTestServer(t, svc)

	rec := srv.do(t, user.Mentor, http.MethodGet, "/requests/my?group=archive&page=2&limit=5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.gotGroup != GroupArchive || svc.gotLimit != 5 || svc.gotOffset != 5 {
		t.Fatalf("group = %q, limit = %d, offset = %d", svc.gotGroup, svc.gotLimit, svc.gotOffset)
	}

	var page struct {
		Total int `json:"total"`
		Page  int `json:"page"`
		Items []any
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &page)
	if page.Total != 11 || page.Page != 2 || len(page.Items) != 1 {
		t.Fatalf("page = %+v", page)
	}

	if rec := srv.do(t, user.Mentor, http.MethodGet, "/requests/my?group=old", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad group: status = %d", rec.Code)
	}
	if rec := srv.do(t, user.Curator, http.MethodGet, "/requests/my", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("curator on /requests/my: status = %d", rec.Code)
	}
}

func TestHandlerCuratorQueue(t *testing.T) {
	svc := &fakeService{returnView: View{StudentFull: Person{LastName: "Петрова"}}}
	srv := newTestServer(t, svc)

	rec := srv.do(t, user.Curator, http.MethodGet, "/curator/requests?status=pending", "")
	if rec.Code != http.StatusOK || svc.gotStatus == nil || *svc.gotStatus != StatusPending {
		t.Fatalf("status = %d, filter = %v", rec.Code, svc.gotStatus)
	}
	if !strings.Contains(rec.Body.String(), `"student_full"`) || !strings.Contains(rec.Body.String(), "Петрова") {
		t.Fatalf("curator must see the full student: %s", rec.Body)
	}

	if rec := srv.do(t, user.Mentor, http.MethodGet, "/curator/requests", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("mentor on curator queue: status = %d", rec.Code)
	}
}
