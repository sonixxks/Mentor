package auth

import (
	"time"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

type LoginResult struct {
	User        user.User
	AccessToken string
	ExpiresAt   time.Time
}
