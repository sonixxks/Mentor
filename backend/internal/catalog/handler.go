package catalog

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
)

type CatalogService interface {
	ListMentors(ctx context.Context, viewer *Viewer, f Filter) ([]MentorSummary, int, error)
	GetMentor(ctx context.Context, viewer *Viewer, id uuid.UUID) (MentorDetail, error)
	ListReviews(ctx context.Context, mentorID uuid.UUID, limit, offset int) ([]Review, int, error)
	ListSkills(ctx context.Context, q SkillQuery) ([]string, error)
}

type Handler struct {
	svc  CatalogService
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(svc CatalogService, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /mentors", h.auth.Optional(h.listMentors))
	mux.Handle("GET /mentors/{id}", h.auth.Optional(h.getMentor))
	mux.HandleFunc("GET /mentors/{id}/reviews", h.listReviews)
	mux.HandleFunc("GET /skills", h.listSkills)
}

func (h *Handler) listMentors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := httpx.ParsePage(q, 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, "validation error", err.Error())
		return
	}

	p := httpx.NewQueryParser(q)
	f := Filter{
		Query:     p.String("q"),
		Skills:    p.List("skills"),
		PriceMin:  p.Float("price_min", 0, 1e9),
		PriceMax:  p.Float("price_max", 0, 1e9),
		HoursMin:  p.Int("hours_min", 1, 10),
		HoursMax:  p.Int("hours_max", 1, 10),
		MinMatch:  p.Int("min_match", 0, 100),
		MinRating: p.Float("min_rating", 0, 5),
		Sort: Sort(p.Enum("sort", string(SortMatch),
			string(SortMatch), string(SortRating), string(SortPriceAsc), string(SortPriceDesc), string(SortExperience))),
		Limit:  page.Limit,
		Offset: page.Offset(),
	}
	if free := p.Bool("has_free_slots"); free != nil {
		f.HasFreeSlots = *free
	}
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, "validation error", errs...)
		return
	}

	list, total, err := h.svc.ListMentors(r.Context(), viewerFrom(r), f)
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]mentorSummaryJSON, len(list))
	for i, m := range list {
		items[i] = toSummaryJSON(m)
	}
	httpx.SendJSON(w, http.StatusOK, pageJSON[mentorSummaryJSON]{Total: total, Page: page.Number, Items: items})
}

func (h *Handler) getMentor(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.SendAPIError(w, http.StatusNotFound, httpx.CodeNotFound, ErrMentorNotFound.Error())
		return
	}

	d, err := h.svc.GetMentor(r.Context(), viewerFrom(r), id)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toDetailJSON(d))
}

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.SendAPIError(w, http.StatusNotFound, httpx.CodeNotFound, ErrMentorNotFound.Error())
		return
	}
	page, err := httpx.ParsePage(r.URL.Query(), 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, "validation error", err.Error())
		return
	}

	list, total, err := h.svc.ListReviews(r.Context(), id, page.Limit, page.Offset())
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]reviewJSON, len(list))
	for i, rv := range list {
		items[i] = toReviewJSON(rv)
	}
	httpx.SendJSON(w, http.StatusOK, pageJSON[reviewJSON]{Total: total, Page: page.Number, Items: items})
}

func (h *Handler) listSkills(w http.ResponseWriter, r *http.Request) {
	p := httpx.NewQueryParser(r.URL.Query())
	q := SkillQuery{Query: p.String("q"), Limit: 20}
	if popular := p.Bool("popular"); popular != nil {
		q.Popular = *popular
	}
	if limit := p.Int("limit", 1, 100); limit != nil {
		q.Limit = *limit
	}
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, "validation error", errs...)
		return
	}

	names, err := h.svc.ListSkills(r.Context(), q)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, nonNil(names))
}

func (h *Handler) sendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMentorNotFound):
		httpx.SendAPIError(w, http.StatusNotFound, httpx.CodeNotFound, err.Error())
	default:
		h.log.Error("catalog handler", "error", err)
		httpx.SendInternalError(w)
	}
}

func viewerFrom(r *http.Request) *Viewer {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		return nil
	}
	return &Viewer{UserID: p.UserID, Role: p.Role}
}
