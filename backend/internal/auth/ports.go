package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// Зависимости сервиса auth. Интерфейсы объявлены здесь, у потребителя;
// реализации живут в своих пакетах (postgres-репозитории, redis, mailer) и собираются в app.go.

type TokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role user.Role) (string, time.Time, error)
}

type UserRepository interface {
	Create(ctx context.Context, userRecord user.UserRecord) (user.User, error)
	FindByEmail(ctx context.Context, email string) (user.UserRecord, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, p profile.Profile) (profile.Profile, error)
}

type Hasher interface {
	GenerateHash(password string) (string, error)
	VerifyHash(password string, hash string) error
	VerifyHashSimulate(password string)
}

type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, f func(context.Context) error) error
}

// RegistrationStore хранит незавершённые регистрации до подтверждения почты (в проде — Redis).
// Ключ — нормализованный email.
type RegistrationStore interface {
	ConsumeCode(ctx context.Context, email, codeHash string, maxAttempts int) (PendingRegistration, error)
	// Save создаёт или перезаписывает запись (счётчик попыток обнуляется) и ставит TTL.
	Save(ctx context.Context, p PendingRegistration, ttl time.Duration) error
	// Get возвращает ErrRegistrationNotFound, если записи нет или TTL истёк.
	Get(ctx context.Context, email string) (PendingRegistration, error)
	// IncrementAttempts атомарно увеличивает счётчик неверных попыток и возвращает новое значение.
	IncrementAttempts(ctx context.Context, email string) (int, error)
	Delete(ctx context.Context, email string) error
}

// ResendLimiter ограничивает частоту отправки писем с кодом.
// В Redis: SET key 1 NX EX <interval>, при неудаче — PTTL key.
type ResendLimiter interface {
	// Acquire занимает ключ на interval. Если ключ уже занят, возвращает false
	// и сколько осталось ждать.
	Acquire(ctx context.Context, key string, interval time.Duration) (acquired bool, retryAfter time.Duration, err error)
	// Release освобождает ключ досрочно (например, если письмо не ушло).
	Release(ctx context.Context, key string) error
}

type Mailer interface {
	SendVerificationCode(ctx context.Context, email, code string) error
}

// CodeGenerator выдаёт одноразовый 6-значный код.
// Отдельный интерфейс нужен, чтобы в тестах код был предсказуемым.
type CodeGenerator interface {
	Generate() (string, error)
}

type RefreshTokenStore interface {
	Save(ctx context.Context, hash string, refreshSession RefreshSession, ttl time.Duration) error
	Consume(ctx context.Context, hash string) (RefreshSession, error) // Consume атомарный и возвращает ErrRefreshTokenNotFound
	Delete(ctx context.Context, hash string) error                    // Delete не падает, если записи нет
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}
