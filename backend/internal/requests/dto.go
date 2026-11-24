package requests

import (
	"time"

	"github.com/google/uuid"
)

type createRequestBody struct {
	MentorID         uuid.UUID  `json:"mentor_id"`
	ServicePackageID *uuid.UUID `json:"service_package_id"`
	CoverLetter      string     `json:"cover_letter"`
}

type curatorCreateBody struct {
	StudentID        uuid.UUID  `json:"student_id"`
	MentorID         uuid.UUID  `json:"mentor_id"`
	ServicePackageID *uuid.UUID `json:"service_package_id"`
}

type statusUpdateBody struct {
	Status          Status `json:"status"`
	RejectionReason string `json:"rejection_reason"`
}

type studentPublicSummaryJSON struct {
	ID                 uuid.UUID `json:"id"`
	FirstName          string    `json:"first_name"`
	TargetSkills       []string  `json:"target_skills"`
	MainGoals          []string  `json:"main_goals"`
	TimeLimit          *int      `json:"time_limit,omitempty"`
	MinDurationMonths  *int      `json:"min_duration_months,omitempty"`
	MaxDurationMonths  *int      `json:"max_duration_months,omitempty"`
	CommunicationStyle *string   `json:"communication_style,omitempty"`
}

type personRefJSON struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url"`
}

// requestJSON — RequestResponse из контракта.
type requestJSON struct {
	ID                  uuid.UUID                `json:"id"`
	StudentID           uuid.UUID                `json:"student_id"`
	MentorID            uuid.UUID                `json:"mentor_id"`
	Student             studentPublicSummaryJSON `json:"student"`
	Mentor              personRefJSON            `json:"mentor"`
	ServicePackageID    *uuid.UUID               `json:"service_package_id,omitempty"`
	ServicePackageTitle *string                  `json:"service_package_title,omitempty"`
	Status              Status                   `json:"status"`
	CoverLetter         *string                  `json:"cover_letter,omitempty"`
	RejectionReason     *string                  `json:"rejection_reason,omitempty"`
	WorkspaceID         *uuid.UUID               `json:"workspace_id,omitempty"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

// curatorRequestJSON — CuratorRequestResponse: куратор видит ученика полностью.
type curatorRequestJSON struct {
	requestJSON
	StudentFull personRefJSON `json:"student_full"`
}

type pageJSON[T any] struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Items []T `json:"items"`
}

func toRequestJSON(v View) requestJSON {
	return requestJSON{
		ID:        v.ID,
		StudentID: v.StudentID,
		MentorID:  v.MentorID,
		Student: studentPublicSummaryJSON{
			ID:                 v.Student.ID,
			FirstName:          v.Student.FirstName,
			TargetSkills:       nonNil(v.Student.TargetSkills),
			MainGoals:          nonNil(v.Student.MainGoals),
			TimeLimit:          v.Student.TimeLimit,
			MinDurationMonths:  v.Student.MinDurationMonths,
			MaxDurationMonths:  v.Student.MaxDurationMonths,
			CommunicationStyle: v.Student.CommunicationStyle,
		},
		Mentor:              toPersonJSON(v.Mentor),
		ServicePackageID:    v.ServicePackageID,
		ServicePackageTitle: v.ServicePackageTitle,
		Status:              v.Status,
		CoverLetter:         v.CoverLetter,
		RejectionReason:     v.RejectionReason,
		WorkspaceID:         v.WorkspaceID,
		CreatedAt:           v.CreatedAt,
		UpdatedAt:           v.UpdatedAt,
	}
}

func toCuratorRequestJSON(v View) curatorRequestJSON {
	return curatorRequestJSON{requestJSON: toRequestJSON(v), StudentFull: toPersonJSON(v.StudentFull)}
}

func toPersonJSON(p Person) personRefJSON {
	return personRefJSON(p)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
