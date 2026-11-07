package requests

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation    = "23505"
	pgInvalidTextFormat  = "22P02" // например, невалидный UUID
	pendingPairIndexName = "uq_requests_pending_pair"
)

const requestColumns = `id, student_id, mentor_id, service_package_id, status::text,
	curator_comment, created_at, updated_at`

const (
	createSQL = `
		INSERT INTO requests (student_id, mentor_id, service_package_id)
		SELECT $1::uuid, $2::uuid, $3::uuid
		WHERE EXISTS (SELECT 1 FROM users WHERE id = $2::uuid AND role = 'mentor')
		  AND ($3::uuid IS NULL OR EXISTS (
		      SELECT 1 FROM service_packages
		      WHERE id = $3::uuid AND mentor_id = $2::uuid AND is_active
		  ))
		RETURNING ` + requestColumns

	getByIDSQL = `SELECT ` + requestColumns + ` FROM requests WHERE id = $1::uuid`

	listByStudentSQL = `SELECT ` + requestColumns + ` FROM requests
		WHERE student_id = $1::uuid ORDER BY created_at DESC`

	listByMentorSQL = `SELECT ` + requestColumns + ` FROM requests
		WHERE mentor_id = $1::uuid ORDER BY created_at DESC`

	// Условие status = 'pending' защищает от гонок: если заявку уже обработали,
	// UPDATE не затронет ни одной строки.
	updateStatusSQL = `
		UPDATE requests
		SET status = $2::request_status,
		    curator_comment = COALESCE($3, curator_comment),
		    updated_at = now()
		WHERE id = $1::uuid AND status = 'pending'
		RETURNING ` + requestColumns

	createWorkspaceSQL = `
		WITH new_workspace AS (SELECT gen_random_uuid() AS id)
		INSERT INTO workspaces (id, request_id, jitsi_link)
		SELECT id, $1::uuid, 'https://meet.jit.si/Curator_' || id::text FROM new_workspace
		RETURNING id, request_id, jitsi_link, created_at`
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, studentID, mentorID string, packageID *string) (Request, error) {
	req, err := scanRequest(r.pool.QueryRow(ctx, createSQL, studentID, mentorID, packageID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrInvalidTarget
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == pendingPairIndexName {
		return Request{}, ErrAlreadyPending
	}
	if isInvalidText(err) {
		return Request{}, ErrInvalidTarget
	}
	if err != nil {
		return Request{}, fmt.Errorf("create request: %w", err)
	}
	return req, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (Request, error) {
	req, err := scanRequest(r.pool.QueryRow(ctx, getByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) || isInvalidText(err) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("get request: %w", err)
	}
	return req, nil
}

func (r *PostgresRepository) ListByStudent(ctx context.Context, studentID string) ([]Request, error) {
	return r.list(ctx, listByStudentSQL, studentID)
}

func (r *PostgresRepository) ListByMentor(ctx context.Context, mentorID string) ([]Request, error) {
	return r.list(ctx, listByMentorSQL, mentorID)
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, to Status, curatorComment *string) (Request, error) {
	req, err := scanRequest(r.pool.QueryRow(ctx, updateStatusSQL, id, string(to), curatorComment))
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrStatusConflict
	}
	if err != nil {
		return Request{}, fmt.Errorf("update request status: %w", err)
	}
	return req, nil
}

// Accept переводит заявку в accepted и создаёт Workspace в одной транзакции.
func (r *PostgresRepository) Accept(ctx context.Context, id string) (Request, Workspace, error) {
	var (
		req Request
		ws  Workspace
	)

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var err error
		req, err = scanRequest(tx.QueryRow(ctx, updateStatusSQL, id, string(StatusAccepted), nil))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStatusConflict
		}
		if err != nil {
			return fmt.Errorf("update request status: %w", err)
		}

		err = tx.QueryRow(ctx, createWorkspaceSQL, id).Scan(&ws.ID, &ws.RequestID, &ws.JitsiLink, &ws.CreatedAt)
		if err != nil {
			return fmt.Errorf("create workspace: %w", err)
		}
		return nil
	})
	if err != nil {
		return Request{}, Workspace{}, err
	}
	return req, ws, nil
}

func (r *PostgresRepository) list(ctx context.Context, query, userID string) ([]Request, error) {
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}

	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Request, error) {
		return scanRequest(row)
	})
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	return list, nil
}

func scanRequest(row pgx.Row) (Request, error) {
	var (
		req    Request
		status string
	)

	err := row.Scan(
		&req.ID,
		&req.StudentID,
		&req.MentorID,
		&req.ServicePackageID,
		&status,
		&req.CuratorComment,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return Request{}, err
	}

	req.Status = Status(status)
	return req, nil
}

func isInvalidText(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgInvalidTextFormat
}
