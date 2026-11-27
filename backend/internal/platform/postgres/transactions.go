package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ctxKey struct{}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	if pool != nil {
		return &TransactionManager{pool: pool}
	}
	panic("pool for TransactionManager was nil")
}

func (m *TransactionManager) ExecuteInTransaction(ctx context.Context, f func(context.Context) error) error {
	_, ok := ctx.Value(ctxKey{}).(pgx.Tx)
	if ok {
		return f(ctx)
	}
	transaction, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(context.WithoutCancel(ctx))
	}()

	transactionCtx := context.WithValue(ctx, ctxKey{}, transaction)

	err = f(transactionCtx)
	if err != nil {
		return err
	}

	commitErr := transaction.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit tx: %w", commitErr)
	}
	return nil
}

func (m *TransactionManager) Executor(ctx context.Context) DBTX {

	transaction, ok := ctx.Value(ctxKey{}).(pgx.Tx)
	if ok {
		return transaction
	}

	return m.pool
}
