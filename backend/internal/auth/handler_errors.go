package auth

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// writeError переводит ошибку сервиса в HTTP-ответ в формате ErrorResponse из api.yaml.
// Неизвестные ошибки логируются и уходят клиенту как 500 без подробностей.
// op — имя операции для лога ("register", "login", ...).
func (h *AuthHandler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) {
	status, resp := mapError(err)

	var rateLimited *RateLimitedError
	if errors.As(err, &rateLimited) {
		// Retry-After в целых секундах, с округлением вверх, чтобы клиент не пришёл на секунду раньше.
		seconds := int(math.Ceil(rateLimited.RetryAfter.Seconds()))
		w.Header().Set("Retry-After", strconv.Itoa(seconds))
	}

	if status == http.StatusInternalServerError {
		h.log.ErrorContext(r.Context(), "auth request failed", "op", op, "error", err)
	}
	httpx.SendError(w, status, resp)
}

// mapError: доменная ошибка -> HTTP-статус и тело ответа.
// switch без выражения + errors.Is, чтобы срабатывало и на ошибках, обёрнутых через %w.
// Новые ошибки сервиса (неверный код, исчерпаны попытки, слишком частая отправка) добавляй сюда же.
func mapError(err error) (int, httpx.ErrorResponse) {
	var rateLimited *RateLimitedError

	switch {
	case errors.As(err, &rateLimited):
		return http.StatusTooManyRequests, httpx.ErrorResponse{
			ErrorCode: httpx.ErrTooManyRequests,
			Message:   "Too many requests, try again later",
		}

	case errors.Is(err, user.ErrEmailTaken):
		return http.StatusConflict, httpx.ErrorResponse{
			ErrorCode: httpx.ErrConflict,
			Message:   "Email is already registered",
		}

	// Сюда попадает только то, что пропустил валидатор:
	// длина пароля в байтах (validator max считает символы) и роль, проверенная в сервисе.
	case errors.Is(err, ErrPasswordTooLong):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrValidation,
			Message:   "Validation failed",
			Details:   []string{fmt.Sprintf("password: must be at most %d bytes long", BcryptLimit)},
		}
	case errors.Is(err, ErrInvalidRole):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrValidation,
			Message:   "Validation failed",
			Details:   []string{"role: must be one of: student mentor"},
		}

	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "Invalid email or password",
		}

	// Невалидный или истёкший access-токен (JWT) — это 401 UNAUTHORIZED.
	// INVALID_TOKEN в api.yaml — это 400 для ссылок из писем (приглашение, сброс), не для JWT.
	case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "Invalid or expired token",
		}

	case errors.Is(err, ErrInvalidCode):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrInvalidCode,
			Message:   "Invalid code",
		}

	case errors.Is(err, ErrTooManyAttempts):
		return http.StatusTooManyRequests, httpx.ErrorResponse{
			ErrorCode: httpx.ErrTooManyRequests,
			Message:   "request a new code",
		}

	case errors.Is(err, ErrRefreshTokenNotFound):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "refresh token not found",
		}

	default:
		return http.StatusInternalServerError, httpx.ErrorResponse{
			ErrorCode: httpx.ErrInternal,
			Message:   "Internal server error",
		}
	}
}
