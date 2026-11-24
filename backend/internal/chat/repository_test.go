package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func TestChatFlow(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	svc := NewService(NewPostgresRepository(pool))
	ctx := context.Background()

	studentID := fx.Student("Анна", "Петрова")
	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{})
	ws := fx.Workspace(studentID, mentorID, "active")
	otherWS := fx.Workspace(fx.Student("Борис", "Иванов"), mentorID, "active")

	student := Actor{UserID: studentID, Role: user.Student}
	mentor := Actor{UserID: mentorID, Role: user.Mentor}

	unreadNotices := func(userID uuid.UUID) int {
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id = $1 AND type = 'new_message' AND NOT is_read`, userID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	for _, text := range []string{"Привет!", "Когда созвон?", "Скинул резюме"} {
		if _, err := svc.Send(ctx, student, ws, SendInput{Text: text}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}
	if n := unreadNotices(mentorID); n != 1 {
		t.Fatalf("mentor unread notices = %d, want 1 per pair", n)
	}

	// Файл из базы знаний пары можно приложить, чужой — нельзя.
	fileID, foreignFile := uuid.New(), uuid.New()
	fx.Exec(`INSERT INTO knowledge_base (id, workspace_id, file_name, file_url, file_size_kb, uploaded_by)
		VALUES ($1, $2, 'cv.pdf', 'https://files.test/cv.pdf', 120, $3), ($4, $5, 'x.pdf', 'https://files.test/x.pdf', 1, $3)`,
		fileID, ws, mentorID, foreignFile, otherWS)
	msg, err := svc.Send(ctx, mentor, ws, SendInput{FileID: &fileID})
	if err != nil || msg.File == nil || msg.File.FileName != "cv.pdf" || msg.Text != nil {
		t.Fatalf("file message = %+v, err = %v", msg, err)
	}
	if _, err := svc.Send(ctx, mentor, ws, SendInput{FileID: &foreignFile}); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("foreign file: error = %v", err)
	}
	if err := InsertSystemMessage(ctx, pool, ws, "Встреча назначена"); err != nil {
		t.Fatal(err)
	}

	all, err := svc.History(ctx, mentor, ws, nil, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 5 || *all[0].Text != "Привет!" || all[4].Type != TypeSystem || all[4].SenderID != nil {
		t.Fatalf("history = %+v", all)
	}

	// Страница назад: два сообщения до четвёртого.
	page, err := svc.History(ctx, mentor, ws, &all[3].CreatedAt, 2)
	if err != nil || len(page) != 2 || *page[0].Text != "Когда созвон?" || *page[1].Text != "Скинул резюме" {
		t.Fatalf("page = %+v, err = %v", page, err)
	}

	if err := svc.MarkRead(ctx, mentor, ws); err != nil {
		t.Fatal(err)
	}
	all, _ = svc.History(ctx, student, ws, nil, 50)
	for _, m := range all {
		fromMentor := m.SenderID != nil && *m.SenderID == mentorID
		if m.IsRead == fromMentor {
			t.Fatalf("after mentor read: message %+v has is_read = %v", m, m.IsRead)
		}
	}
	if n := unreadNotices(mentorID); n != 0 {
		t.Fatalf("mentor notices after read = %d", n)
	}

	if _, err := svc.History(ctx, Actor{UserID: fx.Student("Чужой", "Ученик"), Role: user.Student}, ws, nil, 10); !errors.Is(err, ErrNotMember) {
		t.Fatalf("stranger: error = %v", err)
	}
	if _, err := svc.History(ctx, student, uuid.New(), nil, 10); !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("missing workspace: error = %v", err)
	}
}
