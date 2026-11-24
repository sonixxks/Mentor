package curator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
)

const (
	// Активность пары: сообщения, задачи и комментарии, прошедшие встречи, заметки, файлы.
	workspaceStatsSQL = `
		SELECT w.id, w.status::text, r.student_id, r.mentor_id,
		       btrim(concat_ws(' ', sp.first_name, sp.last_name)),
		       btrim(concat_ws(' ', mp.first_name, mp.last_name)),
		       sp.avatar_url, w.created_at, act.last_activity,
		       t.total, t.completed, t.overdue, lr.rating
		FROM workspaces w
		JOIN requests r ON r.id = w.request_id
		LEFT JOIN profiles sp ON sp.user_id = r.student_id
		LEFT JOIN profiles mp ON mp.user_id = r.mentor_id
		LEFT JOIN LATERAL (
			SELECT count(*)::int AS total,
			       count(*) FILTER (WHERE status = 'completed')::int AS completed,
			       count(*) FILTER (WHERE status <> 'completed' AND deadline_at < now())::int AS overdue
			FROM tasks WHERE workspace_id = w.id
		) t ON TRUE
		LEFT JOIN LATERAL (
			SELECT max(ts) AS last_activity FROM (
				SELECT max(created_at) AS ts FROM messages WHERE workspace_id = w.id AND type = 'text'
				UNION ALL SELECT max(updated_at) FROM tasks WHERE workspace_id = w.id
				UNION ALL SELECT max(tc.created_at) FROM task_comments tc
				          JOIN tasks tt ON tt.id = tc.task_id WHERE tt.workspace_id = w.id
				UNION ALL SELECT max(start_time) FROM meetings WHERE workspace_id = w.id AND start_time <= now()
				UNION ALL SELECT max(updated_at) FROM meeting_notes WHERE workspace_id = w.id
				UNION ALL SELECT max(created_at) FROM knowledge_base WHERE workspace_id = w.id
			) a
		) act ON TRUE
		LEFT JOIN LATERAL (
			SELECT rating FROM session_reviews
			WHERE workspace_id = w.id ORDER BY created_at DESC, id LIMIT 1
		) lr ON TRUE`

	listStatsSQL = workspaceStatsSQL + `
		WHERE $1::text IS NULL
		   OR concat_ws(' ', sp.first_name, sp.last_name) ILIKE $1
		   OR concat_ws(' ', mp.first_name, mp.last_name) ILIKE $1`

	statsByIDSQL = workspaceStatsSQL + ` WHERE w.id = $1`

	// Плитки считаются по всем незавершённым парам, без учёта фильтров списка.
	countsSQL = `
		SELECT count(*) FILTER (WHERE status <> 'completed')::int,
		       count(*) FILTER (WHERE status = 'active')::int,
		       count(*) FILTER (WHERE status = 'paused')::int,
		       count(*) FILTER (WHERE status = 'stagnation')::int
		FROM workspaces`

	lockWorkspaceSQL = `
		SELECT w.status::text, r.student_id, r.mentor_id
		FROM workspaces w JOIN requests r ON r.id = w.request_id
		WHERE w.id = $1
		FOR UPDATE OF w`

	lockMentorSQL = `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`

	mentorSlotsSQL = `
		SELECT COALESCE((SELECT max_students FROM profiles WHERE user_id = $1), 0),
		       (SELECT count(*) FROM workspaces w JOIN requests r ON r.id = w.request_id
		        WHERE r.mentor_id = $1 AND w.status <> 'completed')::int`

	updateWorkspaceStatusSQL = `UPDATE workspaces SET status = $2::workspace_status WHERE id = $1`

	workspaceMembersSQL = `
		SELECT r.student_id, r.mentor_id
		FROM workspaces w JOIN requests r ON r.id = w.request_id
		WHERE w.id = $1`

	profileColumns = `p.id, p.user_id, u.role::text, u.email,
		p.first_name, p.last_name, p.middle_name, p.avatar_url, p.age, p.current_status::text,
		p.university, p.major, p.course, p.experience::text, p.job_title, p.company, p.bio,
		p.hourly_rate::float8, p.max_students, p.is_published, p.communication_style::text,
		p.availability_slots, p.main_goals, p.success_metric, p.time_limit, p.needs_curator_help,
		p.min_duration_months, p.max_duration_months, p.current_skills, p.target_skills,
		p.help_topics, p.target_student_levels, p.notification_preferences,
		p.onboarding_completed, p.verification_status::text, p.verification_assets,
		CASE WHEN u.role = 'mentor' THEN GREATEST(COALESCE(p.max_students, 0) - (
			SELECT count(*) FROM workspaces w JOIN requests r ON r.id = w.request_id
			WHERE r.mentor_id = p.user_id AND w.status <> 'completed'), 0)::int END`

	profileFrom = ` FROM profiles p JOIN users u ON u.id = p.user_id`

	profileWhere = `
		WHERE ($1::bool IS NULL OR p.needs_curator_help = $1)
		  AND ($2::text IS NULL OR p.verification_status::text = $2)`

	listProfilesSQL = `SELECT ` + profileColumns + profileFrom + profileWhere + `
		ORDER BY u.created_at, p.id
		LIMIT $3 OFFSET $4`

	countProfilesSQL = `SELECT count(*)` + profileFrom + profileWhere

	profileByIDSQL = `SELECT ` + profileColumns + profileFrom + ` WHERE p.id = $1`

	lockProfileSQL = `
		SELECT p.user_id, u.role::text, p.needs_curator_help, p.verification_status::text
		FROM profiles p JOIN users u ON u.id = p.user_id
		WHERE p.id = $1
		FOR UPDATE OF p`

	resolveHelpSQL = `UPDATE profiles SET needs_curator_help = FALSE WHERE id = $1`

	setVerificationSQL = `UPDATE profiles SET verification_status = $2::verification_status WHERE id = $1`
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListWorkspaceStats(ctx context.Context, query string) ([]WorkspaceStats, error) {
	var like *string
	if q := strings.TrimSpace(query); q != "" {
		s := "%" + escapeLike(q) + "%"
		like = &s
	}

	rows, err := r.pool.Query(ctx, listStatsSQL, like)
	if err != nil {
		return nil, fmt.Errorf("list workspace stats: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (WorkspaceStats, error) {
		return scanStats(row)
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace stats: %w", err)
	}
	return list, nil
}

func (r *PostgresRepository) GetWorkspaceStats(ctx context.Context, id uuid.UUID) (WorkspaceStats, error) {
	s, err := scanStats(r.pool.QueryRow(ctx, statsByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkspaceStats{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return WorkspaceStats{}, fmt.Errorf("get workspace stats: %w", err)
	}
	return s, nil
}

func (r *PostgresRepository) CountWorkspaces(ctx context.Context) (StatusCounts, error) {
	var c StatusCounts
	if err := r.pool.QueryRow(ctx, countsSQL).Scan(&c.Total, &c.Active, &c.Paused, &c.Stagnation); err != nil {
		return StatusCounts{}, fmt.Errorf("count workspaces: %w", err)
	}
	return c, nil
}

// SetWorkspaceStatus меняет статус пары. Если пара возвращается из completed,
// под блокировкой наставника проверяется свободное место.
func (r *PostgresRepository) SetWorkspaceStatus(
	ctx context.Context,
	id uuid.UUID,
	to WorkspaceStatus,
	notify func(studentID, mentorID uuid.UUID) []notifications.Notification,
) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var (
			from                string
			studentID, mentorID uuid.UUID
		)
		err := tx.QueryRow(ctx, lockWorkspaceSQL, id).Scan(&from, &studentID, &mentorID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWorkspaceNotFound
		}
		if err != nil {
			return fmt.Errorf("lock workspace: %w", err)
		}
		if WorkspaceStatus(from) == to {
			return nil
		}

		if !WorkspaceStatus(from).OccupiesSlot() && to.OccupiesSlot() {
			if _, err := tx.Exec(ctx, lockMentorSQL, mentorID); err != nil {
				return fmt.Errorf("lock mentor: %w", err)
			}
			var maxStudents, occupied int
			if err := tx.QueryRow(ctx, mentorSlotsSQL, mentorID).Scan(&maxStudents, &occupied); err != nil {
				return fmt.Errorf("mentor slots: %w", err)
			}
			if occupied >= maxStudents {
				return ErrMentorLimit
			}
		}

		if _, err := tx.Exec(ctx, updateWorkspaceStatusSQL, id, string(to)); err != nil {
			return fmt.Errorf("update workspace status: %w", err)
		}
		return notifications.Insert(ctx, tx, notify(studentID, mentorID)...)
	})
}

func (r *PostgresRepository) WorkspaceMembers(ctx context.Context, id uuid.UUID) (studentID, mentorID uuid.UUID, err error) {
	err = r.pool.QueryRow(ctx, workspaceMembersSQL, id).Scan(&studentID, &mentorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, ErrWorkspaceNotFound
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("workspace members: %w", err)
	}
	return studentID, mentorID, nil
}

func (r *PostgresRepository) Notify(ctx context.Context, list []notifications.Notification) error {
	return notifications.Insert(ctx, r.pool, list...)
}

func (r *PostgresRepository) ListProfiles(ctx context.Context, f ProfileFilter) ([]Profile, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, countProfilesSQL, f.NeedsHelp, f.VerificationStatus).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count profiles: %w", err)
	}

	rows, err := r.pool.Query(ctx, listProfilesSQL, f.NeedsHelp, f.VerificationStatus, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list profiles: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Profile, error) {
		return scanProfile(row)
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list profiles: %w", err)
	}
	return list, total, nil
}

func (r *PostgresRepository) GetProfile(ctx context.Context, id uuid.UUID) (Profile, error) {
	p, err := scanProfile(r.pool.QueryRow(ctx, profileByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrProfileNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// ProfileState — то, что нужно проверить перед действием куратора над профилем.
type ProfileState struct {
	UserID             uuid.UUID
	Role               string
	NeedsCuratorHelp   bool
	VerificationStatus string
}

// UpdateProfile блокирует профиль, даёт сервису проверить состояние (check)
// и выполняет изменение с уведомлениями в одной транзакции.
// check возвращает новый статус верификации (или nil) и сбрасывать ли запрос помощи.
func (r *PostgresRepository) UpdateProfile(
	ctx context.Context,
	id uuid.UUID,
	check func(ProfileState) (ProfileChange, error),
) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var st ProfileState
		err := tx.QueryRow(ctx, lockProfileSQL, id).Scan(&st.UserID, &st.Role, &st.NeedsCuratorHelp, &st.VerificationStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		if err != nil {
			return fmt.Errorf("lock profile: %w", err)
		}

		change, err := check(st)
		if err != nil {
			return err
		}

		if change.ResolveHelp {
			if _, err := tx.Exec(ctx, resolveHelpSQL, id); err != nil {
				return fmt.Errorf("resolve help request: %w", err)
			}
		}
		if change.Verification != nil {
			if _, err := tx.Exec(ctx, setVerificationSQL, id, *change.Verification); err != nil {
				return fmt.Errorf("set verification status: %w", err)
			}
		}
		return notifications.Insert(ctx, tx, change.Notify...)
	})
}

// ProfileChange — что сделать с профилем после проверки.
type ProfileChange struct {
	ResolveHelp  bool
	Verification *string
	Notify       []notifications.Notification
}

func scanStats(row pgx.Row) (WorkspaceStats, error) {
	var (
		s      WorkspaceStats
		status string
	)
	err := row.Scan(&s.ID, &status, &s.StudentID, &s.MentorID, &s.StudentName, &s.MentorName,
		&s.StudentAvatar, &s.StartedAt, &s.LastActivityAt,
		&s.TasksTotal, &s.TasksCompleted, &s.TasksOverdue, &s.LastRating)
	if err != nil {
		return WorkspaceStats{}, err
	}
	s.Status = WorkspaceStatus(status)
	return s, nil
}

func scanProfile(row pgx.Row) (Profile, error) {
	var p Profile
	var slots, goals, current, target, topics, levels, prefs, assets []byte

	err := row.Scan(
		&p.ID, &p.UserID, &p.Role, &p.Email,
		&p.FirstName, &p.LastName, &p.MiddleName, &p.AvatarURL, &p.Age, &p.CurrentStatus,
		&p.University, &p.Major, &p.Course, &p.Experience, &p.JobTitle, &p.Company, &p.Bio,
		&p.HourlyRate, &p.MaxStudents, &p.IsPublished, &p.CommunicationStyle,
		&slots, &goals, &p.SuccessMetric, &p.TimeLimit, &p.NeedsCuratorHelp,
		&p.MinDurationMonths, &p.MaxDurationMonths, &current, &target,
		&topics, &levels, &prefs,
		&p.OnboardingCompleted, &p.VerificationStatus, &assets,
		&p.FreeSlots,
	)
	if err != nil {
		return Profile{}, err
	}

	p.AvailabilitySlots = raw(slots)
	p.MainGoals = raw(goals)
	p.CurrentSkills = raw(current)
	p.TargetSkills = raw(target)
	p.HelpTopics = raw(topics)
	p.TargetStudentLevels = raw(levels)
	p.NotificationPreferences = raw(prefs)
	p.VerificationAssets = raw(assets)
	return p, nil
}

func raw(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return b
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
