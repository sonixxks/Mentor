package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX — общее для *pgxpool.Pool и pgx.Tx: репозиторий может работать и в транзакции, и без неё.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const (
	codeUniqueViolation   = "23505"
	codeInvalidTextFormat = "22P02"
)

// IsUniqueViolation — нарушение уникального индекса или ограничения с именем constraint.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation && pgErr.ConstraintName == constraint
}

// IsInvalidText — значение не удалось привести к типу (например, неизвестное значение ENUM).
func IsInvalidText(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codeInvalidTextFormat
}
