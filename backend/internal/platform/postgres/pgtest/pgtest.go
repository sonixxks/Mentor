// Package pgtest — помощники для интеграционных тестов репозиториев.
//
// Тесты запускаются, только если задана TEST_DATABASE_URL с накатанными миграциями:
//
//	TEST_DATABASE_URL=postgres://user:pass@127.0.0.1:5432/mentor_test?sslmode=disable go test ./...
//
// Каждый тест очищает таблицы, поэтому не направляйте её на рабочую базу.
package pgtest

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const envURL = "TEST_DATABASE_URL"

// Pool возвращает пул к тестовой базе с пустыми таблицами или пропускает тест.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(envURL)
	if dsn == "" {
		t.Skip(envURL + " is not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	// go test запускает пакеты параллельно, а база одна: держим advisory-lock
	// на время теста, чтобы тесты разных пакетов не чистили таблицы друг другу.
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire connection: %v", err)
	}
	if _, err := conn.Exec(context.Background(), `SELECT pg_advisory_lock($1)`, lockKey); err != nil {
		t.Fatalf("advisory lock: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)
		conn.Release()
	})

	Truncate(t, pool)
	return pool
}

const lockKey = 7_420_001

// Truncate очищает все таблицы с данными пользователей (словарь навыков не трогает).
func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE users, profiles, service_packages, requests, workspaces, tasks, task_comments,
		         knowledge_base, session_reviews, meetings, meeting_notes, messages, notifications
		CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// Fixtures создаёт типовые сущности для тестов.
type Fixtures struct {
	t    *testing.T
	pool *pgxpool.Pool
}

func NewFixtures(t *testing.T, pool *pgxpool.Pool) *Fixtures {
	return &Fixtures{t: t, pool: pool}
}

func (f *Fixtures) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("fixture: %v\n%s", err, sql)
	}
}

func (f *Fixtures) user(role, firstName, lastName string) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, 'x', $3::user_role)`,
		id, id.String()+"@test.local", role)
	f.exec(`INSERT INTO profiles (user_id, first_name, last_name) VALUES ($1, $2, $3)`, id, firstName, lastName)
	return id
}

// Student — ученик с заполненной анкетой.
func (f *Fixtures) Student(firstName, lastName string) uuid.UUID {
	f.t.Helper()
	id := f.user("student", firstName, lastName)
	f.exec(`UPDATE profiles SET
		target_skills = '["Go", "SQL"]', main_goals = '["find_job_or_internship"]',
		time_limit = 4, min_duration_months = 2, max_duration_months = 6,
		communication_style = 'friendly', success_metric = 'Найти стажировку',
		onboarding_completed = TRUE
		WHERE user_id = $1`, id)
	return id
}

// MentorOpts — параметры наставника; нулевые значения заменяются типовыми.
type MentorOpts struct {
	MaxStudents int
	HourlyRate  float64
	TimeLimit   int
	Experience  string
	Skills      string // JSON, например [{"skill":"Go","level":3}]
	Company     string
	Unverified  bool
	Unpublished bool
	Style       string
	JobTitle    string
}

// Mentor — опубликованный верифицированный наставник.
func (f *Fixtures) Mentor(firstName, lastName string, o MentorOpts) uuid.UUID {
	f.t.Helper()
	if o.MaxStudents == 0 {
		o.MaxStudents = 3
	}
	if o.TimeLimit == 0 {
		o.TimeLimit = 5
	}
	if o.Experience == "" {
		o.Experience = "3_to_5"
	}
	if o.Skills == "" {
		o.Skills = `[{"skill":"Go","level":3},{"skill":"SQL","level":2}]`
	}
	if o.Style == "" {
		o.Style = "structured"
	}
	if o.JobTitle == "" {
		o.JobTitle = "Backend-разработчик"
	}
	verification := "verified"
	if o.Unverified {
		verification = "pending"
	}

	id := f.user("mentor", firstName, lastName)
	f.exec(`UPDATE profiles SET
		job_title = $2, company = NULLIF($3, ''), experience = $4::experience_level,
		bio = 'Помогаю войти в IT', current_skills = $5::jsonb,
		help_topics = '["interview_prep","code_review"]', target_student_levels = '["beginner","has_basics"]',
		hourly_rate = $6, time_limit = $7, max_students = $8,
		availability_slots = '{"mon":[19,20],"sat":[10,11]}', communication_style = $9::communication_style,
		is_published = $10, verification_status = $11::verification_status, onboarding_completed = TRUE
		WHERE user_id = $1`,
		id, o.JobTitle, o.Company, o.Experience, o.Skills, o.HourlyRate, o.TimeLimit, o.MaxStudents,
		o.Style, !o.Unpublished, verification)
	return id
}

func (f *Fixtures) Curator(firstName, lastName string) uuid.UUID {
	f.t.Helper()
	return f.user("curator", firstName, lastName)
}

// Package — опубликованная программа наставника.
func (f *Fixtures) Package(mentorID uuid.UUID, title string, sessions int) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO service_packages (id, mentor_id, title, price, session_count, status)
		VALUES ($1, $2, $3, 1500, $4, 'published')`, id, mentorID, title, sessions)
	return id
}

// Request — заявка в нужном статусе.
func (f *Fixtures) Request(studentID, mentorID uuid.UUID, status string) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO requests (id, student_id, mentor_id, status) VALUES ($1, $2, $3, $4::request_status)`,
		id, studentID, mentorID, status)
	return id
}

// Workspace — пара (принятая заявка и воркспейс) в нужном статусе.
func (f *Fixtures) Workspace(studentID, mentorID uuid.UUID, status string) uuid.UUID {
	f.t.Helper()
	requestID := f.Request(studentID, mentorID, "accepted")
	id := uuid.New()
	f.exec(`INSERT INTO workspaces (id, request_id, jitsi_link, status) VALUES ($1, $2, 'https://meet.jit.si/test', $3::workspace_status)`,
		id, requestID, status)
	return id
}

// Exec — произвольная подготовка данных.
func (f *Fixtures) Exec(sql string, args ...any) {
	f.t.Helper()
	f.exec(sql, args...)
}
