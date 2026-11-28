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
			Message:   "слишком частые запросы, попробуйте позже",
		}

	case errors.Is(err, user.ErrEmailTaken):
		return http.StatusConflict, httpx.ErrorResponse{
			ErrorCode: httpx.ErrConflict,
			Message:   "пользователь с такой почтой уже зарегистрирован",
		}

	// Сюда попадает только то, что пропустил валидатор:
	// длина пароля в байтах (validator max считает символы) и роль, проверенная в сервисе.
	case errors.Is(err, ErrPasswordTooLong):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrValidation,
			Message:   httpx.MsgValidation,
			Details:   []string{fmt.Sprintf("password: не длиннее %d байт", BcryptLimit)},
		}
	case errors.Is(err, ErrInvalidRole):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrValidation,
			Message:   httpx.MsgValidation,
			Details:   []string{"role: допустимые значения: student, mentor"},
		}

	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "неверная почта или пароль",
		}

	// Невалидный или истёкший access-токен (JWT) — это 401 UNAUTHORIZED.
	// INVALID_TOKEN в api.yaml — это 400 для ссылок из писем (приглашение, сброс), не для JWT.
	case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "недействительный или истёкший токен",
		}

	case errors.Is(err, ErrInvalidCode):
		return http.StatusBadRequest, httpx.ErrorResponse{
			ErrorCode: httpx.ErrInvalidCode,
			Message:   "неверный или просроченный код",
		}

	case errors.Is(err, ErrTooManyAttempts):
		return http.StatusTooManyRequests, httpx.ErrorResponse{
			ErrorCode: httpx.ErrTooManyRequests,
			Message:   "попытки ввода кода исчерпаны, запросите новый код",
		}

	case errors.Is(err, ErrRefreshTokenNotFound):
		return http.StatusUnauthorized, httpx.ErrorResponse{
			ErrorCode: httpx.ErrUnauthorized,
			Message:   "сессия истекла, войдите заново",
		}

	default:
		return http.StatusInternalServerError, httpx.ErrorResponse{
			ErrorCode: httpx.ErrInternal,
			Message:   "внутренняя ошибка сервера, попробуйте позже",
		}
	}
}
