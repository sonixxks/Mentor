package curator

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type CuratorService interface {
	ListWorkspaces(ctx context.Context, f WorkspaceFilter) (WorkspacePage, error)
	SetWorkspaceStatus(ctx context.Context, id uuid.UUID, to WorkspaceStatus) (WorkspaceRow, error)
	NotifyWorkspace(ctx context.Context, id uuid.UUID, target NotifyTarget, message string) error
	ListProfiles(ctx context.Context, f ProfileFilter) ([]Profile, int, error)
	ResolveHelpRequest(ctx context.Context, profileID uuid.UUID, comment string) (Profile, error)
	SetVerification(ctx context.Context, profileID uuid.UUID, decision VerificationDecision, comment string) (Profile, error)
	RequestDocuments(ctx context.Context, profileID uuid.UUID, comment string) error
}

type Handler struct {
	svc  CuratorService
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(svc CuratorService, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

// Register: все ручки только для куратора. Очередь заявок (/curator/requests)
// живёт в модуле requests.
func (h *Handler) Register(mux *http.ServeMux) {
	curatorOnly := func(f http.HandlerFunc) http.Handler { return h.auth.Require(f, user.Curator) }

	mux.Handle("GET /curator/workspaces", curatorOnly(h.listWorkspaces))
	mux.Handle("PATCH /curator/workspaces/{id}/status", curatorOnly(h.setWorkspaceStatus))
	mux.Handle("POST /curator/workspaces/{id}/notify", curatorOnly(h.notifyWorkspace))

	mux.Handle("GET /curator/profiles", curatorOnly(h.listProfiles))
	mux.Handle("PATCH /curator/profiles/{id}/help-resolved", curatorOnly(h.resolveHelp))
	mux.Handle("PATCH /curator/profiles/{id}/verification", curatorOnly(h.setVerification))
	mux.Handle("POST /curator/profiles/{id}/verification/request-documents", curatorOnly(h.requestDocuments))
}

func (h *Handler) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := httpx.ParsePage(q, 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, httpx.MsgValidation, err.Error())
		return
	}
	p := httpx.NewQueryParser(q)
	f := WorkspaceFilter{
		Query:      p.String("q"),
		Sort:       WorkspaceSort(p.Enum("sort", string(SortSeverity), string(SortSeverity), string(SortActivity))),
		HasSignals: p.Bool("has_signals"),
		Limit:      page.Limit,
		Offset:     page.Offset(),
	}
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, httpx.MsgValidation, errs...)
		return
	}

	res, err := h.svc.ListWorkspaces(r.Context(), f)
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]workspaceRowJSON, len(res.Items))
	for i, row := range res.Items {
		items[i] = toWorkspaceRowJSON(row)
	}
	httpx.SendJSON(w, http.StatusOK, workspacePageJSON{Total: res.Total, Page: page.Number, Items: items, Counts: res.Counts})
}

func (h *Handler) setWorkspaceStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body workspaceStatusBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	row, err := h.svc.SetWorkspaceStatus(r.Context(), id, body.Status)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toWorkspaceSummaryJSON(row, time.Now()))
}

func (h *Handler) notifyWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body notifyBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	if err := h.svc.NotifyWorkspace(r.Context(), id, body.Target, body.Message); err != nil {
		h.sendError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listProfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := httpx.ParsePage(q, 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, httpx.MsgValidation, err.Error())
		return
	}
	p := httpx.NewQueryParser(q)
	f := ProfileFilter{NeedsHelp: p.Bool("needs_help"), Limit: page.Limit, Offset: page.Offset()}
	if v := p.Enum("verification_status", "", "none", "pending", "verified", "rejected"); v != "" {
		f.VerificationStatus = &v
	}
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, httpx.MsgValidation, errs...)
		return
	}

	list, total, err := h.svc.ListProfiles(r.Context(), f)
	if err != nil {
		h.sendError(w, err)
		return
	}
	items := make([]profileJSON, len(list))
	for i, pr := range list {
		items[i] = toProfileJSON(pr)
	}
	httpx.SendJSON(w, http.StatusOK, profilePageJSON{Total: total, Page: page.Number, Items: items})
}

func (h *Handler) resolveHelp(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body commentBody
	if !httpx.DecodeJSON(w, r, &body, true) {
		return
	}

	profile, err := h.svc.ResolveHelpRequest(r.Context(), id, body.Comment)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toProfileJSON(profile))
}

func (h *Handler) setVerification(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body verificationBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	profile, err := h.svc.SetVerification(r.Context(), id, body.Status, body.Comment)
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusOK, toProfileJSON(profile))
}

func (h *Handler) requestDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body commentBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	if err := h.svc.RequestDocuments(r.Context(), id, body.Comment); err != nil {
		h.sendError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWorkspaceNotFound), errors.Is(err, ErrProfileNotFound):
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: err.Error()})
	case errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrInvalidTarget),
		errors.Is(err, ErrCommentRequired), errors.Is(err, ErrCommentTooLong):
		httpx.SendBadRequest(w, err.Error())
	case errors.Is(err, ErrMentorLimit):
		httpx.SendError(w, http.StatusConflict, httpx.ErrorResponse{ErrorCode: httpx.ErrMentorLimitReached, Message: err.Error()})
	case errors.Is(err, ErrNoHelpRequest), errors.Is(err, ErrNotMentor), errors.Is(err, ErrNotPending):
		httpx.SendError(w, http.StatusConflict, httpx.ErrorResponse{ErrorCode: httpx.ErrConflict, Message: err.Error()})
	default:
		h.log.Error("curator handler", "error", err)
		httpx.SendInternalError(w)
	}
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: "не найдено"})
		return uuid.Nil, false
	}
	return id, true
}
