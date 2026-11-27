package curator

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
)

type Repository interface {
	ListWorkspaceStats(ctx context.Context, query string) ([]WorkspaceStats, error)
	GetWorkspaceStats(ctx context.Context, id uuid.UUID) (WorkspaceStats, error)
	CountWorkspaces(ctx context.Context) (StatusCounts, error)
	SetWorkspaceStatus(ctx context.Context, id uuid.UUID, to WorkspaceStatus,
		notify func(studentID, mentorID uuid.UUID) []notifications.Notification) error
	WorkspaceMembers(ctx context.Context, id uuid.UUID) (studentID, mentorID uuid.UUID, err error)
	Notify(ctx context.Context, list []notifications.Notification) error

	ListProfiles(ctx context.Context, f ProfileFilter) ([]Profile, int, error)
	GetProfile(ctx context.Context, id uuid.UUID) (Profile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, check func(ProfileState) (ProfileChange, error)) error
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

type WorkspacePage struct {
	Items  []WorkspaceRow
	Total  int
	Counts StatusCounts
}

// ListWorkspaces — «светофор». Сигналы вычисляются на лету, поэтому фильтр
// has_signals, сортировка и пагинация делаются после расчёта.
func (s *Service) ListWorkspaces(ctx context.Context, f WorkspaceFilter) (WorkspacePage, error) {
	stats, err := s.repo.ListWorkspaceStats(ctx, f.Query)
	if err != nil {
		return WorkspacePage{}, err
	}
	counts, err := s.repo.CountWorkspaces(ctx)
	if err != nil {
		return WorkspacePage{}, err
	}

	now := s.now()
	rows := make([]WorkspaceRow, 0, len(stats))
	for _, st := range stats {
		row := Evaluate(st, now)
		if f.HasSignals != nil && (len(row.Signals) > 0) != *f.HasSignals {
			continue
		}
		rows = append(rows, row)
	}
	SortRows(rows, f.Sort, now)

	total := len(rows)
	start := min(f.Offset, total)
	end := total
	if f.Limit > 0 {
		end = min(start+f.Limit, total)
	}
	return WorkspacePage{Items: rows[start:end], Total: total, Counts: counts}, nil
}

var statusTitles = map[WorkspaceStatus]string{
	WorkspaceActive:     "Куратор возобновил работу пары",
	WorkspacePaused:     "Куратор поставил пару на паузу",
	WorkspaceStagnation: "Куратор отметил пару как стагнирующую",
	WorkspaceCompleted:  "Куратор завершил работу пары",
}

// SetWorkspaceStatus — ручная смена статуса пары, обе стороны получают уведомление.
func (s *Service) SetWorkspaceStatus(ctx context.Context, id uuid.UUID, to WorkspaceStatus) (WorkspaceRow, error) {
	if !to.Valid() {
		return WorkspaceRow{}, ErrInvalidStatus
	}

	err := s.repo.SetWorkspaceStatus(ctx, id, to, func(studentID, mentorID uuid.UUID) []notifications.Notification {
		return alerts(statusTitles[to], nil, id, studentID, mentorID)
	})
	if err != nil {
		return WorkspaceRow{}, err
	}

	stats, err := s.repo.GetWorkspaceStats(ctx, id)
	if err != nil {
		return WorkspaceRow{}, err
	}
	return Evaluate(stats, s.now()), nil
}

// NotifyWorkspace — «Написать паре»: уведомление system_alert выбранной стороне.
func (s *Service) NotifyWorkspace(ctx context.Context, id uuid.UUID, target NotifyTarget, message string) error {
	msg, err := normalizeComment(message, true)
	if err != nil {
		return err
	}

	studentID, mentorID, err := s.repo.WorkspaceMembers(ctx, id)
	if err != nil {
		return err
	}

	var recipients []uuid.UUID
	switch target {
	case TargetStudent:
		recipients = []uuid.UUID{studentID}
	case TargetMentor:
		recipients = []uuid.UUID{mentorID}
	case TargetBoth:
		recipients = []uuid.UUID{studentID, mentorID}
	default:
		return ErrInvalidTarget
	}

	return s.repo.Notify(ctx, alerts("Сообщение от куратора", msg, id, recipients...))
}

func (s *Service) ListProfiles(ctx context.Context, f ProfileFilter) ([]Profile, int, error) {
	return s.repo.ListProfiles(ctx, f)
}

// ResolveHelpRequest закрывает запрос помощи и отправляет комментарий пользователю.
func (s *Service) ResolveHelpRequest(ctx context.Context, profileID uuid.UUID, comment string) (Profile, error) {
	msg, err := normalizeComment(comment, false)
	if err != nil {
		return Profile{}, err
	}

	err = s.repo.UpdateProfile(ctx, profileID, func(st ProfileState) (ProfileChange, error) {
		if !st.NeedsCuratorHelp {
			return ProfileChange{}, ErrNoHelpRequest
		}
		return ProfileChange{
			ResolveHelp: true,
			Notify:      alerts("Куратор обработал ваш запрос помощи", msg, profileID, st.UserID),
		}, nil
	})
	if err != nil {
		return Profile{}, err
	}
	return s.repo.GetProfile(ctx, profileID)
}

type VerificationDecision string

const (
	VerificationVerified VerificationDecision = "verified"
	VerificationRejected VerificationDecision = "rejected"
)

// SetVerification — итог проверки документов наставника (только из pending).
// Комментарий обязателен при отказе и уходит наставнику уведомлением.
func (s *Service) SetVerification(ctx context.Context, profileID uuid.UUID, decision VerificationDecision, comment string) (Profile, error) {
	if decision != VerificationVerified && decision != VerificationRejected {
		return Profile{}, ErrInvalidStatus
	}
	msg, err := normalizeComment(comment, decision == VerificationRejected)
	if err != nil {
		return Profile{}, err
	}

	title := "Профиль прошёл верификацию и виден в каталоге"
	if decision == VerificationRejected {
		title = "Верификация профиля отклонена"
	}

	err = s.repo.UpdateProfile(ctx, profileID, func(st ProfileState) (ProfileChange, error) {
		if err := checkPendingMentor(st); err != nil {
			return ProfileChange{}, err
		}
		status := string(decision)
		return ProfileChange{
			Verification: &status,
			Notify:       alerts(title, msg, profileID, st.UserID),
		}, nil
	})
	if err != nil {
		return Profile{}, err
	}
	return s.repo.GetProfile(ctx, profileID)
}

// RequestDocuments просит наставника дополнить материалы; статус остаётся pending.
func (s *Service) RequestDocuments(ctx context.Context, profileID uuid.UUID, comment string) error {
	msg, err := normalizeComment(comment, true)
	if err != nil {
		return err
	}

	return s.repo.UpdateProfile(ctx, profileID, func(st ProfileState) (ProfileChange, error) {
		if err := checkPendingMentor(st); err != nil {
			return ProfileChange{}, err
		}
		return ProfileChange{
			Notify: alerts("Куратор просит дополнить документы для верификации", msg, profileID, st.UserID),
		}, nil
	})
}

func checkPendingMentor(st ProfileState) error {
	if st.Role != "mentor" {
		return ErrNotMentor
	}
	if st.VerificationStatus != "pending" {
		return ErrNotPending
	}
	return nil
}

func alerts(title string, message *string, entityID uuid.UUID, users ...uuid.UUID) []notifications.Notification {
	list := make([]notifications.Notification, len(users))
	for i, u := range users {
		list[i] = notifications.Notification{
			UserID:   u,
			Type:     notifications.TypeSystemAlert,
			Title:    title,
			Message:  message,
			EntityID: &entityID,
		}
	}
	return list
}

func normalizeComment(s string, required bool) (*string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		if required {
			return nil, ErrCommentRequired
		}
		return nil, nil
	}
	if utf8.RuneCountInString(s) > MaxCommentLen {
		return nil, ErrCommentTooLong
	}
	return &s, nil
}
