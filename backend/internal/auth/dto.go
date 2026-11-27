package auth

import (
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me,omitempty"`
}

type TokenResponse struct {
	//UserID       uuid.UUID `json:"id"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token"`
	//ExpiresAt    time.Time `json:"expires_at"`
}

type RegRequest struct {
	Role          user.Role `json:"role" validate:"required,oneof=student mentor"`
	FirstName     string    `json:"first_name" validate:"required,max=255"`
	LastName      string    `json:"last_name" validate:"required,max=255"`
	MiddleName    string    `json:"middle_name,omitempty" validate:"max=255"`
	Email         string    `json:"email" validate:"required,email,max=255"`
	Password      string    `json:"password" validate:"required,min=8,max=72"`
	TermsAccepted bool      `json:"terms_accepted" validate:"required"`
}

type CodeSentResponse struct {
	ExpiresInSeconds   int `json:"expires_in_seconds"`
	ResendAfterSeconds int `json:"resend_after_seconds"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}
type ResendCodeRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
