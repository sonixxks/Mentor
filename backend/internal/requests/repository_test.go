package requests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func noNotify(Request) []notifications.Notification { return nil }

func countNotifications(t *testing.T, repo *PostgresRepository, userID uuid.UUID) int {
	t.Helper()
	var n int
	if err := repo.pool.QueryRow(context.Background(), `SELECT count(*) FROM notifications WHERE user_id = $1`, userID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRepositoryCreateRules(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	st := fx.Student("Анна", "Петрова")
	m1 := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{})
	m2 := fx.Mentor("Олег", "Котов", pgtest.MentorOpts{})
	m3 := fx.Mentor("Мария", "Лебедева", pgtest.MentorOpts{})
	m4 := fx.Mentor("Пётр", "Орлов", pgtest.MentorOpts{})
	hidden := fx.Mentor("Скрытый", "Наставник", pgtest.MentorOpts{Unverified: true})
	full := fx.Mentor("Занятой", "Наставник", pgtest.MentorOpts{MaxStudents: 1})
	fx.Workspace(fx.Student("Другой", "Ученик"), full, "paused")
	otherPkg := fx.Package(m2, "Чужая программа", 4)

	create := func(mentorID uuid.UUID, pkg *uuid.UUID) error {
		_, err := repo.Create(ctx, CreateParams{StudentID: st, MentorID: mentorID, PackageID: pkg}, func(r Request) []notifications.Notification {
			return []notifications.Notification{notice(r.MentorID, "Новая заявка", nil, r.ID)}
		})
		return err
	}

	if err := create(m1, nil); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if countNotifications(t, repo, m1) != 1 {
		t.Fatal("mentor must get a notification in the same transaction")
	}
	if err := create(m1, nil); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate: error = %v, want %v", err, ErrDuplicate)
	}
	if err := create(hidden, nil); !errors.Is(err, ErrMentorNotFound) {
		t.Fatalf("unverified mentor: error = %v, want %v", err, ErrMentorNotFound)
	}
	if err := create(full, nil); !errors.Is(err, ErrMentorLimit) {
		t.Fatalf("full mentor: error = %v, want %v", err, ErrMentorLimit)
	}
	if err := create(m3, &otherPkg); !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("package of another mentor: error = %v, want %v", err, ErrPackageNotFound)
	}

	pkg := fx.Package(m2, "Подготовка к собеседованию", 4)
	if err := create(m2, &pkg); err != nil {
		t.Fatalf("request with package: %v", err)
	}
	if err := create(m3, nil); err != nil {
		t.Fatalf("third request: %v", err)
	}
	if err := create(m4, nil); !errors.Is(err, ErrRequestLimit) {
		t.Fatalf("fourth request: error = %v, want %v", err, ErrRequestLimit)
	}

	// Заявка куратора в лимит не входит, но нужен запрос помощи от ученика.
	curator := fx.Curator("Елена", "Куратор")
	suggest := CreateParams{StudentID: st, MentorID: m4, SuggestedBy: &curator}
	if _, err := repo.Create(ctx, suggest, noNotify); !errors.Is(err, ErrNoHelpRequest) {
		t.Fatalf("curator without help request: error = %v, want %v", err, ErrNoHelpRequest)
	}
	fx.Exec(`UPDATE profiles SET needs_curator_help = TRUE WHERE user_id = $1`, st)
	if _, err := repo.Create(ctx, suggest, noNotify); err != nil {
		t.Fatalf("curator request over student limit: %v", err)
	}
	if _, err := repo.Create(ctx, CreateParams{StudentID: m1, MentorID: m2, SuggestedBy: &curator}, noNotify); !errors.Is(err, ErrStudentNotFound) {
		t.Fatalf("mentor as student: error = %v, want %v", err, ErrStudentNotFound)
	}

	views, total, err := repo.List(ctx, ListFilter{StudentID: &st, Statuses: ActiveStatuses, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(views) != 2 {
		t.Fatalf("list: total = %d, len = %d, want 4 and 2", total, len(views))
	}
	for _, v := range views {
		if v.Student.FirstName != "Анна" || len(v.Student.TargetSkills) != 2 || v.Mentor.LastName == "" {
			t.Fatalf("view is not filled: %+v", v)
		}
	}
}

func TestRepositoryAccept(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewPostgresRepository(pool)
	svc := NewService(repo)
	ctx := context.Background()

	st := fx.Student("Анна", "Петрова")
	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{MaxStudents: 1})
	other := fx.Mentor("Олег", "Котов", pgtest.MentorOpts{})
	fx.Exec(`UPDATE profiles SET needs_curator_help = TRUE WHERE user_id = $1`, st)

	reqID := fx.Request(st, mentorID, "pending")
	otherReq := fx.Request(st, other, "pending")

	studentActor := Actor{UserID: st, Role: user.Student}
	mentorActor := Actor{UserID: mentorID, Role: user.Mentor}

	if _, err := svc.ChangeStatus(ctx, studentActor, ChangeStatusInput{RequestID: reqID, To: StatusAccepted}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("accept pending: error = %v, want %v", err, ErrInvalidTransition)
	}
	if _, err := svc.ChangeStatus(ctx, mentorActor, ChangeStatusInput{RequestID: reqID, To: StatusApprovedByMentor}); err != nil {
		t.Fatalf("approve: %v", err)
	}

	view, err := svc.ChangeStatus(ctx, studentActor, ChangeStatusInput{RequestID: reqID, To: StatusAccepted})
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if view.Status != StatusAccepted || view.WorkspaceID == nil {
		t.Fatalf("accepted view = %+v, want workspace", view)
	}

	otherView, err := repo.GetView(ctx, otherReq)
	if err != nil {
		t.Fatal(err)
	}
	if otherView.Status != StatusCancelled {
		t.Fatalf("other active request status = %s, want cancelled", otherView.Status)
	}
	if countNotifications(t, repo, other) != 1 {
		t.Fatal("mentor of the cancelled request must be notified")
	}

	var needsHelp bool
	if err := pool.QueryRow(ctx, `SELECT needs_curator_help FROM profiles WHERE user_id = $1`, st).Scan(&needsHelp); err != nil {
		t.Fatal(err)
	}
	if needsHelp {
		t.Fatal("needs_curator_help must be reset after accept")
	}

	// Место у наставника одно и уже занято.
	st2 := fx.Student("Борис", "Иванов")
	req2 := fx.Request(st2, mentorID, "approved_by_mentor")
	_, err = svc.ChangeStatus(ctx, Actor{UserID: st2, Role: user.Student}, ChangeStatusInput{RequestID: req2, To: StatusAccepted})
	if !errors.Is(err, ErrMentorLimit) {
		t.Fatalf("accept over limit: error = %v, want %v", err, ErrMentorLimit)
	}

	// Устаревшее состояние: статус уже сменился.
	if _, err := repo.UpdateStatus(ctx, reqID, StatusPending, StatusCancelled, nil, nil); !errors.Is(err, ErrStatusConflict) {
		t.Fatalf("stale update: error = %v, want %v", err, ErrStatusConflict)
	}
}
