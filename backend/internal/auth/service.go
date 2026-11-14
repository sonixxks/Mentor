package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type TokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role user.Role) (string, time.Time, error)
}

type UserRepository interface {
	Create(ctx context.Context, userRecord user.UserRecord) (user.User, error)
	FindByEmail(ctx context.Context, email string) (user.UserRecord, error)
}

type Hasher interface {
	GenerateHash(password string) (string, error)
	VerifyHash(password string, hash string) error
	VerifyHashSimulate(password string)
}

type Service struct {
	repo   UserRepository
	token  TokenIssuer
	hasher Hasher
}

func NewService(repo UserRepository, tokens TokenIssuer, hasher Hasher) *Service {
	return &Service{repo: repo, token: tokens, hasher: hasher}
}

func (s *Service) Register(ctx context.Context, req RegRequest) (user.User, error) {

	req.Email = strings.TrimSpace(req.Email)
	req.Email = strings.ToLower(req.Email)
	if !req.Role.Registrable() {
		return user.User{}, ErrInvalidRole
	}

	hash, err := s.hasher.GenerateHash(req.Password)
	if err != nil {
		return user.User{}, fmt.Errorf("generate hash: %w", err)
	}

	userRec := user.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		Role:      req.Role,
		LastName:  req.LastName,
	}

	rec := user.UserRecord{User: userRec, Hash: hash}

	createdUser, err := s.repo.Create(ctx, rec)
	if err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResult, error) {

	req.Email = strings.TrimSpace(req.Email)
	req.Email = strings.ToLower(req.Email)

	rec, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			s.hasher.VerifyHashSimulate(req.Password)
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("find user: %w", err)
	}

	err = s.hasher.VerifyHash(req.Password, rec.Hash)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("verify password: %w", err)
	}

	accessToken, expiresAt, err := s.token.IssueAccessToken(rec.User.UserID, rec.User.Role)
	if err != nil {
		return LoginResult{}, fmt.Errorf("get issue access token: %w", err)
	}

	return LoginResult{User: rec.User, AccessToken: accessToken, ExpiresAt: expiresAt}, nil
}
