package requests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
)

const activePairIndex = "uq_requests_active_pair"

const requestColumns = `r.id, r.student_id, r.mentor_id, r.service_package_id, r.status::text,
	r.cover_letter, r.rejection_reason, r.suggested_by, r.created_at, r.updated_at`

const (
	// Ученика блокируем, чтобы параллельные заявки не обошли лимит.
	lockStudentSQL = `
		SELECT COALESCE(p.needs_curator_help, FALSE)
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.role = 'student'
		FOR UPDATE OF u`

	// Свободные места: max_students минус пары в active / paused / stagnation.
	mentorSlotsSQL = `
		SELECT COALESCE(p.max_students, 0),
		       (SELECT count(*)
		        FROM workspaces w
		        JOIN requests wr ON wr.id = w.request_id
		        WHERE wr.mentor_id = u.id AND w.status <> 'completed')
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.role = 'mentor'
		  AND p.is_published AND p.verification_status = 'verified'`

	packageExistsSQL = `
		SELECT EXISTS (
			SELECT 1 FROM service_packages
			WHERE id = $1 AND mentor_id = $2 AND status = 'published'
		)`

	// Заявки из ручного подбора куратора в лимит ученика не входят.
	countActiveSQL = `
		SELECT count(*) FROM requests
		WHERE student_id = $1 AND suggested_by IS NULL
		  AND status IN ('pending', 'approved_by_mentor')`

	insertSQL = `
		INSERT INTO requests AS r (student_id, mentor_id, service_package_id, cover_letter, suggested_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + requestColumns

	getByIDSQL = `SELECT ` + requestColumns + ` FROM requests r WHERE r.id = $1`

	lockByIDSQL = getByIDSQL + ` FOR UPDATE`

	updateStatusSQL = `
		UPDATE requests AS r
		SET status = $3::request_status,
		    rejection_reason = COALESCE($4, r.rejection_reason),
		    updated_at = now()
		WHERE r.id = $1 AND r.status = $2::request_status
		RETURNING ` + requestColumns

	// Наставника блокируем, чтобы два параллельных accept не заняли одно место.
	lockMentorSQL = `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`

	occupiedSlotsSQL = `
		SELECT COALESCE((SELECT max_students FROM profiles WHERE user_id = $1), 0),
		       (SELECT count(*)
		        FROM workspaces w
		        JOIN requests wr ON wr.id = w.request_id
		        WHERE wr.mentor_id = $1 AND w.status <> 'completed')`

	createWorkspaceSQL = `
		WITH new_workspace AS (SELECT gen_random_uuid() AS id)
		INSERT INTO workspaces (id, request_id, jitsi_link)
		SELECT id, $1, 'https://meet.jit.si/Curator_' || replace(id::text, '-', '') FROM new_workspace
		RETURNING id`

	cancelOtherActiveSQL = `
		UPDATE requests
		SET status = 'cancelled', updated_at = now()
		WHERE student_id = $1 AND id <> $2
		  AND status IN ('pending', 'approved_by_mentor')
		RETURNING id, mentor_id`

	resetHelpSQL = `UPDATE profiles SET needs_curator_help = FALSE WHERE user_id = $1`

	viewSelectSQL = `
		SELECT ` + requestColumns + `,
		       sp.title, w.id,
		       COALESCE(s.first_name, ''), COALESCE(s.last_name, ''), s.avatar_url,
		       s.target_skills, s.main_goals, s.time_limit,
		       s.min_duration_months, s.max_duration_months, s.communication_style::text,
		       COALESCE(m.first_name, ''), COALESCE(m.last_name, ''), m.avatar_url
		FROM requests r
		LEFT JOIN profiles s ON s.user_id = r.student_id
		LEFT JOIN profiles m ON m.user_id = r.mentor_id
		LEFT JOIN service_packages sp ON sp.id = r.service_package_id
		LEFT JOIN workspaces w ON w.request_id = r.id`

	listWhereSQL = `
		WHERE ($1::uuid IS NULL OR r.student_id = $1)
		  AND ($2::uuid IS NULL OR r.mentor_id = $2)
		  AND ($3::text[] IS NULL OR r.status::text = ANY ($3))`

	getViewSQL = viewSelectSQL + ` WHERE r.id = $1`

	listSQL = viewSelectSQL + listWhereSQL + `
		ORDER BY r.created_at DESC, r.id
		LIMIT $4 OFFSET $5`

	countSQL = `SELECT count(*) FROM requests r` + listWhereSQL
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	p CreateParams,
	notify func(Request) []notifications.Notification,
) (Request, error) {
	var created Request

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var needsHelp bool
		err := tx.QueryRow(ctx, lockStudentSQL, p.StudentID).Scan(&needsHelp)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStudentNotFound
		}
		if err != nil {
			return fmt.Errorf("lock student: %w", err)
		}
		if p.SuggestedBy != nil && !needsHelp {
			return ErrNoHelpRequest
		}

		var maxStudents, occupied int
		err = tx.QueryRow(ctx, mentorSlotsSQL, p.MentorID).Scan(&maxStudents, &occupied)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMentorNotFound
		}
		if err != nil {
			return fmt.Errorf("mentor slots: %w", err)
		}
		if occupied >= maxStudents {
			return ErrMentorLimit
		}

		if p.PackageID != nil {
			var exists bool
			if err := tx.QueryRow(ctx, packageExistsSQL, *p.PackageID, p.MentorID).Scan(&exists); err != nil {
				return fmt.Errorf("check package: %w", err)
			}
			if !exists {
				return ErrPackageNotFound
			}
		}

		if p.SuggestedBy == nil {
			var active int
			if err := tx.QueryRow(ctx, countActiveSQL, p.StudentID).Scan(&active); err != nil {
				return fmt.Errorf("count active requests: %w", err)
			}
			if active >= MaxActivePerStudent {
				return ErrRequestLimit
			}
		}

		created, err = scanRequest(tx.QueryRow(ctx, insertSQL,
			p.StudentID, p.MentorID, p.PackageID, p.CoverLetter, p.SuggestedBy))
		if postgres.IsUniqueViolation(err, activePairIndex) {
			return ErrDuplicate
		}
		if err != nil {
			return fmt.Errorf("insert request: %w", err)
		}

		return notifications.Insert(ctx, tx, notify(created)...)
	})
	if err != nil {
		return Request{}, err
	}
	return created, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Request, error) {
	req, err := scanRequest(r.pool.QueryRow(ctx, getByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("get request: %w", err)
	}
	return req, nil
}

func (r *PostgresRepository) GetView(ctx context.Context, id uuid.UUID) (View, error) {
	v, err := scanView(r.pool.QueryRow(ctx, getViewSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return View{}, ErrNotFound
	}
	if err != nil {
		return View{}, fmt.Errorf("get request view: %w", err)
	}
	return v, nil
}

func (r *PostgresRepository) List(ctx context.Context, f ListFilter) ([]View, int, error) {
	var statuses []string
	if f.Statuses != nil {
		statuses = make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			statuses[i] = string(s)
		}
	}

	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.StudentID, f.MentorID, statuses).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count requests: %w", err)
	}

	rows, err := r.pool.Query(ctx, listSQL, f.StudentID, f.MentorID, statuses, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list requests: %w", err)
	}

	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (View, error) {
		return scanView(row)
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list requests: %w", err)
	}
	return list, total, nil
}

func (r *PostgresRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	from, to Status,
	reason *string,
	notify []notifications.Notification,
) (Request, error) {
	var updated Request

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var err error
		updated, err = scanRequest(tx.QueryRow(ctx, updateStatusSQL, id, string(from), string(to), reason))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStatusConflict
		}
		if err != nil {
			return fmt.Errorf("update request status: %w", err)
		}
		return notifications.Insert(ctx, tx, notify...)
	})
	if err != nil {
		return Request{}, err
	}
	return updated, nil
}

// Accept в одной транзакции с блокировками:
//   - проверяет, что заявка всё ещё approved_by_mentor;
//   - проверяет свободное место у наставника (иначе ErrMentorLimit);
//   - переводит заявку в accepted и создаёт воркспейс с постоянной ссылкой на Jitsi;
//   - отменяет остальные активные заявки ученика и уведомляет их наставников;
//   - сбрасывает у ученика needs_curator_help.
func (r *PostgresRepository) Accept(ctx context.Context, id uuid.UUID, notify []notifications.Notification) (Request, uuid.UUID, error) {
	var (
		accepted    Request
		workspaceID uuid.UUID
	)

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		req, err := scanRequest(tx.QueryRow(ctx, lockByIDSQL, id))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock request: %w", err)
		}
		if req.Status != StatusApprovedByMentor {
			return ErrStatusConflict
		}

		if _, err := tx.Exec(ctx, lockMentorSQL, req.MentorID); err != nil {
			return fmt.Errorf("lock mentor: %w", err)
		}
		var maxStudents, occupied int
		if err := tx.QueryRow(ctx, occupiedSlotsSQL, req.MentorID).Scan(&maxStudents, &occupied); err != nil {
			return fmt.Errorf("mentor slots: %w", err)
		}
		if occupied >= maxStudents {
			return ErrMentorLimit
		}

		accepted, err = scanRequest(tx.QueryRow(ctx, updateStatusSQL,
			id, string(StatusApprovedByMentor), string(StatusAccepted), nil))
		if err != nil {
			return fmt.Errorf("update request status: %w", err)
		}

		if err := tx.QueryRow(ctx, createWorkspaceSQL, id).Scan(&workspaceID); err != nil {
			return fmt.Errorf("create workspace: %w", err)
		}

		cancelled, err := cancelOtherActive(ctx, tx, req.StudentID, id)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, resetHelpSQL, req.StudentID); err != nil {
			return fmt.Errorf("reset needs_curator_help: %w", err)
		}

		return notifications.Insert(ctx, tx, append(notify, cancelled...)...)
	})
	if err != nil {
		return Request{}, uuid.Nil, err
	}
	return accepted, workspaceID, nil
}

func cancelOtherActive(ctx context.Context, tx pgx.Tx, studentID, acceptedID uuid.UUID) ([]notifications.Notification, error) {
	rows, err := tx.Query(ctx, cancelOtherActiveSQL, studentID, acceptedID)
	if err != nil {
		return nil, fmt.Errorf("cancel other requests: %w", err)
	}

	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (notifications.Notification, error) {
		var requestID, mentorID uuid.UUID
		if err := row.Scan(&requestID, &mentorID); err != nil {
			return notifications.Notification{}, err
		}
		return notice(mentorID, "Ученик выбрал другого наставника, заявка отменена", nil, requestID), nil
	})
	if err != nil {
		return nil, fmt.Errorf("cancel other requests: %w", err)
	}
	return list, nil
}

func scanRequest(row pgx.Row) (Request, error) {
	var (
		req    Request
		status string
	)
	err := row.Scan(
		&req.ID, &req.StudentID, &req.MentorID, &req.ServicePackageID, &status,
		&req.CoverLetter, &req.RejectionReason, &req.SuggestedBy, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return Request{}, err
	}
	req.Status = Status(status)
	return req, nil
}

func scanView(row pgx.Row) (View, error) {
	var (
		v                View
		status           string
		skills, goals    []byte
		studentAvatarURL *string
		studentLastName  string
		studentFirstName string
	)

	err := row.Scan(
		&v.ID, &v.StudentID, &v.MentorID, &v.ServicePackageID, &status,
		&v.CoverLetter, &v.RejectionReason, &v.SuggestedBy, &v.CreatedAt, &v.UpdatedAt,
		&v.ServicePackageTitle, &v.WorkspaceID,
		&studentFirstName, &studentLastName, &studentAvatarURL,
		&skills, &goals, &v.Student.TimeLimit,
		&v.Student.MinDurationMonths, &v.Student.MaxDurationMonths, &v.Student.CommunicationStyle,
		&v.Mentor.FirstName, &v.Mentor.LastName, &v.Mentor.AvatarURL,
	)
	if err != nil {
		return View{}, err
	}

	v.Status = Status(status)
	v.Mentor.ID = v.MentorID
	v.Student.ID = v.StudentID
	v.Student.FirstName = studentFirstName
	v.StudentFull = Person{ID: v.StudentID, FirstName: studentFirstName, LastName: studentLastName, AvatarURL: studentAvatarURL}

	if v.Student.TargetSkills, err = decodeStrings(skills); err != nil {
		return View{}, fmt.Errorf("decode target_skills: %w", err)
	}
	if v.Student.MainGoals, err = decodeStrings(goals); err != nil {
		return View{}, fmt.Errorf("decode main_goals: %w", err)
	}
	return v, nil
}

func decodeStrings(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
