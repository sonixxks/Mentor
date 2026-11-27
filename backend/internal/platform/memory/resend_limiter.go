package memory

import (
	"context"
	"sync"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/auth"
)

var _ auth.ResendLimiter = (*ResendLimiter)(nil)

// ResendLimiter — аналог Redis SET key NX EX <interval> + PTTL key.
type ResendLimiter struct {
	mu   sync.Mutex
	held map[string]time.Time // ключ -> когда освободится
	now  func() time.Time
}

func NewResendLimiter() *ResendLimiter {
	return &ResendLimiter{
		held: make(map[string]time.Time),
		now:  time.Now,
	}
}

// Acquire занимает ключ на interval. Если ключ занят, возвращает false и сколько осталось ждать.
func (l *ResendLimiter) Acquire(_ context.Context, key string, interval time.Duration) (bool, time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if until, ok := l.held[key]; ok && now.Before(until) {
		return false, until.Sub(now), nil
	}
	l.held[key] = now.Add(interval)
	return true, 0, nil
}

// Release освобождает ключ досрочно. Идемпотентен.
func (l *ResendLimiter) Release(_ context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.held, key)
	return nil
}
