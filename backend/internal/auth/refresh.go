package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenSettings struct {
	RefreshTTL         time.Duration
	RefreshRememberTTL time.Duration
}

func NewRefreshTokenSettings(refreshTTL time.Duration, refreshRememberTTL time.Duration) RefreshTokenSettings {
	return RefreshTokenSettings{RefreshTTL: refreshTTL, RefreshRememberTTL: refreshRememberTTL}
}

func (s *RefreshTokenSettings) TTL(rememberMe bool) time.Duration {
	if rememberMe {
		return s.RefreshRememberTTL
	}
	return s.RefreshTTL
}

type RefreshSession struct {
	UserID     uuid.UUID
	RememberMe bool
}

func generateRefreshToken() string {
	return rand.Text()
}

func hashRefreshToken(tokenBytes string) string {
	b := sha256.Sum256([]byte(tokenBytes))
	return base64.RawURLEncoding.EncodeToString(b[:])
}
