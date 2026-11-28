package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/platform/memory"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// memUsers — пользователи в памяти.
type memUsers struct {
	mu   sync.Mutex
	recs map[string]user.UserRecord
}

func (m *memUsers) Create(_ context.Context, rec user.UserRecord) (user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.recs[rec.User.Email]; ok {
		return user.User{}, user.ErrEmailTaken
	}
	rec.User.ID = uuid.New()
	m.recs[rec.User.Email] = rec
	return rec.User, nil
}

func (m *memUsers) FindByEmail(_ context.Context, email string) (user.UserRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.recs[email]; ok {
		return rec, nil
	}
	return user.UserRecord{}, user.ErrUserNotFound
}

func (m *memUsers) FindByID(_ context.Context, id uuid.UUID) (user.UserRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.recs {
		if rec.User.ID == id {
			return rec, nil
		}
	}
	return user.UserRecord{}, user.ErrUserNotFound
}

func (m *memUsers) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, err := m.FindByEmail(ctx, email)
	return err == nil, nil
}

func (m *memUsers) UpdatePassword(_ context.Context, id uuid.UUID, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for email, rec := range m.recs {
		if rec.User.ID == id {
			rec.Hash = &hash
			now := time.Now()
			rec.User.PasswordChangedAt = &now
			m.recs[email] = rec
			return nil
		}
	}
	return user.ErrUserNotFound
}

type noProfiles struct{}

func (noProfiles) Create(_ context.Context, p profile.Profile) (profile.Profile, error) {
	return p, nil
}

// noTx выполняет функцию без транзакции.
type noTx struct{}

func (noTx) ExecuteInTransaction(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

// mailbox запоминает последний код для каждой почты.
type mailbox struct {
	mu    sync.Mutex
	codes map[string]string
	sent  int
}

func (b *mailbox) SendVerificationCode(_ context.Context, email, code string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.codes[email] = code
	b.sent++
	return nil
}

func (b *mailbox) last(email string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.codes[email]
}

type env struct {
	svc   *auth.Service
	users *memUsers
	mail  *mailbox
}

func newEnv(t *testing.T) env {
	t.Helper()
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute, "test")
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := auth.NewPasswordHasher(4)
	if err != nil {
		t.Fatal(err)
	}
	users := &memUsers{recs: map[string]user.UserRecord{}}
	box := &mailbox{codes: map[string]string{}}

	svc := auth.NewService(auth.ServiceDeps{
		Users:         users,
		Profiles:      noProfiles{},
		Tokens:        tokens,
		RefreshTokens: memory.NewRefreshTokenStore(),
		Hasher:        hasher,
		Transactions:  noTx{},
		Registrations: memory.NewRegistrationStore(),
		ResendLimiter: memory.NewResendLimiter(),
		Mailer:        box,
		Codes:         auth.RandomCodeGenerator{},
	}, auth.Settings{
		CodeSettings: auth.CodeSettings{
			// Короткий интервал, чтобы проверить повторную отправку без долгого ожидания.
			TTL: 10 * time.Minute, ResendInterval: 50 * time.Millisecond, MaxAttempts: 3,
			HashSecret: []byte("0123456789abcdef0123456789abcdef"),
		},
		RefreshTokenSettings: auth.NewRefreshTokenSettings(24*time.Hour, 720*time.Hour),
	})
	return env{svc: svc, users: users, mail: box}
}

func register(t *testing.T, e env, email string) auth.TokenPair {
	t.Helper()
	ctx := context.Background()
	_, err := e.svc.Register(ctx, auth.RegRequest{
		Role: user.Student, FirstName: "Анна", LastName: "Петрова",
		Email: email, Password: "secret123", TermsAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	pair, err := e.svc.VerifyEmail(ctx, auth.VerifyEmailRequest{Email: email, Code: e.mail.last(email)})
	if err != nil {
		t.Fatal(err)
	}
	return pair
}

func TestResendCode(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	// Регистрации нет — ответ тот же, письма нет.
	if _, err := e.svc.ResendCode(ctx, auth.ResendCodeRequest{Email: "ghost@example.com"}); err != nil || e.mail.sent != 0 {
		t.Fatalf("ghost: err = %v, sent = %d", err, e.mail.sent)
	}

	_, err := e.svc.Register(ctx, auth.RegRequest{
		Role: user.Mentor, FirstName: "Иван", LastName: "Смирнов",
		Email: "ivan@example.com", Password: "secret123", TermsAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	first := e.mail.last("ivan@example.com")

	// Сразу после регистрации повторить нельзя.
	var limited *auth.RateLimitedError
	if _, err := e.svc.ResendCode(ctx, auth.ResendCodeRequest{Email: "ivan@example.com"}); !errors.As(err, &limited) {
		t.Fatalf("early resend: %v", err)
	}

	time.Sleep(60 * time.Millisecond)
	res, err := e.svc.ResendCode(ctx, auth.ResendCodeRequest{Email: "Ivan@Example.com "})
	if err != nil || res.CodeExpiresIn != 10*time.Minute || e.mail.sent != 2 {
		t.Fatalf("resend: %+v, err = %v, sent = %d", res, err, e.mail.sent)
	}

	// Старый код больше не подходит, новый — да.
	second := e.mail.last("ivan@example.com")
	if first != second {
		if _, err := e.svc.VerifyEmail(ctx, auth.VerifyEmailRequest{Email: "ivan@example.com", Code: first}); !errors.Is(err, auth.ErrInvalidCode) {
			t.Fatalf("old code: %v", err)
		}
	}
	if _, err := e.svc.VerifyEmail(ctx, auth.VerifyEmailRequest{Email: "ivan@example.com", Code: second}); err != nil {
		t.Fatalf("new code: %v", err)
	}
}

func TestRefresh(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	pair := register(t, e, "anna@example.com")

	next, err := e.svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: pair.RefreshToken})
	if err != nil || next.AccessToken == "" || next.RefreshToken == pair.RefreshToken {
		t.Fatalf("refresh: %+v, %v", next, err)
	}
	if _, err := e.svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: pair.RefreshToken}); !errors.Is(err, auth.ErrRefreshTokenNotFound) {
		t.Fatalf("refresh token is one-time: %v", err)
	}
	if _, err := e.svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: "garbage"}); !errors.Is(err, auth.ErrRefreshTokenNotFound) {
		t.Fatalf("unknown token: %v", err)
	}
}
