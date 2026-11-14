package auth

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultCost = 12
	BcryptLimit = 72
)

type PasswordHasher struct {
	cost     int
	fakeHash string
}

func NewPasswordHasher(cost int) (*PasswordHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("bcrypt cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}

	fakeHash, err := bcrypt.GenerateFromPassword([]byte(rand.Text()), cost)
	if err != nil {
		return nil, fmt.Errorf("generate fake hash: %w", err)
	}

	return &PasswordHasher{cost: cost, fakeHash: string(fakeHash)}, nil
}

func (h *PasswordHasher) VerifyHashSimulate(password string) {
	_ = h.VerifyHash(password, h.fakeHash)
}

func (h *PasswordHasher) GenerateHash(password string) (string, error) {
	if len(password) > BcryptLimit {
		return "", ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *PasswordHasher) VerifyHash(password string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) ||
		errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return ErrInvalidCredentials
	}

	return err
}
