package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/profile"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

// Структура для настроек refresh_token и confirmation_code(кода подтверждения)
type Settings struct {
	CodeSettings         CodeSettings
	RefreshTokenSettings RefreshTokenSettings
}

// ServiceDeps — зависимости сервиса. Структура вместо длинного списка
// параметров: при добавлении новой зависимости не ломаются все вызовы NewService.
type ServiceDeps struct {
	Users         UserRepository
	Profiles      ProfileRepository
	Tokens        TokenIssuer
	RefreshTokens RefreshTokenStore
	Hasher        Hasher
	Transactions  TransactionManager
	Registrations RegistrationStore
	ResendLimiter ResendLimiter
	Mailer        Mailer
	Codes         CodeGenerator

	// Восстановление пароля.
	PasswordResets PasswordResetStore
	ResetMailer    PasswordResetMailer
}

type Service struct {
	userRepo      UserRepository
	profileRepo   ProfileRepository
	token         TokenIssuer
	refreshToken  RefreshTokenStore
	hasher        Hasher
	transactions  TransactionManager
	registrations RegistrationStore
	resendLimiter ResendLimiter
	mailer        Mailer
	codes         CodeGenerator

	passwordResets PasswordResetStore
	resetMailer    PasswordResetMailer

	settings Settings
	now      func() time.Time // подменяется в тестах
}

func NewService(deps ServiceDeps, settings Settings) *Service {
	return &Service{
		userRepo:      deps.Users,
		profileRepo:   deps.Profiles,
		token:         deps.Tokens,
		refreshToken:  deps.RefreshTokens,
		hasher:        deps.Hasher,
		transactions:  deps.Transactions,
		registrations: deps.Registrations,
		resendLimiter: deps.ResendLimiter,
		mailer:        deps.Mailer,
		codes:         deps.Codes,

		passwordResets: deps.PasswordResets,
		resetMailer:    deps.ResetMailer,

		settings: settings,
		now:      time.Now,
	}
}

// RegisterResult — что знает клиент после шага 1: сколько живёт код
// и когда можно запросить его повторно (CodeSentResponse).
type RegisterResult struct {
	CodeExpiresIn time.Duration
	ResendAfter   time.Duration
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) issueTokenPair(ctx context.Context, u user.User, rememberMe bool) (TokenPair, error) {
	accessToken, expiresAt, err := s.token.IssueAccessToken(u.ID, u.Role)
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken := generateRefreshToken()
	refreshTokenTTL := s.settings.RefreshTokenSettings.TTL(rememberMe)

	err = s.refreshToken.Save(
		ctx,
		hashRefreshToken(refreshToken),
		RefreshSession{UserID: u.ID, RememberMe: rememberMe},
		refreshTokenTTL,
	)
	if err != nil {
		return TokenPair{}, fmt.Errorf("save refresh token: %w", err)
	}

	return TokenPair{AccessToken: accessToken, AccessExpiresAt: expiresAt, RefreshToken: refreshToken}, nil
}

// Register — шаг 1 регистрации: проверяет данные, сохраняет их до подтверждения
// и отправляет код на почту. В PostgreSQL ничего не пишет: аккаунт создаётся в VerifyEmail.
func (s *Service) Register(ctx context.Context, req RegRequest) (result RegisterResult, err error) {
	email := normalizeEmail(req.Email)

	if !req.Role.Registrable() {
		return RegisterResult{}, ErrInvalidRole
	}

	// Быстрая проверка, чтобы не слать письмо на занятую почту.
	// Гарантии она не даёт (почту могут занять, пока человек ждёт код),
	// поэтому ErrEmailTaken ещё раз обрабатывается в VerifyEmail по уникальному индексу.
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return RegisterResult{}, user.ErrEmailTaken
	}

	// Лимит отправки занимаем до дорогих операций (bcrypt, письмо).
	limiterKey := verifyEmailLimiterKey(email)
	acquired, retryAfter, err := s.resendLimiter.Acquire(ctx, limiterKey, s.settings.CodeSettings.ResendInterval)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("acquire resend limit: %w", err)
	}
	if !acquired {
		return RegisterResult{}, &RateLimitedError{RetryAfter: retryAfter}
	}
	// Если дальше что-то упало, письмо не ушло: освобождаем лимит,
	// чтобы пользователь мог сразу повторить, а не ждать минуту.
	// WithoutCancel — чтобы освобождение прошло, даже если клиент оборвал запрос.
	defer func() {
		if err != nil {
			_ = s.resendLimiter.Release(context.WithoutCancel(ctx), limiterKey)
		}
	}()

	passwordHash, err := s.hasher.GenerateHash(req.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("generate hash: %w", err)
	}

	code, err := s.codes.Generate()
	if err != nil {
		return RegisterResult{}, err
	}

	var middleName *string
	if mn := strings.TrimSpace(req.MiddleName); mn != "" {
		middleName = &mn
	}

	pending := PendingRegistration{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		FirstName:    strings.TrimSpace(req.FirstName),
		LastName:     strings.TrimSpace(req.LastName),
		MiddleName:   middleName,
		ConsentAt:    s.now().UTC(),
		CodeHash:     hashCode(s.settings.CodeSettings.HashSecret, code),
	}

	// Повторная регистрация того же email перезаписывает запись и выдаёт новый код.
	if err = s.registrations.Save(ctx, pending, s.settings.CodeSettings.TTL); err != nil {
		return RegisterResult{}, fmt.Errorf("save pending registration: %w", err)
	}

	if err = s.mailer.SendVerificationCode(ctx, email, code); err != nil {
		// Письмо не ушло: запись без кода у пользователя бесполезна.
		_ = s.registrations.Delete(context.WithoutCancel(ctx), email)
		return RegisterResult{}, fmt.Errorf("send verification code: %w", err)
	}

	return RegisterResult{
		CodeExpiresIn: s.settings.CodeSettings.TTL,
		ResendAfter:   s.settings.CodeSettings.ResendInterval,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	email := normalizeEmail(req.Email)

	rec, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			s.hasher.VerifyHashSimulate(req.Password)
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("find user: %w", err)
	}

	if rec.Hash == nil {
		s.hasher.VerifyHashSimulate(req.Password)
		return TokenPair{}, ErrInvalidCredentials
	}

	err = s.hasher.VerifyHash(req.Password, *rec.Hash)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("verify password: %w", err)
	}

	tokenPair, err := s.issueTokenPair(ctx, rec.User, req.RememberMe)
	if err != nil {
		return TokenPair{}, err
	}

	return tokenPair, nil
}

func (s *Service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) (TokenPair, error) {
	email := normalizeEmail(req.Email)
	pending, err := s.registrations.ConsumeCode(ctx, email, hashCode(s.settings.CodeSettings.HashSecret, req.Code), s.settings.CodeSettings.MaxAttempts)
	if err != nil {
		return TokenPair{}, err
	}
	var created user.User
	err = s.transactions.ExecuteInTransaction(ctx, func(ctx context.Context) error {
		var err error
		created, err = s.userRepo.Create(ctx, user.UserRecord{User: user.User{Email: email, Role: pending.Role, ConsentAt: &pending.ConsentAt}, Hash: &pending.PasswordHash})
		if err != nil {
			return err
		}
		_, err = s.profileRepo.Create(ctx, profile.Profile{UserID: created.ID, FirstName: pending.FirstName, LastName: pending.LastName, MiddleName: pending.MiddleName})
		return err
	})
	if err != nil {
		return TokenPair{}, err
	}
	return s.issueTokenPair(ctx, created, false)
}
