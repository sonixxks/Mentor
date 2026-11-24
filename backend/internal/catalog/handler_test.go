package catalog

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
	viewer *Viewer
	filter Filter
	skills SkillQuery
}

func (f *fakeService) ListMentors(_ context.Context, v *Viewer, filter Filter) ([]MentorSummary, int, error) {
	f.viewer, f.filter = v, filter
	return []MentorSummary{{ID: uuid.New(), FirstName: "Иван"}}, 1, nil
}

func (f *fakeService) GetMentor(context.Context, *Viewer, uuid.UUID) (MentorDetail, error) {
	return MentorDetail{}, ErrMentorNotFound
}

func (f *fakeService) ListReviews(context.Context, uuid.UUID, int, int) ([]Review, int, error) {
	return nil, 0, nil
}

func (f *fakeService) ListSkills(_ context.Context, q SkillQuery) ([]string, error) {
	f.skills = q
	return nil, nil
}

func newMux(t *testing.T, svc CatalogService) (*http.ServeMux, *auth.TokenManager) {
	t.Helper()
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(svc, middleware.NewAuth(tokens), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return mux, tokens
}

func get(mux *http.ServeMux, target, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHandlerListMentors(t *testing.T) {
	svc := &fakeService{}
	mux, tokens := newMux(t, svc)

	rec := get(mux, "/mentors?skills=Go,SQL&price_max=3000&has_free_slots=true&sort=price_asc&page=2&limit=5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	f := svc.filter
	if svc.viewer != nil || len(f.Skills) != 2 || *f.PriceMax != 3000 || !f.HasFreeSlots || f.Sort != SortPriceAsc || f.Offset != 5 {
		t.Fatalf("viewer = %v, filter = %+v", svc.viewer, f)
	}
	if !strings.Contains(rec.Body.String(), `"total":1`) || strings.Contains(rec.Body.String(), `"match"`) {
		t.Fatalf("body = %s", rec.Body)
	}

	token, _, _ := tokens.IssueAccessToken(uuid.New(), user.Student)
	if rec := get(mux, "/mentors", token); rec.Code != http.StatusOK || svc.viewer == nil || svc.viewer.Role != user.Student {
		t.Fatalf("authorized: status = %d, viewer = %v", rec.Code, svc.viewer)
	}

	rec = get(mux, "/mentors?min_rating=6&sort=cheap", "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "min_rating") || !strings.Contains(rec.Body.String(), "sort") {
		t.Fatalf("validation: status = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestHandlerMentorAndSkills(t *testing.T) {
	svc := &fakeService{}
	mux, _ := newMux(t, svc)

	if rec := get(mux, "/mentors/not-a-uuid", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("bad id: status = %d", rec.Code)
	}
	if rec := get(mux, "/mentors/"+uuid.NewString(), ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing mentor: status = %d", rec.Code)
	}

	rec := get(mux, "/skills?q=py&popular=true&limit=5", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("skills: status = %d, body = %s", rec.Code, rec.Body)
	}
	if svc.skills.Query != "py" || !svc.skills.Popular || svc.skills.Limit != 5 {
		t.Fatalf("skills query = %+v", svc.skills)
	}
}
