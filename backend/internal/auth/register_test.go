package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type fakeUsers struct{ exists bool }

func (f *fakeUsers) Create(context.Context, user.UserRecord) (user.User, error) {
	return user.User{}, nil
}
func (f *fakeUsers) FindByEmail(context.Context, string) (user.UserRecord, error) {
	return user.UserRecord{}, user.ErrUserNotFound
}
func (f *fakeUsers) ExistsByEmail(context.Context, string) (bool, error) { return f.exists, nil }
func (f *fakeUsers) FindByID(context.Context, uuid.UUID) (user.UserRecord, error) {
	return user.UserRecord{}, user.ErrUserNotFound
}
func (f *fakeUsers) UpdatePassword(context.Context, uuid.UUID, string) error { return nil }

type fakeProfiles struct{}

func (fakeProfiles) Create(_ context.Context, p profile.Profile) (profile.Profile, error) {
	return p, nil
}

type fakeHasher struct{}

func (fakeHasher) GenerateHash(p string) (string, error) { return "h:" + p, nil }
func (fakeHasher) VerifyHash(string, string) error       { return nil }
func (fakeHasher) VerifyHashSimulate(string)             {}

type fakeStore struct {
	RegistrationStore
	m       map[string]PendingRegistration
	deleted int
}

func (f *fakeStore) Save(_ context.Context, p PendingRegistration, _ time.Duration) error {
	f.m[p.Email] = p
	return nil
}
func (f *fakeStore) Get(_ context.Context, e string) (PendingRegistration, error) {
	p, ok := f.m[e]
	if !ok {
		return p, ErrRegistrationNotFound
	}
	return p, nil
}
func (f *fakeStore) IncrementAttempts(context.Context, string) (int, error) { return 0, nil }
func (f *fakeStore) Delete(_ context.Context, e string) error {
	delete(f.m, e)
	f.deleted++
	return nil
}

type fakeLimiter struct {
	held     map[string]bool
	released int
}

func (f *fakeLimiter) Acquire(_ context.Context, k string, _ time.Duration) (bool, time.Duration, error) {
	if f.held[k] {
		return false, 42500 * time.Millisecond, nil
	}
	f.held[k] = true
	return true, 0, nil
}
func (f *fakeLimiter) Release(_ context.Context, k string) error {
	delete(f.held, k)
	f.released++
	return nil
}

type fakeMailer struct {
	err  error
	code string
	to   string
}

func (f *fakeMailer) SendVerificationCode(_ context.Context, e, c string) error {
	f.to, f.code = e, c
	return f.err
}

type fixedCode struct{}

func (fixedCode) Generate() (string, error) { return "123456", nil }

func newTestService(users *fakeUsers, store *fakeStore, lim *fakeLimiter, mail *fakeMailer) *Service {
	s := NewService(ServiceDeps{
		Users: users, Profiles: fakeProfiles{}, Hasher: fakeHasher{},
		Registrations: store, ResendLimiter: lim, Mailer: mail, Codes: fixedCode{},
	}, Settings{
		CodeSettings: CodeSettings{TTL: 10 * time.Minute, ResendInterval: time.Minute, MaxAttempts: 3, HashSecret: []byte("secret")},
	})
	s.now = func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }
	return s
}

func req() RegRequest {
	return RegRequest{Role: user.Student, FirstName: " Иван ", LastName: "Петров", MiddleName: "  ",
		Email: "  Ivan@Mail.RU ", Password: "password1", TermsAccepted: true}
}

func TestRegisterHappy(t *testing.T) {
	store := &fakeStore{m: map[string]PendingRegistration{}}
	lim := &fakeLimiter{held: map[string]bool{}}
	mail := &fakeMailer{}
	s := newTestService(&fakeUsers{}, store, lim, mail)
	res, err := s.Register(context.Background(), req())
	if err != nil {
		t.Fatal(err)
	}
	if res.CodeExpiresIn != 10*time.Minute || res.ResendAfter != time.Minute {
		t.Fatalf("%+v", res)
	}
	p, ok := store.m["ivan@mail.ru"]
	if !ok {
		t.Fatal("not saved under normalized email")
	}
	if p.FirstName != "Иван" || p.MiddleName != nil || p.PasswordHash != "h:password1" || p.CodeHash == "123456" || p.CodeHash != hashCode([]byte("secret"), "123456") {
		t.Fatalf("%+v", p)
	}
	if mail.to != "ivan@mail.ru" || mail.code != "123456" {
		t.Fatal("mail")
	}
	// second immediate call -> rate limited
	_, err = s.Register(context.Background(), req())
	var rl *RateLimitedError
	if !errors.As(err, &rl) || rl.RetryAfter <= 0 {
		t.Fatalf("want rate limit, got %v", err)
	}
}

func TestRegisterEmailTaken(t *testing.T) {
	lim := &fakeLimiter{held: map[string]bool{}}
	s := newTestService(&fakeUsers{exists: true}, &fakeStore{m: map[string]PendingRegistration{}}, lim, &fakeMailer{})
	_, err := s.Register(context.Background(), req())
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Fatal(err)
	}
	if len(lim.held) != 0 {
		t.Fatal("limiter should not be taken")
	}
}

func TestRegisterMailFailReleases(t *testing.T) {
	store := &fakeStore{m: map[string]PendingRegistration{}}
	lim := &fakeLimiter{held: map[string]bool{}}
	s := newTestService(&fakeUsers{}, store, lim, &fakeMailer{err: errors.New("smtp down")})
	_, err := s.Register(context.Background(), req())
	if err == nil {
		t.Fatal("want err")
	}
	if len(store.m) != 0 || lim.released != 1 || len(lim.held) != 0 {
		t.Fatalf("cleanup failed: store=%d released=%d", len(store.m), lim.released)
	}
	if st, _ := mapError(err); st != 500 {
		t.Fatal(st)
	}
}

func TestWriteErrorRetryAfter(t *testing.T) {
	h := &AuthHandler{log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	rec := httptest.NewRecorder()
	h.writeError(rec, httptest.NewRequest("POST", "/", nil), "register", &RateLimitedError{RetryAfter: 42500 * time.Millisecond})
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") != "43" {
		t.Fatalf("%d %q %s", rec.Code, rec.Header().Get("Retry-After"), rec.Body.String())
	}
}
