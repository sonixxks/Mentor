package user

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
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

func (r *Repository) Create(ctx context.Context, userRecord UserRecord) (User, error) {

	u := userRecord.User
	err := r.transactions.Executor(ctx).QueryRow(ctx, `INSERT INTO users(email,password_hash,role,consent_at)
VALUES ($1,$2,$3,$4) RETURNING id, created_at`, u.Email, userRecord.Hash, u.Role, u.ConsentAt).Scan(&u.ID, &u.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
		return User{}, ErrEmailTaken
	}
	return u, err
}
func (r *Repository) FindByEmail(ctx context.Context, email string) (UserRecord, error) {
	var rec UserRecord
	u := &rec.User
	err := r.transactions.Executor(ctx).QueryRow(ctx, `SELECT id,email,role,yandex_id,consent_at,password_changed_at,created_at,password_hash FROM users WHERE email=$1`, email).Scan(&u.ID, &u.Email, &u.Role, &u.YandexID, &u.ConsentAt, &u.PasswordChangedAt, &u.CreatedAt, &rec.Hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrUserNotFound
	}
	return rec, err
}
func (r *Repository) GetByID(userID uuid.UUID) (*User, error) {
	//executor := r.transactions.Executor(ctx)
	return nil, nil
}

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.transactions.Executor(ctx).QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	return exists, err
}
