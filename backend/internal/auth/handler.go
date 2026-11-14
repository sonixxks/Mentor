package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func NewHandler(validate *validator.Validate, log *slog.Logger, authService AuthService) *AuthHandler {
	return &AuthHandler{
		validate:    validate,
		log:         log,
		authService: authService,
	}
}

type AuthHandler struct {
	validate    *validator.Validate
	log         *slog.Logger
	authService AuthService
}

type AuthService interface {
	Register(ctx context.Context, req RegRequest) (user.User, error)
	Login(ctx context.Context, req LoginRequest) (LoginResult, error)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	req := RegRequest{}

	if !httpx.ReadBodyWithLimit(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	err := h.validate.Struct(req)
	if err != nil {
		httpx.SendValidationErrors(w, err, http.StatusBadRequest)
		return
	}

	created, err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			httpx.SendError(w, http.StatusConflict, httpx.ErrorResponse{Error: user.ErrEmailTaken.Error()})
		} else if errors.Is(err, ErrPasswordTooLong) {
			httpx.SendError(w, http.StatusBadRequest, httpx.ErrorResponse{Error: ErrPasswordTooLong.Error()})
		} else if errors.Is(err, ErrInvalidRole) {
			httpx.SendError(w, http.StatusBadRequest, httpx.ErrorResponse{Error: ErrInvalidRole.Error()})
		} else {
			h.log.Error("Failed to register",
				"error", err)
			httpx.SendError(w, http.StatusInternalServerError, httpx.ErrorResponse{Error: "Internal server error"})
		}
		return
	}

	resp := RegResponse{
		UserID:    created.UserID,
		Role:      created.Role,
		FirstName: created.FirstName,
		LastName:  created.LastName,
		Email:     created.Email,
	}

	httpx.SendJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {

	req := LoginRequest{}
	if !httpx.ReadBodyWithLimit(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	err := h.validate.Struct(req)
	if err != nil {
		httpx.SendValidationErrors(w, err, http.StatusBadRequest)
		return
	}

	loginRes, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.SendError(w, http.StatusUnauthorized, httpx.ErrorResponse{Error: ErrInvalidCredentials.Error()})
		} else {
			h.log.Error("Failed to login",
				"error", err)
			httpx.SendError(w, http.StatusInternalServerError, httpx.ErrorResponse{Error: "Internal server error"})
		}
		return
	}

	resp := LoginResponse{
		UserID:      loginRes.User.UserID,
		AccessToken: loginRes.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   loginRes.ExpiresAt,
	}

	httpx.SendJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(mux *http.ServeMux) {

	//Registration
	mux.HandleFunc("POST /auth/register", h.handleRegister)

	//Login
	mux.HandleFunc("POST /auth/login", h.handleLogin)
}
