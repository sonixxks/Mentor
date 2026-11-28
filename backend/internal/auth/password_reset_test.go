package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/auth"
)

func TestPasswordReset(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	session := register(t, e, "anna@example.com")
	sentBefore := e.mail.sent

	// Неизвестная почта: ответ тот же, письма нет, но частоту всё равно ограничиваем.
	if _, err := e.svc.ForgotPassword(ctx, auth.ForgotPasswordRequest{Email: "ghost@example.com"}); err != nil {
		t.Fatal(err)
	}
	var limited *auth.RateLimitedError
	if _, err := e.svc.ForgotPassword(ctx, auth.ForgotPasswordRequest{Email: "ghost@example.com"}); !errors.As(err, &limited) {
		t.Fatalf("ghost resend: %v", err)
	}
	if e.mail.sent != sentBefore {
		t.Fatal("no mail for unknown email")
	}

	if _, err := e.svc.ForgotPassword(ctx, auth.ForgotPasswordRequest{Email: "Anna@example.com"}); err != nil {
		t.Fatal(err)
	}
	code := e.mail.last("anna@example.com")

	check := func(c string) error {
		return e.svc.VerifyResetCode(ctx, auth.VerifyResetCodeRequest{Email: "anna@example.com", Code: c})
	}
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	if err := check(wrong); !errors.Is(err, auth.ErrInvalidCode) {
		t.Fatalf("wrong code: %v", err)
	}
	if err := check(code); err != nil {
		t.Fatalf("verify code: %v", err)
	}

	// После проверки код всё ещё действует и подходит для смены пароля.
	pair, err := e.svc.ResetPassword(ctx, auth.ResetPasswordRequest{Email: "anna@example.com", Code: code, NewPassword: "new-secret-1"})
	if err != nil || pair.RefreshToken == "" {
		t.Fatalf("reset: %+v, %v", pair, err)
	}
	if _, err := e.svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: session.RefreshToken}); !errors.Is(err, auth.ErrRefreshTokenNotFound) {
		t.Fatalf("old sessions must be revoked: %v", err)
	}
	if _, err := e.svc.Login(ctx, auth.LoginRequest{Email: "anna@example.com", Password: "secret123"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("old password: %v", err)
	}
	if _, err := e.svc.Login(ctx, auth.LoginRequest{Email: "anna@example.com", Password: "new-secret-1"}); err != nil {
		t.Fatalf("new password: %v", err)
	}
	if err := check(code); !errors.Is(err, auth.ErrInvalidCode) {
		t.Fatalf("code is one-time: %v", err)
	}

	rec, _ := e.users.FindByEmail(ctx, "anna@example.com")
	if rec.User.PasswordChangedAt == nil {
		t.Fatal("password_changed_at must be set")
	}
}

func TestPasswordResetAttempts(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	register(t, e, "anna@example.com")
	if _, err := e.svc.ForgotPassword(ctx, auth.ForgotPasswordRequest{Email: "anna@example.com"}); err != nil {
		t.Fatal(err)
	}
	code := e.mail.last("anna@example.com")
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}

	req := auth.VerifyResetCodeRequest{Email: "anna@example.com", Code: wrong}
	for i := 0; i < 2; i++ {
		if err := e.svc.VerifyResetCode(ctx, req); !errors.Is(err, auth.ErrInvalidCode) {
			t.Fatalf("attempt %d: %v", i+1, err)
		}
	}
	if err := e.svc.VerifyResetCode(ctx, req); !errors.Is(err, auth.ErrTooManyAttempts) {
		t.Fatalf("third attempt: %v", err)
	}
	// Код сгорел — даже правильный не подходит.
	if err := e.svc.VerifyResetCode(ctx, auth.VerifyResetCodeRequest{Email: "anna@example.com", Code: code}); !errors.Is(err, auth.ErrInvalidCode) {
		t.Fatalf("burned code: %v", err)
	}

	// Через интервал можно запросить новый код.
	time.Sleep(60 * time.Millisecond)
	if _, err := e.svc.ForgotPassword(ctx, auth.ForgotPasswordRequest{Email: "anna@example.com"}); err != nil {
		t.Fatal(err)
	}
}
