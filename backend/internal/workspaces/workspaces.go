// Package workspaces — список пар текущего пользователя (GET /workspaces).
package workspaces

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

var ErrForbidden = errors.New("список пар доступен ученику и наставнику")

type Person struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url"`
}

// Summary — WorkspaceSummary из контракта. После создания пары партнёр виден с фамилией и фото.
type Summary struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	Partner         Person    `json:"partner"`
	ProgressPercent int       `json:"progress_percent"`
	StartedAt       time.Time `json:"started_at"`
	UnreadMessages  int       `json:"unread_messages"`
	TasksOnReview   *int      `json:"tasks_on_review,omitempty"`
	InactiveDays    *int      `json:"inactive_days"`

	lastActivity time.Time
}

// listSQL: $1 — пользователь, $2 — его роль ('student' или 'mentor').
// Сверху пары с самой свежей активностью — как список диалогов.
const listSQL = `
	WITH mine AS (
		SELECT w.id, w.status::text AS status, w.created_at,
		       CASE WHEN $2 = 'student' THEN r.mentor_id ELSE r.student_id END AS partner_id
		FROM workspaces w
		JOIN requests r ON r.id = w.request_id
		WHERE ($2 = 'student' AND r.student_id = $1) OR ($2 = 'mentor' AND r.mentor_id = $1)
	)
	SELECT m.id, m.status, m.created_at, m.partner_id,
	       COALESCE(pp.first_name, ''), COALESCE(pp.last_name, ''), pp.avatar_url,
	       t.total, t.completed, t.on_review,
	       (SELECT count(*) FROM messages msg
	        WHERE msg.workspace_id = m.id AND NOT msg.is_read AND msg.sender_id IS DISTINCT FROM $1)::int,
	       GREATEST(m.created_at, COALESCE(act.last_activity, m.created_at)),
	       count(*) OVER ()
	FROM mine m
	LEFT JOIN profiles pp ON pp.user_id = m.partner_id
	LEFT JOIN LATERAL (
		SELECT count(*)::int AS total,
		       count(*) FILTER (WHERE status = 'completed')::int AS completed,
		       count(*) FILTER (WHERE status = 'on_review')::int AS on_review
		FROM tasks WHERE workspace_id = m.id
	) t ON TRUE
	LEFT JOIN LATERAL (
		SELECT max(ts) AS last_activity FROM (
			SELECT max(created_at) AS ts FROM messages WHERE workspace_id = m.id AND type = 'text'
			UNION ALL SELECT max(updated_at) FROM tasks WHERE workspace_id = m.id
			UNION ALL SELECT max(tc.created_at) FROM task_comments tc
			          JOIN tasks tt ON tt.id = tc.task_id WHERE tt.workspace_id = m.id
			UNION ALL SELECT max(start_time) FROM meetings WHERE workspace_id = m.id AND start_time <= now()
			UNION ALL SELECT max(updated_at) FROM meeting_notes WHERE workspace_id = m.id
			UNION ALL SELECT max(created_at) FROM knowledge_base WHERE workspace_id = m.id
		) a
	) act ON TRUE
	ORDER BY 12 DESC, m.id
	LIMIT $3 OFFSET $4`

type Repository struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, now: time.Now}
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, role user.Role, limit, offset int) ([]Summary, int, error) {
	if role != user.Student && role != user.Mentor {
		return nil, 0, ErrForbidden
	}

	rows, err := r.pool.Query(ctx, listSQL, userID, string(role), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list workspaces: %w", err)
	}

	total := 0
	now := r.now()
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Summary, error) {
		var (
			s                          Summary
			tasksTotal, done, onReview int
		)
		err := row.Scan(&s.ID, &s.Status, &s.StartedAt, &s.Partner.ID,
			&s.Partner.FirstName, &s.Partner.LastName, &s.Partner.AvatarURL,
			&tasksTotal, &done, &onReview, &s.UnreadMessages, &s.lastActivity, &total)
		if err != nil {
			return Summary{}, err
		}
		if tasksTotal > 0 {
			s.ProgressPercent = done * 100 / tasksTotal
		}
		if role == user.Mentor {
			s.TasksOnReview = &onReview
		}
		if days := int(now.Sub(s.lastActivity) / (24 * time.Hour)); days > 0 {
			s.InactiveDays = &days
		}
		return s, nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list workspaces: %w", err)
	}

	// count(*) OVER () нет, если страница пустая — тогда считаем отдельно.
	if len(list) == 0 && offset > 0 {
		if total, err = r.count(ctx, userID, role); err != nil {
			return nil, 0, err
		}
	}
	return list, total, nil
}

const countSQL = `
	SELECT count(*) FROM workspaces w JOIN requests r ON r.id = w.request_id
	WHERE ($2 = 'student' AND r.student_id = $1) OR ($2 = 'mentor' AND r.mentor_id = $1)`

func (r *Repository) count(ctx context.Context, userID uuid.UUID, role user.Role) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, countSQL, userID, string(role)).Scan(&n); err != nil {
		return 0, fmt.Errorf("count workspaces: %w", err)
	}
	return n, nil
}
