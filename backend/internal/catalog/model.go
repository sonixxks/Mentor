// Package catalog — публичный каталог наставников: ручной поиск с фильтрами,
// карточка наставника, отзывы и словарь навыков.
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMentorNotFound = errors.New("наставник не найден")
)

type Sort string

const (
	SortMatch      Sort = "match"
	SortRating     Sort = "rating"
	SortPriceAsc   Sort = "price_asc"
	SortPriceDesc  Sort = "price_desc"
	SortExperience Sort = "experience"
)

// Filter — параметры GET /mentors. nil-поля не ограничивают выдачу.
type Filter struct {
	Query        string
	Skills       []string
	PriceMin     *float64
	PriceMax     *float64
	HoursMin     *int
	HoursMax     *int
	MinMatch     *int
	MinRating    *float64
	HasFreeSlots bool
	Sort         Sort
	// Limit == 0 — без ограничения (нужно для сортировки по совместимости).
	Limit  int
	Offset int
}

type Skill struct {
	Skill string `json:"skill"`
	Level int    `json:"level"`
}

// MatchInfo — процент совместимости с анкетой ученика (считает модуль подбора).
type MatchInfo struct {
	Percent   int            `json:"percent"`
	Breakdown MatchBreakdown `json:"breakdown"`
}

type MatchBreakdown struct {
	Skills int `json:"skills"`
	Goals  int `json:"goals"`
	Time   int `json:"time"`
	Style  int `json:"style"`
}

// Matcher — расчёт совместимости ученика с наставниками (умный подбор).
// Возвращает ok = false, если анкета ученика не заполнена.
type Matcher interface {
	Match(ctx context.Context, studentID uuid.UUID, mentorIDs []uuid.UUID) (map[uuid.UUID]MatchInfo, bool, error)
}

// MentorSummary — карточка в списке каталога.
type MentorSummary struct {
	ID                 uuid.UUID
	FirstName          string
	LastName           string
	AvatarURL          *string
	JobTitle           *string
	Company            *string
	Experience         *string
	Bio                *string
	HourlyRate         *float64
	VerificationStatus string
	CurrentSkills      []Skill
	RatingAvg          *float64
	ReviewsCount       int
	FreeSlots          int
	TimeLimit          *int
	CommunicationStyle *string
	Match              *MatchInfo
}

type ServicePackage struct {
	ID           uuid.UUID
	Title        string
	Description  *string
	Result       *string
	Price        *float64
	SessionCount *int
	Skills       []string
	Status       string
	PairsCount   int
}

// MentorDetail — полная карточка наставника.
type MentorDetail struct {
	MentorSummary
	HelpTopics          []string
	TargetStudentLevels []string
	AvailabilitySlots   json.RawMessage
	MaxStudents         *int
	ServicePackages     []ServicePackage
	IsPublished         bool
}

type Review struct {
	ID              uuid.UUID
	Rating          int
	Comment         *string
	AuthorFirstName string
	MeetingDate     *time.Time
	CreatedAt       time.Time
}

type SkillQuery struct {
	Query   string
	Popular bool
	Limit   int
}
