package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
)

func names(list []MentorSummary) []string {
	out := make([]string, len(list))
	for i, m := range list {
		out[i] = m.FirstName
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRepositoryListMentors(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	goMentor := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{
		HourlyRate: 2000, TimeLimit: 6, Experience: "5_to_10", Company: "Яндекс", MaxStudents: 1,
	})
	pyMentor := fx.Mentor("Мария", "Лебедева", pgtest.MentorOpts{
		HourlyRate: 1000, TimeLimit: 3, Experience: "1_to_3",
		Skills: `[{"skill":"Python","level":3},{"skill":"SQL","level":2}]`,
	})
	freeMentor := fx.Mentor("Олег", "Котов", pgtest.MentorOpts{
		HourlyRate: 0, TimeLimit: 10, Experience: "over_10", Skills: `[{"skill":"React","level":3}]`,
	})
	fx.Mentor("Скрытый", "Наставник", pgtest.MentorOpts{Unpublished: true})
	fx.Mentor("Непроверенный", "Наставник", pgtest.MentorOpts{Unverified: true})

	// У Ивана единственное место занято, у Марии отзывы 5 и 3, у Олега 5.
	st := fx.Student("Анна", "Петрова")
	ws := fx.Workspace(st, goMentor, "active")
	meeting := uuid.New()
	fx.Exec(`INSERT INTO meetings (id, workspace_id, start_time) VALUES ($1, $2, now())`, meeting, ws)
	for _, rv := range []struct {
		target uuid.UUID
		rating int
	}{{pyMentor, 5}, {pyMentor, 3}, {freeMentor, 5}} {
		author := fx.Student("Автор", "Отзыва")
		fx.Exec(`INSERT INTO session_reviews (workspace_id, meeting_id, author_id, target_id, rating)
			VALUES ($1, $2, $3, $4, $5)`, ws, meeting, author, rv.target, rv.rating)
	}

	tests := []struct {
		name   string
		filter Filter
		want   []string
		total  int
	}{
		{"all by rating", Filter{Sort: SortRating}, []string{"Олег", "Мария", "Иван"}, 3},
		{"price asc", Filter{Sort: SortPriceAsc}, []string{"Олег", "Мария", "Иван"}, 3},
		{"price desc", Filter{Sort: SortPriceDesc}, []string{"Иван", "Мария", "Олег"}, 3},
		{"experience", Filter{Sort: SortExperience}, []string{"Олег", "Иван", "Мария"}, 3},
		{"skill filter any of, case-insensitive", Filter{Skills: []string{"sql"}, Sort: SortRating}, []string{"Мария", "Иван"}, 2},
		{"search by company", Filter{Query: "яндекс"}, []string{"Иван"}, 1},
		{"search by skill", Filter{Query: "pyth"}, []string{"Мария"}, 1},
		{"search escapes LIKE", Filter{Query: "%"}, []string{}, 0},
		{"price range", Filter{PriceMin: ptr(500.0), PriceMax: ptr(1500.0)}, []string{"Мария"}, 1},
		{"hours range", Filter{HoursMin: ptr(5), HoursMax: ptr(8)}, []string{"Иван"}, 1},
		{"min rating", Filter{MinRating: ptr(4.0), Sort: SortRating}, []string{"Олег", "Мария"}, 2},
		{"free slots only", Filter{HasFreeSlots: true, Sort: SortRating}, []string{"Олег", "Мария"}, 2},
		{"pagination", Filter{Sort: SortRating, Limit: 1, Offset: 1}, []string{"Мария"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, total, err := repo.ListMentors(ctx, tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if !equal(names(list), tt.want) || total != tt.total {
				t.Fatalf("got %v (total %d), want %v (total %d)", names(list), total, tt.want, tt.total)
			}
		})
	}

	list, _, err := repo.ListMentors(ctx, Filter{Query: "Мария"})
	if err != nil {
		t.Fatal(err)
	}
	m := list[0]
	if m.RatingAvg == nil || *m.RatingAvg != 4 || m.ReviewsCount != 2 || m.FreeSlots != 3 || len(m.CurrentSkills) != 2 {
		t.Fatalf("summary = %+v", m)
	}
}

func TestRepositoryGetMentor(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{})
	fx.Package(mentorID, "Собеседования", 4)
	fx.Exec(`INSERT INTO service_packages (mentor_id, title, status) VALUES ($1, 'Черновик', 'draft')`, mentorID)

	d, err := repo.GetMentor(ctx, mentorID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.ServicePackages) != 1 || d.ServicePackages[0].Title != "Собеседования" {
		t.Fatalf("only published packages expected: %+v", d.ServicePackages)
	}
	if len(d.HelpTopics) != 2 || len(d.AvailabilitySlots) == 0 || d.MaxStudents == nil {
		t.Fatalf("detail = %+v", d)
	}

	hidden := fx.Mentor("Скрытый", "Наставник", pgtest.MentorOpts{Unverified: true})
	if _, err := repo.GetMentor(ctx, hidden, nil); !errors.Is(err, ErrMentorNotFound) {
		t.Fatalf("hidden mentor for guest: error = %v", err)
	}
	other := uuid.New()
	if _, err := repo.GetMentor(ctx, hidden, &other); !errors.Is(err, ErrMentorNotFound) {
		t.Fatalf("hidden mentor for another user: error = %v", err)
	}
	if d, err := repo.GetMentor(ctx, hidden, &hidden); err != nil || d.VerificationStatus != "pending" {
		t.Fatalf("owner must see own hidden card: %v, %+v", err, d)
	}
}

func TestRepositoryReviewsAndSkills(t *testing.T) {
	pool := pgtest.Pool(t)
	fx := pgtest.NewFixtures(t, pool)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	mentorID := fx.Mentor("Иван", "Смирнов", pgtest.MentorOpts{})
	st := fx.Student("Анна", "Петрова")
	ws := fx.Workspace(st, mentorID, "active")
	meeting := uuid.New()
	fx.Exec(`INSERT INTO meetings (id, workspace_id, start_time) VALUES ($1, $2, now() - interval '1 day')`, meeting, ws)
	fx.Exec(`INSERT INTO session_reviews (workspace_id, meeting_id, author_id, target_id, rating, comment)
		VALUES ($1, $2, $3, $4, 5, 'Отлично'), ($1, $2, $4, $3, 2, 'Оценка ученика видна только куратору')`,
		ws, meeting, st, mentorID)

	list, total, err := repo.ListReviews(ctx, mentorID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || list[0].AuthorFirstName != "Анна" || list[0].MeetingDate == nil {
		t.Fatalf("reviews = %+v (total %d)", list, total)
	}
	if _, _, err := repo.ListReviews(ctx, st, 10, 0); !errors.Is(err, ErrMentorNotFound) {
		t.Fatalf("reviews of a student: error = %v", err)
	}

	skills, err := repo.ListSkills(ctx, SkillQuery{Query: "java", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if !equal(skills, []string{"Java", "JavaScript"}) {
		t.Fatalf("skills = %v", skills)
	}
	popular, err := repo.ListSkills(ctx, SkillQuery{Popular: true, Limit: 3})
	if err != nil || len(popular) != 3 {
		t.Fatalf("popular = %v, %v", popular, err)
	}
}

func ptr[T any](v T) *T { return &v }
