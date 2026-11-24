package requests

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type RequestService interface {
	Create(ctx context.Context, actor Actor, in CreateInput) (View, error)
	CreateByCurator(ctx context.Context, actor Actor, in CuratorCreateInput) (View, error)
	ListMine(ctx context.Context, actor Actor, group Group, limit, offset int) ([]View, int, error)
	ListAll(ctx context.Context, actor Actor, status *Status, limit, offset int) ([]View, int, error)
	ChangeStatus(ctx context.Context, actor Actor, in ChangeStatusInput) (View, error)
}

type Handler struct {
	svc  RequestService
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(svc RequestService, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /requests", h.auth.Require(h.create, user.Student))
	mux.Handle("GET /requests/my", h.auth.Require(h.listMine, user.Student, user.Mentor))
	mux.Handle("PATCH /requests/{id}/status", h.auth.Require(h.changeStatus, user.Student, user.Mentor, user.Curator))

	mux.Handle("GET /curator/requests", h.auth.Require(h.listForCurator, user.Curator))
	mux.Handle("POST /curator/requests", h.auth.Require(h.createByCurator, user.Curator))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body createRequestBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}
	if body.MentorID == uuid.Nil {
		httpx.SendBadRequest(w, "validation error", "mentor_id: field is required")
		return
	}

	view, err := h.svc.Create(r.Context(), actorFrom(r), CreateInput{
		MentorID:    body.MentorID,
		PackageID:   body.ServicePackageID,
		CoverLetter: body.CoverLetter,
	})
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusCreated, toRequestJSON(view))
}

func (h *Handler) createByCurator(w http.ResponseWriter, r *http.Request) {
	var body curatorCreateBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	var details []string
	if body.StudentID == uuid.Nil {
		details = append(details, "student_id: field is required")
	}
	if body.MentorID == uuid.Nil {
		details = append(details, "mentor_id: field is required")
	}
	if len(details) > 0 {
		httpx.SendBadRequest(w, "validation error", details...)
		return
	}

	view, err := h.svc.CreateByCurator(r.Context(), actorFrom(r), CuratorCreateInput{
		StudentID: body.StudentID,
		MentorID:  body.MentorID,
		PackageID: body.ServicePackageID,
	})
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusCreated, toCuratorRequestJSON(view))
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := httpx.ParsePage(q, 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, "validation error", err.Error())
		return
	}
	p := httpx.NewQueryParser(q)
	group := Group(p.Enum("group", "", string(GroupActive), string(GroupArchive)))
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, "validation error", errs...)
		return
	}

	views, total, err := h.svc.ListMine(r.Context(), actorFrom(r), group, page.Limit, page.Offset())
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]requestJSON, len(views))
	for i, v := range views {
		items[i] = toRequestJSON(v)
	}
	httpx.SendJSON(w, http.StatusOK, pageJSON[requestJSON]{Total: total, Page: page.Number, Items: items})
}

func (h *Handler) listForCurator(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := httpx.ParsePage(q, 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, "validation error", err.Error())
		return
	}

	var status *Status
	if raw := q.Get("status"); raw != "" {
		s := Status(raw)
		status = &s
	}

	views, total, err := h.svc.ListAll(r.Context(), actorFrom(r), status, page.Limit, page.Offset())
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]curatorRequestJSON, len(views))
	for i, v := range views {
		items[i] = toCuratorRequestJSON(v)
	}
	httpx.SendJSON(w, http.StatusOK, pageJSON[curatorRequestJSON]{Total: total, Page: page.Number, Items: items})
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.SendBadRequest(w, "validation error", err.Error())
		return
	}

	var body statusUpdateBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	view, err := h.svc.ChangeStatus(r.Context(), actorFrom(r), ChangeStatusInput{
		RequestID:       id,
		To:              body.Status,
		RejectionReason: body.RejectionReason,
	})
	if err != nil {
		h.sendError(w, err)
		return
	}

	if actorFrom(r).Role == user.Curator {
		httpx.SendJSON(w, http.StatusOK, toCuratorRequestJSON(view))
		return
	}
	httpx.SendJSON(w, http.StatusOK, toRequestJSON(view))
}

func (h *Handler) sendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		httpx.SendAPIError(w, http.StatusForbidden, httpx.CodeForbidden, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrMentorNotFound),
		errors.Is(err, ErrPackageNotFound), errors.Is(err, ErrStudentNotFound):
		httpx.SendAPIError(w, http.StatusNotFound, httpx.CodeNotFound, err.Error())
	case errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrReasonRequired), errors.Is(err, ErrCoverLetterLength):
		httpx.SendBadRequest(w, err.Error())
	case errors.Is(err, ErrInvalidTransition):
		httpx.SendAPIError(w, http.StatusConflict, httpx.CodeInvalidTransition, err.Error())
	case errors.Is(err, ErrRequestLimit):
		httpx.SendAPIError(w, http.StatusConflict, httpx.CodeRequestLimitReached, err.Error())
	case errors.Is(err, ErrDuplicate):
		httpx.SendAPIError(w, http.StatusConflict, httpx.CodeDuplicateRequest, err.Error())
	case errors.Is(err, ErrMentorLimit):
		httpx.SendAPIError(w, http.StatusConflict, httpx.CodeMentorLimitReached, err.Error())
	case errors.Is(err, ErrStatusConflict), errors.Is(err, ErrNoHelpRequest):
		httpx.SendAPIError(w, http.StatusConflict, httpx.CodeConflict, err.Error())
	default:
		h.log.Error("requests handler", "error", err)
		httpx.SendInternalError(w)
	}
}

func actorFrom(r *http.Request) Actor {
	p, _ := middleware.PrincipalFrom(r.Context())
	return Actor{UserID: p.UserID, Role: p.Role}
}
