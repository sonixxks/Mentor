// Package account — настройки аккаунта текущего пользователя (/users/me).
package account

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

type Users interface {
	FindByID(ctx context.Context, id uuid.UUID) (user.UserRecord, error)
}

type Handler struct {
	users Users
	auth  *middleware.Auth
	log   *slog.Logger
}

func NewHandler(users Users, auth *middleware.Auth, log *slog.Logger) *Handler {
	return &Handler{users: users, auth: auth, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /users/me", h.auth.Require(h.me))
}

// meJSON — UserMeResponse из контракта.
type meJSON struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Role              user.Role  `json:"role"`
	HasPassword       bool       `json:"has_password"`
	PasswordChangedAt *time.Time `json:"password_changed_at"`
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFrom(r.Context())

	rec, err := h.users.FindByID(r.Context(), p.UserID)
	if errors.Is(err, user.ErrUserNotFound) {
		httpx.SendError(w, http.StatusUnauthorized, httpx.ErrorResponse{ErrorCode: httpx.ErrUnauthorized, Message: "аккаунт не найден"})
		return
	}
	if err != nil {
		h.log.Error("get account", "error", err)
		httpx.SendInternalError(w)
		return
	}

	httpx.SendJSON(w, http.StatusOK, meJSON{
		ID:                rec.User.ID,
		Email:             rec.User.Email,
		Role:              rec.User.Role,
		HasPassword:       rec.Hash != nil,
		PasswordChangedAt: rec.User.PasswordChangedAt,
	})
}
