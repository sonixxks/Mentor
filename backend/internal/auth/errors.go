package auth

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidCode        = errors.New("invalid code")
	ErrTooManyAttempts    = errors.New("too many attempts")
	// ErrRegistrationNotFound возвращает RegistrationStore, если записи нет или она истекла.
	ErrRegistrationNotFound = errors.New("pending registration not found")
	ErrRefreshTokenNotFound = errors.New("resresh token not found")
)

// RateLimitedError — действие повторено слишком рано. Хендлер отдаёт 429
// и заголовок Retry-After из RetryAfter.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}
