package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Интеграционные тесты: нужен запущенный PostgreSQL.
// Пример запуска:
//
//	TEST_DB_DSN="postgres://user:pass@127.0.0.1:5432/mentor?sslmode=disable" go test ./internal/platform/postgres/ -v
//
// Без TEST_DB_DSN тесты пропускаются, поэтому `go test ./...` не падает без базы.

var errBoom = errors.New("boom")

// setupTestDB создаёт пул и отдельную таблицу для одного теста.
// Обычная, а не TEMP таблица: TEMP видна только в одном соединении,
// а пул раздаёт разные соединения.
func setupTestDB(t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN не задан, интеграционный тест пропущен")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping db: %v", err)
	}

	table := pgx.Identifier{fmt.Sprintf("tx_manager_test_%d", time.Now().UnixNano())}.Sanitize()

	if _, err := pool.Exec(ctx, "CREATE TABLE "+table+" (val TEXT NOT NULL)"); err != nil {
		pool.Close()
		t.Fatalf("create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DROP TABLE IF EXISTS "+table)
		pool.Close()
	})

	return pool, table
}

// insert пишет строку через исполнителя из ctx — так же, как будут делать репозитории.
func insert(ctx context.Context, m *TransactionManager, table, val string) error {
	_, err := m.Executor(ctx).Exec(ctx, "INSERT INTO "+table+" (val) VALUES ($1)", val)
	return err
}

// countRows считает строки через пул напрямую, то есть видит только закоммиченное.
func countRows(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return n
}

func TestExecuteInTransaction_CommitsOnSuccess(t *testing.T) {
	pool, table := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
		if err := insert(ctx, m, table, "a"); err != nil {
			return err
		}
		return insert(ctx, m, table, "b")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := countRows(t, pool, table); got != 2 {
		t.Fatalf("rows after commit = %d, want 2", got)
	}
}

func TestExecuteInTransaction_RollsBackOnError(t *testing.T) {
	pool, table := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
		if err := insert(ctx, m, table, "a"); err != nil {
			return err
		}
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want %v", err, errBoom)
	}

	if got := countRows(t, pool, table); got != 0 {
		t.Fatalf("rows after rollback = %d, want 0", got)
	}
}

func TestExecuteInTransaction_RollsBackOnPanic(t *testing.T) {
	pool, table := setupTestDB(t)
	m := NewTransactionManager(pool)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic to propagate")
			}
		}()

		_ = m.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
			if err := insert(ctx, m, table, "a"); err != nil {
				return err
			}
			panic("boom")
		})
	}()

	if got := countRows(t, pool, table); got != 0 {
		t.Fatalf("rows after panic = %d, want 0", got)
	}
}

func TestExecuteInTransaction_NestedErrorRollsBackEverything(t *testing.T) {
	pool, table := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(outerCtx context.Context) error {
		if err := insert(outerCtx, m, table, "outer"); err != nil {
			return err
		}

		return m.ExecuteInTransaction(outerCtx, func(innerCtx context.Context) error {
			if err := insert(innerCtx, m, table, "inner"); err != nil {
				return err
			}
			return errBoom
		})
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want %v (ошибка из вложенного вызова должна дойти наверх)", err, errBoom)
	}

	if got := countRows(t, pool, table); got != 0 {
		t.Fatalf("rows after nested rollback = %d, want 0", got)
	}
}

func TestExecuteInTransaction_NestedUsesSameTransaction(t *testing.T) {
	pool, _ := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(outerCtx context.Context) error {
		outer := m.Executor(outerCtx)

		return m.ExecuteInTransaction(outerCtx, func(innerCtx context.Context) error {
			if inner := m.Executor(innerCtx); inner != outer {
				t.Error("вложенный вызов открыл новую транзакцию вместо переиспользования внешней")
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteInTransaction_UncommittedNotVisibleOutside(t *testing.T) {
	pool, table := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
		if err := insert(ctx, m, table, "a"); err != nil {
			return err
		}
		// Через пул (другое соединение) незакоммиченную строку не видно.
		// Именно поэтому репозитории внутри транзакции обязаны брать ctx из параметра функции.
		if got := countRows(t, pool, table); got != 0 {
			t.Errorf("rows visible outside tx before commit = %d, want 0", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := countRows(t, pool, table); got != 1 {
		t.Fatalf("rows after commit = %d, want 1", got)
	}
}

func TestGetExecutor_WithoutTransactionReturnsPool(t *testing.T) {
	pool, _ := setupTestDB(t)
	m := NewTransactionManager(pool)

	if got := m.Executor(context.Background()); got != DBTX(pool) {
		t.Fatalf("GetExecutor without tx returned %T, want *pgxpool.Pool", got)
	}
}

func TestGetExecutor_InsideTransactionReturnsTx(t *testing.T) {
	pool, _ := setupTestDB(t)
	m := NewTransactionManager(pool)

	err := m.ExecuteInTransaction(context.Background(), func(ctx context.Context) error {
		if _, ok := m.Executor(ctx).(pgx.Tx); !ok {
			t.Errorf("GetExecutor inside tx returned %T, want pgx.Tx", m.Executor(ctx))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
