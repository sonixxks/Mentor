// Package curator — API панели куратора: «светофор» пар, ручная смена статуса пары,
// сообщения паре, очередь запросов помощи и верификации наставников.
package curator

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkspaceNotFound = errors.New("пара не найдена")
	ErrProfileNotFound   = errors.New("профиль не найден")
	ErrInvalidStatus     = errors.New("неизвестный статус")
	ErrMentorLimit       = errors.New("у наставника нет свободных мест")
	ErrNoHelpRequest     = errors.New("у пользователя нет открытого запроса помощи")
	ErrNotMentor         = errors.New("профиль не принадлежит наставнику")
	ErrNotPending        = errors.New("профиль не ожидает проверки")
	ErrCommentRequired   = errors.New("укажите комментарий")
	ErrCommentTooLong    = errors.New("комментарий длиннее 1000 символов")
	ErrInvalidTarget     = errors.New("получатель должен быть student, mentor или both")
)

const MaxCommentLen = 1000

type WorkspaceStatus string

const (
	WorkspaceActive     WorkspaceStatus = "active"
	WorkspacePaused     WorkspaceStatus = "paused"
	WorkspaceStagnation WorkspaceStatus = "stagnation"
	WorkspaceCompleted  WorkspaceStatus = "completed"
)

func (s WorkspaceStatus) Valid() bool {
	switch s {
	case WorkspaceActive, WorkspacePaused, WorkspaceStagnation, WorkspaceCompleted:
		return true
	}
	return false
}

// OccupiesSlot — пара в этом статусе занимает место наставника.
func (s WorkspaceStatus) OccupiesSlot() bool {
	return s != WorkspaceCompleted
}

// WorkspaceStats — сырые показатели пары из БД; сигналы считаются в Evaluate.
type WorkspaceStats struct {
	ID             uuid.UUID
	Status         WorkspaceStatus
	StudentID      uuid.UUID
	MentorID       uuid.UUID
	StudentName    string
	MentorName     string
	StudentAvatar  *string
	StartedAt      time.Time
	LastActivityAt *time.Time
	TasksTotal     int
	TasksCompleted int
	TasksOverdue   int
	LastRating     *int
}

type SignalType string

const (
	SignalNoActivity   SignalType = "no_activity"
	SignalOverdueTasks SignalType = "overdue_tasks"
	SignalBadReviews   SignalType = "bad_reviews"
)

type Signal struct {
	Type  SignalType `json:"type"`
	Value int        `json:"value"`
}

type Severity string

const (
	SeverityGreen  Severity = "green"
	SeverityYellow Severity = "yellow"
	SeverityRed    Severity = "red"
)

// WorkspaceRow — строка «светофора».
type WorkspaceRow struct {
	WorkspaceStats
	Severity        Severity
	Signals         []Signal
	ProgressPercent int
}

type StatusCounts struct {
	Total      int `json:"total"`
	Active     int `json:"active"`
	Paused     int `json:"paused"`
	Stagnation int `json:"stagnation"`
}

type WorkspaceSort string

const (
	SortSeverity WorkspaceSort = "severity"
	SortActivity WorkspaceSort = "activity"
)

type WorkspaceFilter struct {
	Query      string
	Sort       WorkspaceSort
	HasSignals *bool
	Limit      int
	Offset     int
}

type NotifyTarget string

const (
	TargetStudent NotifyTarget = "student"
	TargetMentor  NotifyTarget = "mentor"
	TargetBoth    NotifyTarget = "both"
)

// Profile — анкета пользователя целиком (CuratorProfileResponse).
// JSONB-поля отдаются как есть.
type Profile struct {
	ID                      uuid.UUID
	UserID                  uuid.UUID
	Role                    string
	Email                   string
	FirstName               string
	LastName                string
	MiddleName              *string
	AvatarURL               *string
	Age                     *int
	CurrentStatus           *string
	University              *string
	Major                   *string
	Course                  *int
	Experience              *string
	JobTitle                *string
	Company                 *string
	Bio                     *string
	HourlyRate              *float64
	MaxStudents             *int
	IsPublished             bool
	CommunicationStyle      *string
	AvailabilitySlots       json.RawMessage
	MainGoals               json.RawMessage
	SuccessMetric           *string
	TimeLimit               *int
	NeedsCuratorHelp        bool
	MinDurationMonths       *int
	MaxDurationMonths       *int
	CurrentSkills           json.RawMessage
	TargetSkills            json.RawMessage
	HelpTopics              json.RawMessage
	TargetStudentLevels     json.RawMessage
	NotificationPreferences json.RawMessage
	OnboardingCompleted     bool
	VerificationStatus      string
	VerificationAssets      json.RawMessage
	FreeSlots               *int
}

type ProfileFilter struct {
	NeedsHelp          *bool
	VerificationStatus *string
	Limit              int
	Offset             int
}
