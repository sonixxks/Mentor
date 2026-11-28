package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// PasswordReset — выданный код сброса пароля. Живёт в PasswordResetStore с TTL = CodeSettings.TTL.
type PasswordReset struct {
	Email    string
	UserID   uuid.UUID
	CodeHash string // HMAC-SHA256 от кода, сам код не хранится
	Attempts int
}

// PasswordResetStore хранит коды сброса пароля (в проде — Redis). Ключ — нормализованный email.
type PasswordResetStore interface {
	// Save создаёт или перезаписывает код (счётчик попыток обнуляется) и ставит TTL.
	Save(ctx context.Context, r PasswordReset, ttl time.Duration) error
	// Check сверяет код. Неверный код тратит попытку, после maxAttempts код сгорает
	// (ErrTooManyAttempts). consume = true удаляет код после успешной проверки.
	Check(ctx context.Context, email, codeHash string, maxAttempts int, consume bool) (PasswordReset, error)
}

type PasswordResetMailer interface {
	SendPasswordResetCode(ctx context.Context, email, code string) error
}

func passwordResetLimiterKey(email string) string {
	return "password-reset:" + email
}

// ForgotPassword отправляет код сброса. Отвечает одинаково, есть такой аккаунт или нет,
// а лимит повторной отправки действует для любой почты — чтобы не раскрывать список аккаунтов.
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (result RegisterResult, err error) {
	email := normalizeEmail(req.Email)
	result = RegisterResult{
		CodeExpiresIn: s.settings.CodeSettings.TTL,
		ResendAfter:   s.settings.CodeSettings.ResendInterval,
	}

	limiterKey := passwordResetLimiterKey(email)
	acquired, retryAfter, err := s.resendLimiter.Acquire(ctx, limiterKey, s.settings.CodeSettings.ResendInterval)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("acquire resend limit: %w", err)
	}
	if !acquired {
		return RegisterResult{}, &RateLimitedError{RetryAfter: retryAfter}
	}
	defer func() {
		if err != nil {
			_ = s.resendLimiter.Release(context.WithoutCancel(ctx), limiterKey)
		}
	}()

	rec, err := s.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrUserNotFound) {
		return result, nil
	}
	if err != nil {
		return RegisterResult{}, fmt.Errorf("find user: %w", err)
	}

	code, err := s.codes.Generate()
	if err != nil {
		return RegisterResult{}, err
	}
	reset := PasswordReset{Email: email, UserID: rec.User.ID, CodeHash: hashCode(s.settings.CodeSettings.HashSecret, code)}
	if err = s.passwordResets.Save(ctx, reset, s.settings.CodeSettings.TTL); err != nil {
		return RegisterResult{}, fmt.Errorf("save password reset: %w", err)
	}
	if err = s.resetMailer.SendPasswordResetCode(ctx, email, code); err != nil {
		return RegisterResult{}, fmt.Errorf("send password reset code: %w", err)
	}
	return result, nil
}

// VerifyResetCode проверяет код до ввода нового пароля; код остаётся действительным.
func (s *Service) VerifyResetCode(ctx context.Context, req VerifyResetCodeRequest) error {
	_, err := s.passwordResets.Check(ctx, normalizeEmail(req.Email),
		hashCode(s.settings.CodeSettings.HashSecret, req.Code), s.settings.CodeSettings.MaxAttempts, false)
	return err
}

// ResetPassword задаёт новый пароль, отзывает все refresh-токены пользователя и сразу выполняет вход.
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) (TokenPair, error) {
	reset, err := s.passwordResets.Check(ctx, normalizeEmail(req.Email),
		hashCode(s.settings.CodeSettings.HashSecret, req.Code), s.settings.CodeSettings.MaxAttempts, true)
	if err != nil {
		return TokenPair{}, err
	}

	hash, err := s.hasher.GenerateHash(req.NewPassword)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate hash: %w", err)
	}
	if err := s.userRepo.UpdatePassword(ctx, reset.UserID, hash); err != nil {
		return TokenPair{}, fmt.Errorf("update password: %w", err)
	}
	if err := s.refreshToken.DeleteAllForUser(ctx, reset.UserID); err != nil {
		return TokenPair{}, fmt.Errorf("revoke refresh tokens: %w", err)
	}

	rec, err := s.userRepo.FindByID(ctx, reset.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("find user: %w", err)
	}
	return s.issueTokenPair(ctx, rec.User, false)
}
