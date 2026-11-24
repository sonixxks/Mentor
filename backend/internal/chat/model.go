// Package chat — текстовый чат внутри пары (ученик и наставник).
// Куратор переписку не читает. Доставка — через запросы истории (polling).
package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrNotMember         = errors.New("only members of the pair can use the chat")
	ErrEmptyMessage      = errors.New("message must contain text or file_id")
	ErrTextTooLong       = errors.New("message text is too long")
	ErrFileNotFound      = errors.New("file not found in this workspace")
)

const (
	MaxTextLen       = 4000
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

type MessageType string

const (
	TypeText   MessageType = "text"
	TypeSystem MessageType = "system"
)

type File struct {
	ID         uuid.UUID
	Type       string
	FileName   string
	FileURL    string
	FileSizeKB *int
	UploadedBy uuid.UUID
	CreatedAt  time.Time
}

type Message struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	SenderID    *uuid.UUID
	Type        MessageType
	Text        *string
	File        *File
	IsRead      bool
	CreatedAt   time.Time
}

// Members — участники пары.
type Members struct {
	StudentID uuid.UUID
	MentorID  uuid.UUID
}

func (m Members) Has(userID uuid.UUID) bool {
	return userID == m.StudentID || userID == m.MentorID
}

func (m Members) PartnerOf(userID uuid.UUID) uuid.UUID {
	if userID == m.StudentID {
		return m.MentorID
	}
	return m.StudentID
}
