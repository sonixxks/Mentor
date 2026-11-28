package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

// ResendCode повторно отправляет код подтверждения почты (не чаще раза в ResendInterval).
// Если незавершённой регистрации нет, письмо не уходит, но ответ тот же —
// чтобы по нему нельзя было узнать, какие почты сейчас регистрируются.
func (s *Service) ResendCode(ctx context.Context, req ResendCodeRequest) (result RegisterResult, err error) {
	email := normalizeEmail(req.Email)
	result = RegisterResult{
		CodeExpiresIn: s.settings.CodeSettings.TTL,
		ResendAfter:   s.settings.CodeSettings.ResendInterval,
	}

	pending, err := s.registrations.Get(ctx, email)
	if errors.Is(err, ErrRegistrationNotFound) {
		return result, nil
	}
	if err != nil {
		return RegisterResult{}, fmt.Errorf("get pending registration: %w", err)
	}

	limiterKey := verifyEmailLimiterKey(email)
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

	code, err := s.codes.Generate()
	if err != nil {
		return RegisterResult{}, err
	}
	// Новый код: старый перестаёт действовать, счётчик попыток обнуляется.
	pending.CodeHash = hashCode(s.settings.CodeSettings.HashSecret, code)
	if err = s.registrations.Save(ctx, pending, s.settings.CodeSettings.TTL); err != nil {
		return RegisterResult{}, fmt.Errorf("save pending registration: %w", err)
	}
	if err = s.mailer.SendVerificationCode(ctx, email, code); err != nil {
		return RegisterResult{}, fmt.Errorf("send verification code: %w", err)
	}
	return result, nil
}

// Refresh обменивает одноразовый refresh-токен на новую пару; remember_me сохраняется.
func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (TokenPair, error) {
	session, err := s.refreshToken.Consume(ctx, hashRefreshToken(req.RefreshToken))
	if err != nil {
		return TokenPair{}, err
	}

	rec, err := s.userRepo.FindByID(ctx, session.UserID)
	if errors.Is(err, user.ErrUserNotFound) {
		return TokenPair{}, ErrRefreshTokenNotFound
	}
	if err != nil {
		return TokenPair{}, fmt.Errorf("find user: %w", err)
	}
	return s.issueTokenPair(ctx, rec.User, session.RememberMe)
}
