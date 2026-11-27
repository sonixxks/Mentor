package profile

import (
	"context"

	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
)

type Repository struct {
	transactions TransactionManager
}

func NewRepository(transactions TransactionManager) *Repository {
	return &Repository{transactions: transactions}
}

type TransactionManager interface {
	Executor(ctx context.Context) postgres.DBTX
}

func (r *Repository) Create(ctx context.Context, p Profile) (Profile, error) {
	err := r.transactions.Executor(ctx).QueryRow(ctx, `INSERT INTO profiles(user_id,first_name,last_name,middle_name) VALUES ($1,$2,$3,$4) RETURNING id`, p.UserID, p.FirstName, p.LastName, p.MiddleName).Scan(&p.ID)
	return p, err
}
