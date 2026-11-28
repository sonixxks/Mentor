package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

const profileColumns = `p.id, p.user_id, u.role::text,
	p.first_name, p.last_name, p.middle_name, p.avatar_url,
	p.age, p.current_status::text, p.university, p.major, p.course, p.experience::text,
	p.job_title, p.company, p.bio, p.hourly_rate::float8, p.max_students, p.is_published,
	p.communication_style::text, p.availability_slots, p.main_goals, p.success_metric, p.time_limit,
	p.needs_curator_help, p.min_duration_months, p.max_duration_months,
	p.current_skills, p.target_skills, p.help_topics, p.target_student_levels, p.notification_preferences,
	p.onboarding_completed, p.verification_status::text, p.verification_assets,
	CASE WHEN u.role = 'mentor' THEN GREATEST(COALESCE(p.max_students, 0) - (
		SELECT count(*) FROM workspaces w JOIN requests r ON r.id = w.request_id
		WHERE r.mentor_id = p.user_id AND w.status <> 'completed'), 0)::int END`

const (
	byUserSQL     = `SELECT ` + profileColumns + ` FROM profiles p JOIN users u ON u.id = p.user_id WHERE p.user_id = $1`
	lockByUserSQL = byUserSQL + ` FOR UPDATE OF p`

	updateSQL = `
		UPDATE profiles SET
			first_name = $2, last_name = $3, middle_name = $4,
			age = $5, current_status = $6::current_status, university = $7, major = $8, course = $9,
			experience = $10::experience_level,
			job_title = $11, company = $12, bio = $13, hourly_rate = $14, max_students = $15, is_published = $16,
			communication_style = $17::communication_style, availability_slots = $18, main_goals = $19,
			success_metric = $20, time_limit = $21, needs_curator_help = $22,
			min_duration_months = $23, max_duration_months = $24,
			current_skills = $25, target_skills = $26, help_topics = $27, target_student_levels = $28,
			notification_preferences = $29, onboarding_completed = $30,
			verification_status = $31::verification_status, verification_assets = $32
		WHERE user_id = $1`

	canonicalSkillsSQL = `SELECT name FROM skills WHERE lower(name) = lower($1)`
)

type Repository struct {
	transactions TransactionManager
}

func NewRepository(transactions TransactionManager) *Repository {
	return &Repository{transactions: transactions}
}

type TransactionManager interface {
	Executor(ctx context.Context) postgres.DBTX
	ExecuteInTransaction(ctx context.Context, f func(context.Context) error) error
}

// Create — пустой профиль при подтверждении почты (вызывается в транзакции регистрации).
func (r *Repository) Create(ctx context.Context, p Profile) (Profile, error) {
	err := r.transactions.Executor(ctx).QueryRow(ctx, `INSERT INTO profiles(user_id,first_name,last_name,middle_name) VALUES ($1,$2,$3,$4) RETURNING id`, p.UserID, p.FirstName, p.LastName, p.MiddleName).Scan(&p.ID)
	return p, err
}

func (r *Repository) Get(ctx context.Context, userID uuid.UUID) (Profile, error) {
	p, err := scanProfile(r.transactions.Executor(ctx).QueryRow(ctx, byUserSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrProfileNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// Update блокирует профиль, даёт сервису изменить его (change) и сохраняет результат.
func (r *Repository) Update(ctx context.Context, userID uuid.UUID, change func(*Profile) error) (Profile, error) {
	var updated Profile
	err := r.transactions.ExecuteInTransaction(ctx, func(ctx context.Context) error {
		tx := r.transactions.Executor(ctx)
		p, err := scanProfile(tx.QueryRow(ctx, lockByUserSQL, userID))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		if err != nil {
			return fmt.Errorf("lock profile: %w", err)
		}

		if err := change(&p); err != nil {
			return err
		}
		if err := save(ctx, tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return Profile{}, err
	}
	return r.Get(ctx, updated.UserID)
}

// CanonicalSkill — название навыка из словаря (без учёта регистра) или "", если такого нет.
func (r *Repository) CanonicalSkill(ctx context.Context, name string) (string, error) {
	var canonical string
	err := r.transactions.Executor(ctx).QueryRow(ctx, canonicalSkillsSQL, strings.TrimSpace(name)).Scan(&canonical)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("lookup skill: %w", err)
	}
	return canonical, nil
}

func save(ctx context.Context, db postgres.DBTX, p Profile) error {
	jsonArgs := []any{p.AvailabilitySlots, p.MainGoals, p.CurrentSkills, p.TargetSkills,
		p.HelpTopics, p.TargetStudentLevels, p.NotificationPreferences, p.VerificationAssets}
	encoded := make([]any, len(jsonArgs))
	for i, v := range jsonArgs {
		b, err := encodeJSON(v)
		if err != nil {
			return err
		}
		encoded[i] = b
	}

	_, err := db.Exec(ctx, updateSQL, p.UserID,
		p.FirstName, p.LastName, p.MiddleName,
		p.Age, p.CurrentStatus, p.University, p.Major, p.Course, p.Experience,
		p.JobTitle, p.Company, p.Bio, p.HourlyRate, p.MaxStudents, p.IsPublished,
		p.CommunicationStyle, encoded[0], encoded[1], p.SuccessMetric, p.TimeLimit, p.NeedsCuratorHelp,
		p.MinDurationMonths, p.MaxDurationMonths,
		encoded[2], encoded[3], encoded[4], encoded[5], encoded[6], p.OnboardingCompleted,
		p.VerificationStatus, encoded[7],
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// encodeJSON: nil-значения пишутся как SQL NULL, а не как JSON null.
func encodeJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode profile json: %w", err)
	}
	if string(b) == "null" {
		return nil, nil
	}
	return b, nil
}

func scanProfile(row pgx.Row) (Profile, error) {
	var (
		p                                                    Profile
		role                                                 string
		slots, goals, current, target, topics, levels, prefs []byte
		assets                                               []byte
	)
	err := row.Scan(&p.ID, &p.UserID, &role,
		&p.FirstName, &p.LastName, &p.MiddleName, &p.AvatarURL,
		&p.Age, &p.CurrentStatus, &p.University, &p.Major, &p.Course, &p.Experience,
		&p.JobTitle, &p.Company, &p.Bio, &p.HourlyRate, &p.MaxStudents, &p.IsPublished,
		&p.CommunicationStyle, &slots, &goals, &p.SuccessMetric, &p.TimeLimit,
		&p.NeedsCuratorHelp, &p.MinDurationMonths, &p.MaxDurationMonths,
		&current, &target, &topics, &levels, &prefs,
		&p.OnboardingCompleted, &p.VerificationStatus, &assets, &p.FreeSlots,
	)
	if err != nil {
		return Profile{}, err
	}
	p.Role = user.Role(role)

	for _, item := range []struct {
		raw []byte
		dst any
	}{
		{slots, &p.AvailabilitySlots}, {goals, &p.MainGoals}, {current, &p.CurrentSkills},
		{target, &p.TargetSkills}, {topics, &p.HelpTopics}, {levels, &p.TargetStudentLevels},
		{prefs, &p.NotificationPreferences}, {assets, &p.VerificationAssets},
	} {
		// Пустые значения ('[]', '{}', null) считаем незаполненными.
		if len(item.raw) == 0 || string(item.raw) == "[]" || string(item.raw) == "{}" || string(item.raw) == "null" {
			continue
		}
		if err := json.Unmarshal(item.raw, item.dst); err != nil {
			return Profile{}, fmt.Errorf("decode profile json: %w", err)
		}
	}
	return p, nil
}
