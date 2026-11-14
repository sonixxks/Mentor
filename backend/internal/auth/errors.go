package auth

import "errors"

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrInvalidRole        = errors.New("invalid role")
)
