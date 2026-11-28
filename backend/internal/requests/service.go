package requests

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Repository interface {
	// Create проверяет наставника, программу и лимиты и создаёт заявку в одной транзакции.
	Create(ctx context.Context, p CreateParams, notify func(Request) []notifications.Notification) (Request, error)
	GetByID(ctx context.Context, id uuid.UUID) (Request, error)
	GetView(ctx context.Context, id uuid.UUID) (View, error)
	List(ctx context.Context, f ListFilter) ([]View, int, error)
	// UpdateStatus меняет статус, только если он всё ещё from (защита от гонок).
	UpdateStatus(ctx context.Context, id uuid.UUID, from, to Status, reason *string, notify []notifications.Notification) (Request, error)
	// Accept переводит заявку в accepted и создаёт воркспейс (см. описание в репозитории).
	Accept(ctx context.Context, id uuid.UUID, notify []notifications.Notification) (Request, uuid.UUID, error)
}

// StudentMatcher — совместимость учеников с наставником (пакет matching).
type StudentMatcher interface {
	MatchStudents(ctx context.Context, mentorID uuid.UUID, studentIDs []uuid.UUID) (map[uuid.UUID]MatchInfo, error)
}

type Service struct {
	repo    Repository
	matcher StudentMatcher
}

// NewService: matcher может быть nil — тогда заявки отдаются без match.
func NewService(repo Repository, matcher StudentMatcher) *Service {
	return &Service{repo: repo, matcher: matcher}
}

// attachMatch добавляет совместимость в заявки, которые видит наставник.
func (s *Service) attachMatch(ctx context.Context, actor Actor, views []View) error {
	if s.matcher == nil || actor.Role != user.Mentor || len(views) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(views))
	for i, v := range views {
		ids[i] = v.StudentID
	}
	matches, err := s.matcher.MatchStudents(ctx, actor.UserID, ids)
	if err != nil {
		return fmt.Errorf("match students: %w", err)
	}
	for i := range views {
		if m, ok := matches[views[i].StudentID]; ok {
			views[i].Match = &m
		}
	}
	return nil
}

type CreateInput struct {
	MentorID    uuid.UUID
	PackageID   *uuid.UUID
	CoverLetter string
}

// Create — заявка ученика наставнику (открытый запрос или на программу).
func (s *Service) Create(ctx context.Context, actor Actor, in CreateInput) (View, error) {
	if actor.Role != user.Student {
		return View{}, ErrForbidden
	}

	letter, err := normalizeCoverLetter(in.CoverLetter)
	if err != nil {
		return View{}, err
	}

	req, err := s.repo.Create(ctx, CreateParams{
		StudentID:   actor.UserID,
		MentorID:    in.MentorID,
		PackageID:   in.PackageID,
		CoverLetter: letter,
	}, func(r Request) []notifications.Notification {
		return []notifications.Notification{
			notice(r.MentorID, "Новая заявка на менторство", nil, r.ID),
		}
	})
	if err != nil {
		return View{}, err
	}
	return s.repo.GetView(ctx, req.ID)
}

type CuratorCreateInput struct {
	StudentID uuid.UUID
	MentorID  uuid.UUID
	PackageID *uuid.UUID
}

// CreateByCurator — ручной подбор: обычная заявка pending от ученика к наставнику,
// не входит в лимит ученика. Наставник одобряет, ученик подтверждает.
func (s *Service) CreateByCurator(ctx context.Context, actor Actor, in CuratorCreateInput) (View, error) {
	if actor.Role != user.Curator {
		return View{}, ErrForbidden
	}

	curatorID := actor.UserID
	req, err := s.repo.Create(ctx, CreateParams{
		StudentID:   in.StudentID,
		MentorID:    in.MentorID,
		PackageID:   in.PackageID,
		SuggestedBy: &curatorID,
	}, func(r Request) []notifications.Notification {
		return []notifications.Notification{
			notice(r.MentorID, "Куратор предложил вам ученика", nil, r.ID),
			notice(r.StudentID, "Куратор подобрал вам наставника", nil, r.ID),
		}
	})
	if err != nil {
		return View{}, err
	}
	return s.repo.GetView(ctx, req.ID)
}

// Group — вкладки «Активные» и «Архив» на экране заявок.
type Group string

const (
	GroupAll     Group = ""
	GroupActive  Group = "active"
	GroupArchive Group = "archive"
)

// ListMine — для ученика отправленные заявки, для наставника адресованные ему.
func (s *Service) ListMine(ctx context.Context, actor Actor, group Group, limit, offset int) ([]View, int, error) {
	f := ListFilter{Limit: limit, Offset: offset}

	switch actor.Role {
	case user.Student:
		f.StudentID = &actor.UserID
	case user.Mentor:
		f.MentorID = &actor.UserID
	default:
		return nil, 0, ErrForbidden
	}

	switch group {
	case GroupActive:
		f.Statuses = ActiveStatuses
	case GroupArchive:
		f.Statuses = ArchiveStatuses
	}

	views, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if err := s.attachMatch(ctx, actor, views); err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

// ListAll — очередь заявок куратора с необязательным фильтром по статусу.
func (s *Service) ListAll(ctx context.Context, actor Actor, status *Status, limit, offset int) ([]View, int, error) {
	if actor.Role != user.Curator {
		return nil, 0, ErrForbidden
	}

	f := ListFilter{Limit: limit, Offset: offset}
	if status != nil {
		if !status.Valid() {
			return nil, 0, ErrInvalidStatus
		}
		f.Statuses = []Status{*status}
	}
	return s.repo.List(ctx, f)
}

type ChangeStatusInput struct {
	RequestID       uuid.UUID
	To              Status
	RejectionReason string
}

func (s *Service) ChangeStatus(ctx context.Context, actor Actor, in ChangeStatusInput) (View, error) {
	if !in.To.Valid() {
		return View{}, ErrInvalidStatus
	}

	req, err := s.repo.GetByID(ctx, in.RequestID)
	if err != nil {
		return View{}, err
	}
	// Чужая заявка выглядит как несуществующая.
	if !canAccess(actor, req) {
		return View{}, ErrNotFound
	}

	if err := CheckTransition(req.Status, in.To, actor.Role); err != nil {
		return View{}, err
	}

	var reason *string
	if r := strings.TrimSpace(in.RejectionReason); r != "" {
		reason = &r
	}

	switch in.To {
	case StatusRejectedByCurator:
		if reason == nil {
			return View{}, ErrReasonRequired
		}
	case StatusRejectedByMentor:
		// причина необязательна
	default:
		reason = nil
	}

	notify := transitionNotices(req, in.To, reason)

	if in.To == StatusAccepted {
		if _, _, err := s.repo.Accept(ctx, req.ID, notify); err != nil {
			return View{}, err
		}
	} else {
		if _, err := s.repo.UpdateStatus(ctx, req.ID, req.Status, in.To, reason, notify); err != nil {
			return View{}, err
		}
	}

	view, err := s.repo.GetView(ctx, req.ID)
	if err != nil {
		return View{}, err
	}
	views := []View{view}
	if err := s.attachMatch(ctx, actor, views); err != nil {
		return View{}, err
	}
	return views[0], nil
}

// transitionNotices — кому сообщить о смене статуса: второй стороне,
// а при решении куратора — обеим.
func transitionNotices(req Request, to Status, reason *string) []notifications.Notification {
	switch to {
	case StatusCancelled:
		return []notifications.Notification{notice(req.MentorID, "Ученик отозвал заявку", nil, req.ID)}
	case StatusApprovedByMentor:
		return []notifications.Notification{notice(req.StudentID, "Наставник одобрил заявку — подтвердите её", nil, req.ID)}
	case StatusRejectedByMentor:
		return []notifications.Notification{notice(req.StudentID, "Наставник отклонил заявку", reason, req.ID)}
	case StatusRejectedByCurator:
		return []notifications.Notification{
			notice(req.StudentID, "Заявка отклонена куратором", reason, req.ID),
			notice(req.MentorID, "Заявка отклонена куратором", reason, req.ID),
		}
	case StatusAccepted:
		return []notifications.Notification{notice(req.MentorID, "Ученик подтвердил заявку — пара создана", nil, req.ID)}
	default:
		return nil
	}
}

func notice(userID uuid.UUID, title string, message *string, requestID uuid.UUID) notifications.Notification {
	return notifications.Notification{
		UserID:   userID,
		Type:     notifications.TypeRequestUpdate,
		Title:    title,
		Message:  message,
		EntityID: &requestID,
	}
}

func normalizeCoverLetter(s string) (*string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(s) > MaxCoverLetterLen {
		return nil, ErrCoverLetterLength
	}
	return &s, nil
}

func canAccess(actor Actor, req Request) bool {
	switch actor.Role {
	case user.Curator:
		return true
	case user.Student:
		return req.StudentID == actor.UserID
	case user.Mentor:
		return req.MentorID == actor.UserID
	default:
		return false
	}
}
