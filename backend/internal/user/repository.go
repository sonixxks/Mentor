package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func (r *Repository) Create(ctx context.Context, userRecord UserRecord) (*User, error) {
	return nil, nil
}
func (r *Repository) FindByEmail(ctx context.Context, email string) (*UserRecord, error) {
	return nil, nil
}

func (r *Repository) GetByID(userID uuid.UUID) (*User, error) {
	return nil, nil
}
