package catalog

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type fakeRepo struct {
	mentors []MentorSummary
	got     Filter
}

func (f *fakeRepo) ListMentors(_ context.Context, filter Filter) ([]MentorSummary, int, error) {
	f.got = filter
	if filter.Limit > 0 {
		return paginate(f.mentors, filter.Limit, filter.Offset), len(f.mentors), nil
	}
	return append([]MentorSummary(nil), f.mentors...), len(f.mentors), nil
}

func (f *fakeRepo) GetMentor(context.Context, uuid.UUID, *uuid.UUID) (MentorDetail, error) {
	return MentorDetail{MentorSummary: f.mentors[0]}, nil
}

func (f *fakeRepo) ListReviews(context.Context, uuid.UUID, int, int) ([]Review, int, error) {
	return nil, 0, nil
}

func (f *fakeRepo) ListSkills(context.Context, SkillQuery) ([]string, error) { return nil, nil }

type fakeMatcher struct {
	percents map[uuid.UUID]int
	complete bool
}

func (m fakeMatcher) Match(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]MatchInfo, bool, error) {
	if !m.complete {
		return nil, false, nil
	}
	out := make(map[uuid.UUID]MatchInfo, len(ids))
	for _, id := range ids {
		out[id] = MatchInfo{Percent: m.percents[id]}
	}
	return out, true, nil
}

func threeMentors() ([]MentorSummary, map[uuid.UUID]int) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	// Порядок из БД — по рейтингу: a, b, c.
	return []MentorSummary{{ID: a, FirstName: "A"}, {ID: b, FirstName: "B"}, {ID: c, FirstName: "C"}},
		map[uuid.UUID]int{a: 40, b: 90, c: 70}
}

func TestListMentorsAnonymousFallsBackToRating(t *testing.T) {
	mentors, percents := threeMentors()
	repo := &fakeRepo{mentors: mentors}
	svc := NewService(repo, fakeMatcher{percents: percents, complete: true})

	minMatch := 50
	list, total, err := svc.ListMentors(context.Background(), nil, Filter{Limit: 10, MinMatch: &minMatch})
	if err != nil {
		t.Fatal(err)
	}
	if repo.got.Sort != SortRating || repo.got.MinMatch != nil {
		t.Fatalf("anonymous filter = %+v, want rating sort without min_match", repo.got)
	}
	if total != 3 || list[0].Match != nil {
		t.Fatalf("anonymous list must have no match: %+v", list)
	}
}

func TestListMentorsStudentSortedByMatch(t *testing.T) {
	mentors, percents := threeMentors()
	repo := &fakeRepo{mentors: mentors}
	svc := NewService(repo, fakeMatcher{percents: percents, complete: true})
	viewer := &Viewer{UserID: uuid.New(), Role: user.Student}

	minMatch := 50
	list, total, err := svc.ListMentors(context.Background(), viewer, Filter{Limit: 1, Offset: 1, MinMatch: &minMatch})
	if err != nil {
		t.Fatal(err)
	}
	if repo.got.Limit != 0 {
		t.Fatal("match sort must load all filtered mentors")
	}
	// После фильтра: B (90), C (70); вторая страница по одному — C.
	if total != 2 || len(list) != 1 || list[0].FirstName != "C" || list[0].Match.Percent != 70 {
		t.Fatalf("total = %d, list = %+v", total, list)
	}
}

func TestListMentorsStudentOtherSortUsesDBPage(t *testing.T) {
	mentors, percents := threeMentors()
	repo := &fakeRepo{mentors: mentors}
	svc := NewService(repo, fakeMatcher{percents: percents, complete: true})
	viewer := &Viewer{UserID: uuid.New(), Role: user.Student}

	list, _, err := svc.ListMentors(context.Background(), viewer, Filter{Sort: SortPriceAsc, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if repo.got.Limit != 2 || len(list) != 2 || list[0].Match == nil || list[0].Match.Percent != 40 {
		t.Fatalf("filter = %+v, list = %+v", repo.got, list)
	}
}

func TestListMentorsIncompleteProfile(t *testing.T) {
	mentors, _ := threeMentors()
	repo := &fakeRepo{mentors: mentors}
	svc := NewService(repo, fakeMatcher{complete: false})
	viewer := &Viewer{UserID: uuid.New(), Role: user.Student}

	list, total, err := svc.ListMentors(context.Background(), viewer, Filter{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if repo.got.Sort != SortRating || total != 3 || len(list) != 2 || list[0].Match != nil {
		t.Fatalf("filter = %+v, list = %+v", repo.got, list)
	}
}

func TestPaginate(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	if got := paginate(list, 2, 4); len(got) != 1 || got[0] != 5 {
		t.Fatalf("last page = %v", got)
	}
	if got := paginate(list, 2, 10); len(got) != 0 {
		t.Fatalf("beyond end = %v", got)
	}
}
