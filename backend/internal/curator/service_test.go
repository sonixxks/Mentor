package curator

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
)

func newTestService(t *testing.T) (*Service, *pgtest.Fixtures, *PostgresRepository) {
	t.Helper()
	pool := pgtest.Pool(t)
	repo := NewPostgresRepository(pool)
	return NewService(repo), pgtest.NewFixtures(t, pool), repo
}

func notificationsOf(t *testing.T, repo *PostgresRepository, userID uuid.UUID) []string {
	t.Helper()
	rows, err := repo.pool.Query(context.Background(),
		`SELECT title FROM notifications WHERE user_id = $1 AND type = 'system_alert' ORDER BY created_at`, userID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var titles []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		titles = append(titles, s)
	}
	return titles
}

func TestListWorkspaces(t *testing.T) {
	svc, fx, _ := newTestService(t)
	ctx := context.Background()

	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{MaxStudents: 10})

	// Здоровая пара: свежее сообщение, одна задача из двух выполнена.
	healthy := fx.Workspace(fx.Student("Анна", "Петрова"), mentorID, "active")
	fx.Exec(`INSERT INTO messages (workspace_id, sender_id, text) SELECT $1, r.student_id, 'Привет'
		FROM workspaces w JOIN requests r ON r.id = w.request_id WHERE w.id = $1`, healthy)
	fx.Exec(`INSERT INTO tasks (workspace_id, title, status, deadline_at) VALUES
		($1, 'Резюме', 'completed', now() + interval '1 day'), ($1, 'Pet-проект', 'todo', now() + interval '5 days')`, healthy)

	// Проблемная пара: создана 10 дней назад, две просроченные задачи созданы тогда же.
	problem := fx.Workspace(fx.Student("Борис", "Иванов"), mentorID, "active")
	fx.Exec(`UPDATE workspaces SET created_at = now() - interval '10 days' WHERE id = $1`, problem)
	fx.Exec(`INSERT INTO tasks (workspace_id, title, deadline_at, created_at, updated_at) VALUES
		($1, 'A', now() - interval '3 days', now() - interval '10 days', now() - interval '10 days'),
		($1, 'B', now() - interval '2 days', now() - interval '10 days', now() - interval '10 days')`, problem)

	// Пара на паузе и завершённая пара.
	fx.Workspace(fx.Student("Вера", "Соколова"), mentorID, "paused")
	fx.Workspace(fx.Student("Глеб", "Морозов"), mentorID, "completed")

	page, err := svc.ListWorkspaces(ctx, WorkspaceFilter{Sort: SortSeverity, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || page.Counts != (StatusCounts{Total: 3, Active: 2, Paused: 1}) {
		t.Fatalf("total = %d, counts = %+v", page.Total, page.Counts)
	}

	first := page.Items[0]
	if first.ID != problem || first.Severity != SeverityRed || len(first.Signals) != 2 {
		t.Fatalf("first row = %+v, want red problem pair with 2 signals", first)
	}
	for _, row := range page.Items {
		if row.ID == healthy && (row.Severity != SeverityGreen || row.ProgressPercent != 50 || row.LastActivityAt == nil) {
			t.Fatalf("healthy row = %+v", row)
		}
	}

	withSignals := true
	page, err = svc.ListWorkspaces(ctx, WorkspaceFilter{HasSignals: &withSignals, Limit: 10})
	if err != nil || page.Total != 1 {
		t.Fatalf("has_signals: total = %d, err = %v", page.Total, err)
	}

	page, err = svc.ListWorkspaces(ctx, WorkspaceFilter{Query: "соколова", Limit: 10})
	if err != nil || page.Total != 1 || page.Items[0].StudentName != "Вера Соколова" {
		t.Fatalf("search: %+v, err = %v", page.Items, err)
	}
}

func TestSetWorkspaceStatusAndNotify(t *testing.T) {
	svc, fx, repo := newTestService(t)
	ctx := context.Background()

	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{MaxStudents: 1})
	studentID := fx.Student("Анна", "Петрова")
	done := fx.Workspace(studentID, mentorID, "completed")
	fx.Workspace(fx.Student("Борис", "Иванов"), mentorID, "active")

	if _, err := svc.SetWorkspaceStatus(ctx, done, WorkspaceActive); !errors.Is(err, ErrMentorLimit) {
		t.Fatalf("reopen over limit: error = %v, want %v", err, ErrMentorLimit)
	}
	if _, err := svc.SetWorkspaceStatus(ctx, done, "archived"); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("bad status: error = %v", err)
	}
	if _, err := svc.SetWorkspaceStatus(ctx, uuid.New(), WorkspacePaused); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("missing workspace: error = %v", err)
	}

	row, err := svc.SetWorkspaceStatus(ctx, done, WorkspaceStagnation)
	if !errors.Is(err, ErrMentorLimit) {
		t.Fatalf("stagnation also occupies a slot: row = %+v, err = %v", row, err)
	}

	fx.Exec(`UPDATE profiles SET max_students = 2 WHERE user_id = $1`, mentorID)
	row, err = svc.SetWorkspaceStatus(ctx, done, WorkspacePaused)
	if err != nil || row.Status != WorkspacePaused {
		t.Fatalf("reopen with free slot: row = %+v, err = %v", row, err)
	}
	if got := notificationsOf(t, repo, studentID); len(got) != 1 {
		t.Fatalf("student notifications = %v", got)
	}

	if err := svc.NotifyWorkspace(ctx, done, TargetMentor, "  Как дела у пары?  "); err != nil {
		t.Fatal(err)
	}
	if got := notificationsOf(t, repo, mentorID); len(got) != 2 || got[1] != "Сообщение от куратора" {
		t.Fatalf("mentor notifications = %v", got)
	}
	if err := svc.NotifyWorkspace(ctx, done, "everyone", "x"); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("bad target: error = %v", err)
	}
	if err := svc.NotifyWorkspace(ctx, done, TargetBoth, " "); !errors.Is(err, ErrCommentRequired) {
		t.Fatalf("empty message: error = %v", err)
	}
}

func profileID(t *testing.T, repo *PostgresRepository, userID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := repo.pool.QueryRow(context.Background(), `SELECT id FROM profiles WHERE user_id = $1`, userID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestProfilesQueue(t *testing.T) {
	svc, fx, repo := newTestService(t)
	ctx := context.Background()

	student := fx.Student("Анна", "Петрова")
	fx.Exec(`UPDATE profiles SET needs_curator_help = TRUE WHERE user_id = $1`, student)
	mentor := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{Unverified: true})
	fx.Mentor("Олег", "Котов", pgtest.MentorOpts{})

	needsHelp := true
	list, total, err := svc.ListProfiles(ctx, ProfileFilter{NeedsHelp: &needsHelp, Limit: 10})
	if err != nil || total != 1 || list[0].UserID != student || list[0].Email == "" || list[0].FreeSlots != nil {
		t.Fatalf("help queue: %+v, total = %d, err = %v", list, total, err)
	}

	pending := "pending"
	list, total, err = svc.ListProfiles(ctx, ProfileFilter{VerificationStatus: &pending, Limit: 10})
	if err != nil || total != 1 || list[0].UserID != mentor || list[0].FreeSlots == nil || *list[0].FreeSlots != 3 {
		t.Fatalf("verification queue: %+v, total = %d, err = %v", list, total, err)
	}

	studentProfile := profileID(t, repo, student)
	p, err := svc.ResolveHelpRequest(ctx, studentProfile, "Подобрали двух наставников")
	if err != nil || p.NeedsCuratorHelp {
		t.Fatalf("resolve help: %+v, err = %v", p, err)
	}
	if _, err := svc.ResolveHelpRequest(ctx, studentProfile, ""); !errors.Is(err, ErrNoHelpRequest) {
		t.Fatalf("resolve twice: error = %v", err)
	}
	if got := notificationsOf(t, repo, student); len(got) != 1 {
		t.Fatalf("student notifications = %v", got)
	}

	mentorProfile := profileID(t, repo, mentor)
	if err := svc.RequestDocuments(ctx, studentProfile, "Пришлите диплом"); !errors.Is(err, ErrNotMentor) {
		t.Fatalf("documents from student: error = %v", err)
	}
	if err := svc.RequestDocuments(ctx, mentorProfile, "Пришлите диплом"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetVerification(ctx, mentorProfile, VerificationRejected, ""); !errors.Is(err, ErrCommentRequired) {
		t.Fatalf("reject without comment: error = %v", err)
	}
	p, err = svc.SetVerification(ctx, mentorProfile, VerificationVerified, "")
	if err != nil || p.VerificationStatus != "verified" {
		t.Fatalf("verify: %+v, err = %v", p, err)
	}
	if _, err := svc.SetVerification(ctx, mentorProfile, VerificationRejected, "поздно"); !errors.Is(err, ErrNotPending) {
		t.Fatalf("verify twice: error = %v", err)
	}
	if got := notificationsOf(t, repo, mentor); len(got) != 2 {
		t.Fatalf("mentor notifications = %v", got)
	}
	if _, err := svc.ResolveHelpRequest(ctx, uuid.New(), ""); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("missing profile: error = %v", err)
	}
}
