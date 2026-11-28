package authctx

import (
	"context"

	"github.com/google/uuid"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Principal struct {
	UserID uuid.UUID
	Role   user.Role
}

type principalKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
