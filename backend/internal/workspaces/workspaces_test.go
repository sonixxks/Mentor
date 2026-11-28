package workspaces

import (
	"context"
	"errors"
	"testing"

	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func TestList(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewRepository(pool)
	ctx := context.Background()

	mentor := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{MaxStudents: 5})
	anna := fx.Student("Анна", "Петрова")
	boris := fx.Student("Борис", "Иванов")

	old := fx.Workspace(boris, mentor, "active")
	fx.Exec(`UPDATE workspaces SET created_at = now() - interval '9 days' WHERE id = $1`, old)

	fresh := fx.Workspace(anna, mentor, "active")
	fx.Exec(`INSERT INTO tasks (workspace_id, title, status, deadline_at) VALUES
		($1, 'A', 'completed', now()), ($1, 'B', 'on_review', now()), ($1, 'C', 'todo', now()), ($1, 'D', 'todo', now())`, fresh)
	fx.Exec(`INSERT INTO messages (workspace_id, sender_id, text) VALUES ($1, $2, 'Привет'), ($1, $2, 'Есть вопрос')`, fresh, anna)

	list, total, err := repo.List(ctx, mentor, user.Mentor, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(list) != 2 || list[0].ID != fresh || list[1].ID != old {
		t.Fatalf("mentor list = %+v (total %d)", list, total)
	}
	first := list[0]
	if first.Partner.LastName != "Петрова" || first.ProgressPercent != 25 || first.UnreadMessages != 2 ||
		first.TasksOnReview == nil || *first.TasksOnReview != 1 || first.InactiveDays != nil {
		t.Fatalf("fresh pair = %+v", first)
	}
	if list[1].InactiveDays == nil || *list[1].InactiveDays != 9 {
		t.Fatalf("old pair inactive days = %v", list[1].InactiveDays)
	}

	list, total, err = repo.List(ctx, anna, user.Student, 10, 0)
	if err != nil || total != 1 || list[0].Partner.FirstName != "Иван" || list[0].UnreadMessages != 0 || list[0].TasksOnReview != nil {
		t.Fatalf("student list = %+v, err = %v", list, err)
	}

	if list, total, err := repo.List(ctx, mentor, user.Mentor, 10, 5); err != nil || len(list) != 0 || total != 2 {
		t.Fatalf("empty page: %v %d %v", list, total, err)
	}
	if _, _, err := repo.List(ctx, mentor, user.Curator, 10, 0); !errors.Is(err, ErrForbidden) {
		t.Fatalf("curator: %v", err)
	}
}
