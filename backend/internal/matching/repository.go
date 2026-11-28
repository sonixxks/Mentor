package matching

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	studentSQL = `
		SELECT p.user_id, p.onboarding_completed, p.target_skills, p.current_skills, p.main_goals,
		       p.availability_slots, p.time_limit, p.communication_style::text
		FROM profiles p
		WHERE p.user_id = ANY ($1)`

	mentorSQL = `
		SELECT p.user_id, p.current_skills, p.target_student_levels, p.help_topics,
		       p.availability_slots, p.communication_style::text
		FROM profiles p
		WHERE p.user_id = ANY ($1)`
)

// Repository загружает анкеты и считает совместимость.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Match — совместимость ученика с наставниками (каталог). ok = false, если анкета ученика не заполнена.
func (r *Repository) Match(ctx context.Context, studentID uuid.UUID, mentorIDs []uuid.UUID) (map[uuid.UUID]Info, bool, error) {
	students, err := r.students(ctx, []uuid.UUID{studentID})
	if err != nil {
		return nil, false, err
	}
	st, found := students[studentID]
	if !found || !st.completed {
		return nil, false, nil
	}

	mentors, err := r.mentors(ctx, mentorIDs)
	if err != nil {
		return nil, false, err
	}
	out := make(map[uuid.UUID]Info, len(mentors))
	for id, m := range mentors {
		out[id] = Calculate(st.Student, m)
	}
	return out, true, nil
}

// MatchStudents — совместимость учеников с наставником (входящие заявки наставника).
func (r *Repository) MatchStudents(ctx context.Context, mentorID uuid.UUID, studentIDs []uuid.UUID) (map[uuid.UUID]Info, error) {
	mentors, err := r.mentors(ctx, []uuid.UUID{mentorID})
	if err != nil {
		return nil, err
	}
	m, found := mentors[mentorID]
	if !found {
		return map[uuid.UUID]Info{}, nil
	}

	students, err := r.students(ctx, studentIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]Info, len(students))
	for id, st := range students {
		out[id] = Calculate(st.Student, m)
	}
	return out, nil
}

type studentRow struct {
	Student
	completed bool
}

func (r *Repository) students(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]studentRow, error) {
	rows, err := r.pool.Query(ctx, studentSQL, ids)
	if err != nil {
		return nil, fmt.Errorf("load students: %w", err)
	}
	out := map[uuid.UUID]studentRow{}
	_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (struct{}, error) {
		var (
			id                            uuid.UUID
			s                             studentRow
			target, current, goals, slots []byte
		)
		if err := row.Scan(&id, &s.completed, &target, &current, &goals, &slots, &s.TimeLimit, &s.CommunicationStyle); err != nil {
			return struct{}{}, err
		}
		if err := decodeAll(map[*[]byte]any{&target: &s.TargetSkills, &current: &s.CurrentSkills, &goals: &s.MainGoals, &slots: &s.AvailabilitySlots}); err != nil {
			return struct{}{}, err
		}
		out[id] = s
		return struct{}{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("load students: %w", err)
	}
	return out, nil
}

func (r *Repository) mentors(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]Mentor, error) {
	rows, err := r.pool.Query(ctx, mentorSQL, ids)
	if err != nil {
		return nil, fmt.Errorf("load mentors: %w", err)
	}
	out := map[uuid.UUID]Mentor{}
	_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (struct{}, error) {
		var (
			id                             uuid.UUID
			m                              Mentor
			current, levels, topics, slots []byte
		)
		if err := row.Scan(&id, &current, &levels, &topics, &slots, &m.CommunicationStyle); err != nil {
			return struct{}{}, err
		}
		if err := decodeAll(map[*[]byte]any{&current: &m.CurrentSkills, &levels: &m.TargetStudentLevels, &topics: &m.HelpTopics, &slots: &m.AvailabilitySlots}); err != nil {
			return struct{}{}, err
		}
		out[id] = m
		return struct{}{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("load mentors: %w", err)
	}
	return out, nil
}

// decodeAll разбирает JSONB-поля; пустые и не того типа (старые '[]' вместо объекта) пропускает.
func decodeAll(fields map[*[]byte]any) error {
	for raw, dst := range fields {
		if len(*raw) == 0 {
			continue
		}
		if err := json.Unmarshal(*raw, dst); err != nil {
			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &typeErr) {
				continue
			}
			return fmt.Errorf("decode profile json: %w", err)
		}
	}
	return nil
}
