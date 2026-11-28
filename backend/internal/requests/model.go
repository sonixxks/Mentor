package requests

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/matching"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Status string

const (
	StatusPending           Status = "pending"
	StatusApprovedByMentor  Status = "approved_by_mentor"
	StatusAccepted          Status = "accepted"
	StatusRejectedByMentor  Status = "rejected_by_mentor"
	StatusRejectedByCurator Status = "rejected_by_curator"
	StatusCancelled         Status = "cancelled"
)

// ActiveStatuses — заявка ещё в работе: занимает слот в лимите ученика.
var ActiveStatuses = []Status{StatusPending, StatusApprovedByMentor}

// ArchiveStatuses — финальные статусы.
var ArchiveStatuses = []Status{StatusAccepted, StatusRejectedByMentor, StatusRejectedByCurator, StatusCancelled}

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApprovedByMentor, StatusAccepted,
		StatusRejectedByMentor, StatusRejectedByCurator, StatusCancelled:
		return true
	}
	return false
}

// MaxActivePerStudent — сколько активных заявок может отправить ученик сам.
const MaxActivePerStudent = 3

// MaxCoverLetterLen — длина сопроводительного письма в символах.
const MaxCoverLetterLen = 1000

// Actor — пользователь, выполняющий действие (из JWT).
type Actor struct {
	UserID uuid.UUID
	Role   user.Role
}

type Request struct {
	ID               uuid.UUID
	StudentID        uuid.UUID
	MentorID         uuid.UUID
	ServicePackageID *uuid.UUID
	Status           Status
	CoverLetter      *string
	RejectionReason  *string
	SuggestedBy      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// StudentSummary — то, что видит наставник об ученике до создания пары.
type StudentSummary struct {
	ID                 uuid.UUID
	FirstName          string
	TargetSkills       []string
	MainGoals          []string
	TimeLimit          *int
	MinDurationMonths  *int
	MaxDurationMonths  *int
	CommunicationStyle *string
}

type Person struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	AvatarURL *string
}

// MatchInfo — совместимость ученика с наставником.
type MatchInfo = matching.Info

// View — заявка с данными для экранов: участники, программа, воркспейс.
type View struct {
	Request
	Student             StudentSummary
	StudentFull         Person
	Mentor              Person
	ServicePackageTitle *string
	WorkspaceID         *uuid.UUID
	// Match — только в ответах наставнику.
	Match *MatchInfo
}

// ListFilter — выборка заявок. Пустые поля не ограничивают выдачу.
type ListFilter struct {
	StudentID *uuid.UUID
	MentorID  *uuid.UUID
	Statuses  []Status
	Limit     int
	Offset    int
}

type CreateParams struct {
	StudentID   uuid.UUID
	MentorID    uuid.UUID
	PackageID   *uuid.UUID
	CoverLetter *string
	// SuggestedBy — куратор, если заявка из ручного подбора.
	SuggestedBy *uuid.UUID
}
