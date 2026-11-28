package memory

import (
	"context"
	"crypto/hmac"
	"sync"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/auth"
)

var _ auth.PasswordResetStore = (*PasswordResetStore)(nil)

type resetEntry struct {
	reset     auth.PasswordReset
	expiresAt time.Time
}

// PasswordResetStore хранит коды сброса пароля по нормализованному email.
type PasswordResetStore struct {
	mu      sync.Mutex
	entries map[string]resetEntry
	now     func() time.Time
}

func NewPasswordResetStore() *PasswordResetStore {
	return &PasswordResetStore{entries: make(map[string]resetEntry), now: time.Now}
}

func (s *PasswordResetStore) Save(_ context.Context, r auth.PasswordReset, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r.Attempts = 0
	s.entries[r.Email] = resetEntry{reset: r, expiresAt: s.now().Add(ttl)}
	return nil
}

func (s *PasswordResetStore) Check(_ context.Context, email, codeHash string, maxAttempts int, consume bool) (auth.PasswordReset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[email]
	if !ok || !s.now().Before(e.expiresAt) {
		delete(s.entries, email)
		return auth.PasswordReset{}, auth.ErrInvalidCode
	}

	if !hmac.Equal([]byte(e.reset.CodeHash), []byte(codeHash)) {
		e.reset.Attempts++
		if e.reset.Attempts >= maxAttempts {
			delete(s.entries, email)
			return auth.PasswordReset{}, auth.ErrTooManyAttempts
		}
		s.entries[email] = e
		return auth.PasswordReset{}, auth.ErrInvalidCode
	}

	if consume {
		delete(s.entries, email)
	}
	return e.reset, nil
}
