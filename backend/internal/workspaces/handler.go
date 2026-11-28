package workspaces

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

type Lister interface {
	List(ctx context.Context, userID uuid.UUID, role user.Role, limit, offset int) ([]Summary, int, error)
}

type Handler struct {
	repo Lister
	auth *middleware.Auth
	log  *slog.Logger
}

func NewHandler(repo Lister, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{repo: repo, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /workspaces", h.auth.Require(h.list, user.Student, user.Mentor))
}

type pageJSON struct {
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Items []Summary `json:"items"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, err := httpx.ParsePage(r.URL.Query(), 10, 50)
	if err != nil {
		httpx.SendBadRequest(w, httpx.MsgValidation, err.Error())
		return
	}

	p, _ := middleware.PrincipalFrom(r.Context())
	list, total, err := h.repo.List(r.Context(), p.UserID, p.Role, page.Limit, page.Offset())
	switch {
	case errors.Is(err, ErrForbidden):
		httpx.SendError(w, http.StatusForbidden, httpx.ErrorResponse{ErrorCode: httpx.ErrForbidden, Message: err.Error()})
		return
	case err != nil:
		h.log.Error("list workspaces", "error", err)
		httpx.SendInternalError(w)
		return
	}

	if list == nil {
		list = []Summary{}
	}
	httpx.SendJSON(w, http.StatusOK, pageJSON{Total: total, Page: page.Number, Items: list})
}
