package user

import "errors"

var (
	ErrEmailTaken   = errors.New("email taken")
	ErrUserNotFound = errors.New("user not found")
)
