package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	UserID      uuid.UUID `json:"id"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
type RegRequest struct {
	Role      user.Role `json:"role" validate:"required,oneof=student mentor"`
	FirstName string    `json:"first_name" validate:"required,max=255"`
	LastName  string    `json:"last_name" validate:"required,max=255"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required,min=8,max=72"`
}

type RegResponse struct {
	UserID    uuid.UUID `json:"id"`
	Role      user.Role `json:"role"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}
