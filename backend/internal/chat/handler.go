package chat

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

type ChatService interface {
	History(ctx context.Context, actor Actor, workspaceID uuid.UUID, before *time.Time, limit int) ([]Message, error)
	Send(ctx context.Context, actor Actor, workspaceID uuid.UUID, in SendInput) (Message, error)
	MarkRead(ctx context.Context, actor Actor, workspaceID uuid.UUID) error
}

type Handler struct {
	svc  ChatService
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(svc ChatService, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	members := func(f http.HandlerFunc) http.Handler { return h.auth.Require(f, user.Student, user.Mentor) }

	mux.Handle("GET /workspaces/{id}/messages", members(h.history))
	mux.Handle("POST /workspaces/{id}/messages", members(h.send))
	mux.Handle("POST /workspaces/{id}/messages/read", members(h.markRead))
}

type fileJSON struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	FileName   string    `json:"file_name"`
	FileURL    string    `json:"file_url"`
	FileSizeKB *int      `json:"file_size_kb"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type messageJSON struct {
	ID        uuid.UUID   `json:"id"`
	Type      MessageType `json:"type"`
	SenderID  *uuid.UUID  `json:"sender_id"`
	Text      *string     `json:"text"`
	File      *fileJSON   `json:"file"`
	IsRead    bool        `json:"is_read"`
	CreatedAt time.Time   `json:"created_at"`
}

type sendBody struct {
	Text   string     `json:"text"`
	FileID *uuid.UUID `json:"file_id"`
}

func toMessageJSON(m Message) messageJSON {
	out := messageJSON{ID: m.ID, Type: m.Type, SenderID: m.SenderID, Text: m.Text, IsRead: m.IsRead, CreatedAt: m.CreatedAt}
	if m.File != nil {
		f := fileJSON(*m.File)
		out.File = &f
	}
	return out
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	var before *time.Time
	if raw := q.Get("before"); raw != "" {
		t, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			httpx.SendBadRequest(w, "validation error", "before: must be an ISO 8601 date-time with offset")
			return
		}
		before = &t
	}
	p := httpx.NewQueryParser(q)
	limit := DefaultPageLimit
	if l := p.Int("limit", 1, MaxPageLimit); l != nil {
		limit = *l
	}
	if errs := p.Errors(); len(errs) > 0 {
		httpx.SendBadRequest(w, "validation error", errs...)
		return
	}

	list, err := h.svc.History(r.Context(), actorFrom(r), id, before, limit)
	if err != nil {
		h.sendError(w, err)
		return
	}

	items := make([]messageJSON, len(list))
	for i, m := range list {
		items[i] = toMessageJSON(m)
	}
	httpx.SendJSON(w, http.StatusOK, items)
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}
	var body sendBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}

	msg, err := h.svc.Send(r.Context(), actorFrom(r), id, SendInput(body))
	if err != nil {
		h.sendError(w, err)
		return
	}
	httpx.SendJSON(w, http.StatusCreated, toMessageJSON(msg))
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	id, ok := workspaceID(w, r)
	if !ok {
		return
	}
	if err := h.svc.MarkRead(r.Context(), actorFrom(r), id); err != nil {
		h.sendError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWorkspaceNotFound):
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: err.Error()})
	case errors.Is(err, ErrNotMember):
		httpx.SendError(w, http.StatusForbidden, httpx.ErrorResponse{ErrorCode: httpx.ErrForbidden, Message: err.Error()})
	case errors.Is(err, ErrEmptyMessage), errors.Is(err, ErrTextTooLong), errors.Is(err, ErrFileNotFound):
		httpx.SendBadRequest(w, err.Error())
	default:
		h.log.Error("chat handler", "error", err)
		httpx.SendInternalError(w)
	}
}

func workspaceID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusNotFound, httpx.ErrorResponse{ErrorCode: httpx.ErrNotFound, Message: ErrWorkspaceNotFound.Error()})
		return uuid.Nil, false
	}
	return id, true
}

func actorFrom(r *http.Request) Actor {
	p, _ := middleware.PrincipalFrom(r.Context())
	return Actor{UserID: p.UserID, Role: p.Role}
}
