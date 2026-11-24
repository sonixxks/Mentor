package catalog

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type mentorSummaryJSON struct {
	ID                 uuid.UUID  `json:"id"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	AvatarURL          *string    `json:"avatar_url"`
	JobTitle           *string    `json:"job_title,omitempty"`
	Company            *string    `json:"company"`
	Experience         *string    `json:"experience,omitempty"`
	Bio                *string    `json:"bio,omitempty"`
	HourlyRate         *float64   `json:"hourly_rate,omitempty"`
	VerificationStatus string     `json:"verification_status"`
	CurrentSkills      []Skill    `json:"current_skills"`
	RatingAvg          *float64   `json:"rating_avg"`
	ReviewsCount       int        `json:"reviews_count"`
	FreeSlots          int        `json:"free_slots"`
	TimeLimit          *int       `json:"time_limit,omitempty"`
	CommunicationStyle *string    `json:"communication_style,omitempty"`
	Match              *MatchInfo `json:"match,omitempty"`
}

type servicePackageJSON struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	Result       *string   `json:"result,omitempty"`
	Price        *float64  `json:"price,omitempty"`
	SessionCount *int      `json:"session_count,omitempty"`
	Skills       []string  `json:"skills"`
	Status       string    `json:"status"`
	PairsCount   int       `json:"pairs_count"`
}

type mentorDetailJSON struct {
	mentorSummaryJSON
	HelpTopics          []string             `json:"help_topics"`
	TargetStudentLevels []string             `json:"target_student_levels"`
	AvailabilitySlots   json.RawMessage      `json:"availability_slots,omitempty"`
	MaxStudents         *int                 `json:"max_students,omitempty"`
	ServicePackages     []servicePackageJSON `json:"service_packages"`
}

type reviewJSON struct {
	ID              uuid.UUID  `json:"id"`
	Rating          int        `json:"rating"`
	Comment         *string    `json:"comment,omitempty"`
	AuthorFirstName string     `json:"author_first_name"`
	MeetingDate     *time.Time `json:"meeting_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type pageJSON[T any] struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Items []T `json:"items"`
}

func toSummaryJSON(m MentorSummary) mentorSummaryJSON {
	return mentorSummaryJSON{
		ID:                 m.ID,
		FirstName:          m.FirstName,
		LastName:           m.LastName,
		AvatarURL:          m.AvatarURL,
		JobTitle:           m.JobTitle,
		Company:            m.Company,
		Experience:         m.Experience,
		Bio:                m.Bio,
		HourlyRate:         m.HourlyRate,
		VerificationStatus: m.VerificationStatus,
		CurrentSkills:      nonNil(m.CurrentSkills),
		RatingAvg:          m.RatingAvg,
		ReviewsCount:       m.ReviewsCount,
		FreeSlots:          m.FreeSlots,
		TimeLimit:          m.TimeLimit,
		CommunicationStyle: m.CommunicationStyle,
		Match:              m.Match,
	}
}

func toDetailJSON(d MentorDetail) mentorDetailJSON {
	packages := make([]servicePackageJSON, len(d.ServicePackages))
	for i, p := range d.ServicePackages {
		packages[i] = servicePackageJSON{
			ID:           p.ID,
			Title:        p.Title,
			Description:  p.Description,
			Result:       p.Result,
			Price:        p.Price,
			SessionCount: p.SessionCount,
			Skills:       nonNil(p.Skills),
			Status:       p.Status,
			PairsCount:   p.PairsCount,
		}
	}
	return mentorDetailJSON{
		mentorSummaryJSON:   toSummaryJSON(d.MentorSummary),
		HelpTopics:          nonNil(d.HelpTopics),
		TargetStudentLevels: nonNil(d.TargetStudentLevels),
		AvailabilitySlots:   d.AvailabilitySlots,
		MaxStudents:         d.MaxStudents,
		ServicePackages:     packages,
	}
}

func toReviewJSON(r Review) reviewJSON {
	return reviewJSON(r)
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
