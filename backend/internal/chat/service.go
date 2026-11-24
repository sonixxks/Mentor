package chat

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Repository interface {
	Members(ctx context.Context, workspaceID uuid.UUID) (Members, error)
	List(ctx context.Context, workspaceID uuid.UUID, before *time.Time, limit int) ([]Message, error)
	Create(ctx context.Context, workspaceID, senderID uuid.UUID, text *string, fileID *uuid.UUID,
		notify func(Message) notifications.Notification) (Message, error)
	MarkRead(ctx context.Context, workspaceID, readerID uuid.UUID) error
}

type Actor struct {
	UserID uuid.UUID
	Role   user.Role
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// History — сообщения старше before (или последние), по возрастанию времени.
func (s *Service) History(ctx context.Context, actor Actor, workspaceID uuid.UUID, before *time.Time, limit int) ([]Message, error) {
	if _, err := s.members(ctx, actor, workspaceID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > MaxPageLimit {
		limit = DefaultPageLimit
	}
	return s.repo.List(ctx, workspaceID, before, limit)
}

type SendInput struct {
	Text   string
	FileID *uuid.UUID
}

func (s *Service) Send(ctx context.Context, actor Actor, workspaceID uuid.UUID, in SendInput) (Message, error) {
	text := strings.TrimSpace(in.Text)
	if text == "" && in.FileID == nil {
		return Message{}, ErrEmptyMessage
	}
	if utf8.RuneCountInString(text) > MaxTextLen {
		return Message{}, ErrTextTooLong
	}

	members, err := s.members(ctx, actor, workspaceID)
	if err != nil {
		return Message{}, err
	}

	var textPtr *string
	if text != "" {
		textPtr = &text
	}

	partner := members.PartnerOf(actor.UserID)
	return s.repo.Create(ctx, workspaceID, actor.UserID, textPtr, in.FileID, func(m Message) notifications.Notification {
		return notifications.Notification{
			UserID:   partner,
			Type:     notifications.TypeNewMessage,
			Title:    "Новое сообщение в чате пары",
			Message:  preview(m),
			EntityID: &m.ID,
		}
	})
}

func (s *Service) MarkRead(ctx context.Context, actor Actor, workspaceID uuid.UUID) error {
	if _, err := s.members(ctx, actor, workspaceID); err != nil {
		return err
	}
	return s.repo.MarkRead(ctx, workspaceID, actor.UserID)
}

// members проверяет, что пользователь — участник пары. Куратор переписку не читает.
func (s *Service) members(ctx context.Context, actor Actor, workspaceID uuid.UUID) (Members, error) {
	if actor.Role != user.Student && actor.Role != user.Mentor {
		return Members{}, ErrNotMember
	}
	m, err := s.repo.Members(ctx, workspaceID)
	if err != nil {
		return Members{}, err
	}
	if !m.Has(actor.UserID) {
		return Members{}, ErrNotMember
	}
	return m, nil
}

const previewLen = 100

func preview(m Message) *string {
	var s string
	switch {
	case m.Text != nil:
		s = *m.Text
		if utf8.RuneCountInString(s) > previewLen {
			s = string([]rune(s)[:previewLen]) + "…"
		}
	case m.File != nil:
		s = "Файл: " + m.File.FileName
	default:
		return nil
	}
	return &s
}
