// Package memory — реализации хранилищ auth в памяти процесса.
// Только для разработки и тестов: данные пропадают при перезапуске
// и не делятся между несколькими экземплярами приложения. В проде — Redis.
package memory

import (
	"context"
	"crypto/hmac"
	"sync"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/auth"
)

// Проверка на этапе компиляции, что тип реализует интерфейс auth.
var _ auth.RegistrationStore = (*RegistrationStore)(nil)

type registrationEntry struct {
	pending   auth.PendingRegistration
	expiresAt time.Time
}

// RegistrationStore хранит незавершённые регистрации по нормализованному email.
type RegistrationStore struct {
	mu      sync.Mutex
	entries map[string]registrationEntry
	now     func() time.Time
}

func NewRegistrationStore() *RegistrationStore {
	return &RegistrationStore{
		entries: make(map[string]registrationEntry),
		now:     time.Now,
	}
}

// Save создаёт или перезаписывает запись; счётчик попыток обнуляется.
func (s *RegistrationStore) Save(_ context.Context, p auth.PendingRegistration, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.Attempts = 0
	s.entries[p.Email] = registrationEntry{pending: p, expiresAt: s.now().Add(ttl)}
	return nil
}

func (s *RegistrationStore) Get(_ context.Context, email string) (auth.PendingRegistration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.aliveLocked(email)
	if !ok {
		return auth.PendingRegistration{}, auth.ErrRegistrationNotFound
	}
	return e.pending, nil
}

// IncrementAttempts увеличивает счётчик под мьютексом — аналог атомарного HINCRBY.
// Для истёкшей или отсутствующей записи возвращает ErrRegistrationNotFound
// и ничего не создаёт (в Redis это нужно будет обеспечить отдельно).
func (s *RegistrationStore) IncrementAttempts(_ context.Context, email string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.aliveLocked(email)
	if !ok {
		return 0, auth.ErrRegistrationNotFound
	}
	e.pending.Attempts++
	s.entries[email] = e
	return e.pending.Attempts, nil
}

// Delete идемпотентен: отсутствие записи не ошибка.
func (s *RegistrationStore) Delete(_ context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, email)
	return nil
}

// aliveLocked возвращает запись, если она есть и не истекла; истёкшую удаляет.
// Вызывать только под s.mu.
func (s *RegistrationStore) aliveLocked(email string) (registrationEntry, bool) {
	e, ok := s.entries[email]
	if !ok {
		return registrationEntry{}, false
	}
	if !s.now().Before(e.expiresAt) {
		delete(s.entries, email)
		return registrationEntry{}, false
	}
	return e, true
}

func (s *RegistrationStore) ConsumeCode(_ context.Context, email, codeHash string, maxAttempts int) (auth.PendingRegistration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.aliveLocked(email)
	if !ok {
		return auth.PendingRegistration{}, auth.ErrInvalidCode
	}
	if e.pending.Attempts >= maxAttempts {
		return auth.PendingRegistration{}, auth.ErrTooManyAttempts
	}
	if !hmac.Equal([]byte(e.pending.CodeHash), []byte(codeHash)) {
		e.pending.Attempts++
		s.entries[email] = e
		if e.pending.Attempts >= maxAttempts {
			return auth.PendingRegistration{}, auth.ErrTooManyAttempts
		}
		return auth.PendingRegistration{}, auth.ErrInvalidCode
	}
	delete(s.entries, email)
	return e.pending, nil
}
