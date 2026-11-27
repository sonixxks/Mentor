package authctx

import (
	"github.com/google/uuid"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Principal struct {
	UserID uuid.UUID
	Role   user.Role
}

type PrincipalCtxKey struct{}

// func PutPrincipalInContext() context.Context {

// }

// func GetPrincipalFromContext() (Principal, bool) {

// }
