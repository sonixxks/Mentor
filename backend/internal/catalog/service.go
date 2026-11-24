package catalog

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Repository interface {
	ListMentors(ctx context.Context, f Filter) ([]MentorSummary, int, error)
	GetMentor(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (MentorDetail, error)
	ListReviews(ctx context.Context, mentorID uuid.UUID, limit, offset int) ([]Review, int, error)
	ListSkills(ctx context.Context, q SkillQuery) ([]string, error)
}

// Viewer — авторизованный пользователь; nil для анонимного запроса.
type Viewer struct {
	UserID uuid.UUID
	Role   user.Role
}

type Service struct {
	repo    Repository
	matcher Matcher
}

// NewService: matcher может быть nil, пока модуль подбора не подключён —
// тогда каталог работает без процента совместимости.
func NewService(repo Repository, matcher Matcher) *Service {
	return &Service{repo: repo, matcher: matcher}
}

// ListMentors — каталог с фильтрами. Для ученика с заполненной анкетой в карточках
// есть match, работают сортировка по совместимости и фильтр min_match.
// Без этого сортировка match заменяется на rating, а min_match не применяется.
func (s *Service) ListMentors(ctx context.Context, viewer *Viewer, f Filter) ([]MentorSummary, int, error) {
	if f.Sort == "" {
		f.Sort = SortMatch
	}

	canMatch := s.matcher != nil && viewer != nil && viewer.Role == user.Student
	if !canMatch {
		return s.listWithoutMatch(ctx, f)
	}

	// Без сортировки и фильтра по совместимости хватает одной страницы из БД.
	if f.Sort != SortMatch && f.MinMatch == nil {
		list, total, err := s.repo.ListMentors(ctx, f)
		if err != nil {
			return nil, 0, err
		}
		if _, err := s.attachMatch(ctx, viewer.UserID, list); err != nil {
			return nil, 0, err
		}
		return list, total, nil
	}

	// Совместимость считается в коде, поэтому берём всех подходящих по фильтрам
	// (наставников в MVP немного) и пагинируем после сортировки.
	limit, offset := f.Limit, f.Offset
	f.Limit, f.Offset = 0, 0

	all, _, err := s.repo.ListMentors(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	ok, err := s.attachMatch(ctx, viewer.UserID, all)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		// Анкета не заполнена: ведём себя как для гостя.
		f.Limit, f.Offset = limit, offset
		return s.listWithoutMatch(ctx, f)
	}

	if f.MinMatch != nil {
		filtered := all[:0]
		for _, m := range all {
			if m.Match != nil && m.Match.Percent >= *f.MinMatch {
				filtered = append(filtered, m)
			}
		}
		all = filtered
	}

	if f.Sort == SortMatch {
		// Стабильно: при равной совместимости остаётся порядок по рейтингу из БД.
		sort.SliceStable(all, func(i, j int) bool {
			return matchPercent(all[i]) > matchPercent(all[j])
		})
	}

	return paginate(all, limit, offset), len(all), nil
}

func (s *Service) listWithoutMatch(ctx context.Context, f Filter) ([]MentorSummary, int, error) {
	if f.Sort == SortMatch {
		f.Sort = SortRating
	}
	f.MinMatch = nil
	return s.repo.ListMentors(ctx, f)
}

func (s *Service) attachMatch(ctx context.Context, studentID uuid.UUID, list []MentorSummary) (bool, error) {
	if len(list) == 0 {
		return true, nil
	}
	ids := make([]uuid.UUID, len(list))
	for i, m := range list {
		ids[i] = m.ID
	}

	matches, ok, err := s.matcher.Match(ctx, studentID, ids)
	if err != nil {
		return false, fmt.Errorf("match mentors: %w", err)
	}
	if !ok {
		return false, nil
	}
	for i := range list {
		if m, found := matches[list[i].ID]; found {
			list[i].Match = &m
		}
	}
	return true, nil
}

func (s *Service) GetMentor(ctx context.Context, viewer *Viewer, id uuid.UUID) (MentorDetail, error) {
	var viewerID *uuid.UUID
	if viewer != nil {
		viewerID = &viewer.UserID
	}

	d, err := s.repo.GetMentor(ctx, id, viewerID)
	if err != nil {
		return MentorDetail{}, err
	}

	if s.matcher != nil && viewer != nil && viewer.Role == user.Student {
		list := []MentorSummary{d.MentorSummary}
		if _, err := s.attachMatch(ctx, viewer.UserID, list); err != nil {
			return MentorDetail{}, err
		}
		d.Match = list[0].Match
	}
	return d, nil
}

func (s *Service) ListReviews(ctx context.Context, mentorID uuid.UUID, limit, offset int) ([]Review, int, error) {
	return s.repo.ListReviews(ctx, mentorID, limit, offset)
}

func (s *Service) ListSkills(ctx context.Context, q SkillQuery) ([]string, error) {
	return s.repo.ListSkills(ctx, q)
}

func matchPercent(m MentorSummary) int {
	if m.Match == nil {
		return -1
	}
	return m.Match.Percent
}

func paginate[T any](list []T, limit, offset int) []T {
	if offset >= len(list) {
		return []T{}
	}
	end := len(list)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return list[offset:end]
}
