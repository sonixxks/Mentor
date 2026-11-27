package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
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
	VerifyEmail(context.Context, VerifyEmailRequest) (TokenPair, error)
	Register(ctx context.Context, req RegRequest) (RegisterResult, error)
	Login(ctx context.Context, req LoginRequest) (TokenPair, error)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	req := RegRequest{}

	if !httpx.ReadBodyWithLimit(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.MiddleName = strings.TrimSpace(req.MiddleName)

	err := h.validate.Struct(req)
	if err != nil {
		httpx.SendValidationErrors(w, err, http.StatusBadRequest)
		return
	}

	regResult, err := h.authService.Register(r.Context(), req)
	if err != nil {
		h.writeError(w, r, "register", err)
		return
	}

	resp := CodeSentResponse{
		ExpiresInSeconds:   int(regResult.CodeExpiresIn.Seconds()),
		ResendAfterSeconds: int(regResult.ResendAfter.Seconds()),
	}

	httpx.SendJSON(w, http.StatusOK, resp)
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
		h.writeError(w, r, "login", err)
		return
	}

	resp := TokenResponse{
		AccessToken:  loginRes.AccessToken,
		RefreshToken: loginRes.RefreshToken,
	}

	httpx.SendJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(mux *http.ServeMux) {

	mux.HandleFunc("POST /auth/verify-email", h.handleVerifyEmail)
	//Registration
	mux.HandleFunc("POST /auth/register", h.handleRegister)

	//Login
	mux.HandleFunc("POST /auth/login", h.handleLogin)
}

func (h *AuthHandler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if !httpx.ReadBodyWithLimit(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if err := h.validate.Struct(req); err != nil {
		httpx.SendValidationErrors(w, err, http.StatusBadRequest)
		return
	}
	pair, err := h.authService.VerifyEmail(r.Context(), req)
	if err != nil {
		h.writeError(w, r, "verify-email", err)
		return
	}
	httpx.SendJSON(w, http.StatusCreated, TokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
}
