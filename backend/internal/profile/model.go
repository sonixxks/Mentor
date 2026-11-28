// Package profile — анкета пользователя: создание при регистрации, чтение и изменение
// по шагам (/profiles/me), материалы для верификации наставника.
package profile

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

var (
	ErrProfileNotFound = errors.New("профиль не найден")
	ErrMentorOnly      = errors.New("доступно только наставнику")
	ErrAssetNotFound   = errors.New("материал не найден")
)

// ValidationError — ошибки по полям, уходят клиенту в details.
type ValidationError struct {
	Details []string
}

func (e *ValidationError) Error() string {
	return "проверьте правильность заполнения полей"
}

type Skill struct {
	Skill string `json:"skill"`
	Level int    `json:"level"`
}

type Channels struct {
	Site  bool `json:"site"`
	Email bool `json:"email"`
}

type Asset struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}

// Profile — анкета целиком. Поля JSONB хранятся в типизированном виде.
type Profile struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Role   user.Role

	FirstName  string
	LastName   string
	MiddleName *string
	AvatarURL  *string

	Age           *int
	CurrentStatus *string
	University    *string
	Major         *string
	Course        *int
	Experience    *string

	JobTitle    *string
	Company     *string
	Bio         *string
	HourlyRate  *float64
	MaxStudents *int
	IsPublished bool

	CommunicationStyle *string
	AvailabilitySlots  map[string][]int
	MainGoals          []string
	SuccessMetric      *string
	TimeLimit          *int
	NeedsCuratorHelp   bool
	MinDurationMonths  *int
	MaxDurationMonths  *int

	CurrentSkills           []Skill
	TargetSkills            []string
	HelpTopics              []string
	TargetStudentLevels     []string
	NotificationPreferences map[string]Channels

	OnboardingCompleted bool
	VerificationStatus  string
	VerificationAssets  []Asset

	// FreeSlots — только для наставника: max_students минус пары в работе.
	FreeSlots *int
}

// Справочники из контракта.
var (
	currentStatuses     = set("student", "graduate", "junior", "switcher")
	studentExperience   = set("none", "under_1", "1_to_3", "over_3")
	mentorExperience    = set("1_to_3", "3_to_5", "5_to_10", "over_10")
	communicationStyles = set("friendly", "structured", "free")
	mainGoals           = set("find_job_or_internship", "learn_new_skill", "build_project", "interview_prep", "change_direction", "develop_soft_skills")
	helpTopics          = set("interview_prep", "code_review", "portfolio", "career_growth", "change_direction", "soft_skills", "first_job_in_it")
	studentLevels       = set("beginner", "has_basics", "advanced")
	weekDays            = set("mon", "tue", "wed", "thu", "fri", "sat", "sun")
	notificationKinds   = set("requests", "tasks", "messages", "recommendations")
)

func set(values ...string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}

// Лимиты материалов для верификации.
const (
	MaxAssetsTotal   = 10
	MaxFilesPerCall  = 5
	MaxLinksPerCall  = 5
	MaxAssetFileSize = 10 << 20
)
