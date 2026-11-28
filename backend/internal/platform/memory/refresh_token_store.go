package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/auth"
)

var _ auth.RefreshTokenStore = (*RefreshTokenStore)(nil)

type refreshEntry struct {
	session   auth.RefreshSession
	expiresAt time.Time
}

// RefreshTokenStore хранит refresh-сессии по хэшу токена.
// byUser — аналог Redis-множества user_refresh:<userID>, нужен для DeleteAllForUser.
type RefreshTokenStore struct {
	mu     sync.Mutex
	tokens map[string]refreshEntry
	byUser map[uuid.UUID]map[string]struct{}
	now    func() time.Time
}

func NewRefreshTokenStore() *RefreshTokenStore {
	return &RefreshTokenStore{
		tokens: make(map[string]refreshEntry),
		byUser: make(map[uuid.UUID]map[string]struct{}),
		now:    time.Now,
	}
}

func (s *RefreshTokenStore) Save(_ context.Context, hash string, session auth.RefreshSession, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[hash] = refreshEntry{session: session, expiresAt: s.now().Add(ttl)}

	set, ok := s.byUser[session.UserID]
	if !ok {
		set = make(map[string]struct{})
		s.byUser[session.UserID] = set
	}
	set[hash] = struct{}{}
	return nil
}

// Consume атомарно (под одним захватом мьютекса) читает и удаляет запись — аналог GETDEL.
// Второй вызов с тем же хэшем вернёт ErrRefreshTokenNotFound: токен одноразовый.
func (s *RefreshTokenStore) Consume(_ context.Context, hash string) (auth.RefreshSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.tokens[hash]
	if !ok {
		return auth.RefreshSession{}, auth.ErrRefreshTokenNotFound
	}
	s.removeLocked(hash, e.session.UserID)

	if !s.now().Before(e.expiresAt) {
		return auth.RefreshSession{}, auth.ErrRefreshTokenNotFound
	}
	return e.session, nil
}

// Delete идемпотентен: отсутствие записи не ошибка (повторный logout -> 204).
func (s *RefreshTokenStore) Delete(_ context.Context, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.tokens[hash]; ok {
		s.removeLocked(hash, e.session.UserID)
	}
	return nil
}

// DeleteAllForUser отзывает все refresh-токены пользователя (смена/сброс пароля).
func (s *RefreshTokenStore) DeleteAllForUser(_ context.Context, userID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hash := range s.byUser[userID] {
		delete(s.tokens, hash)
	}
	delete(s.byUser, userID)
	return nil
}

// removeLocked удаляет токен и его хэш из множества пользователя. Вызывать только под s.mu.
func (s *RefreshTokenStore) removeLocked(hash string, userID uuid.UUID) {
	delete(s.tokens, hash)
	if set, ok := s.byUser[userID]; ok {
		delete(set, hash)
		if len(set) == 0 {
			delete(s.byUser, userID)
		}
	}
}
