package curator

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type workspaceRowJSON struct {
	ID              uuid.UUID       `json:"id"`
	Status          WorkspaceStatus `json:"status"`
	Severity        Severity        `json:"severity"`
	StudentName     string          `json:"student_name"`
	MentorName      string          `json:"mentor_name"`
	ProgressPercent int             `json:"progress_percent"`
	StartedAt       time.Time       `json:"started_at"`
	LastActivityAt  *time.Time      `json:"last_activity_at"`
	Signals         []Signal        `json:"signals"`
}

type workspacePageJSON struct {
	Total  int                `json:"total"`
	Page   int                `json:"page"`
	Items  []workspaceRowJSON `json:"items"`
	Counts StatusCounts       `json:"counts"`
}

type personRefJSON struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url"`
}

// workspaceSummaryJSON — WorkspaceSummary после смены статуса.
// Для куратора partner — ученик пары.
type workspaceSummaryJSON struct {
	ID              uuid.UUID       `json:"id"`
	Status          WorkspaceStatus `json:"status"`
	Partner         personRefJSON   `json:"partner"`
	ProgressPercent int             `json:"progress_percent"`
	StartedAt       time.Time       `json:"started_at"`
	InactiveDays    *int            `json:"inactive_days"`
}

type profileJSON struct {
	ID                      uuid.UUID       `json:"id"`
	UserID                  uuid.UUID       `json:"user_id"`
	Role                    string          `json:"role"`
	Email                   string          `json:"email"`
	FirstName               string          `json:"first_name"`
	LastName                string          `json:"last_name"`
	MiddleName              *string         `json:"middle_name,omitempty"`
	AvatarURL               *string         `json:"avatar_url,omitempty"`
	Age                     *int            `json:"age,omitempty"`
	CurrentStatus           *string         `json:"current_status,omitempty"`
	University              *string         `json:"university,omitempty"`
	Major                   *string         `json:"major,omitempty"`
	Course                  *int            `json:"course,omitempty"`
	Experience              *string         `json:"experience,omitempty"`
	JobTitle                *string         `json:"job_title,omitempty"`
	Company                 *string         `json:"company,omitempty"`
	Bio                     *string         `json:"bio,omitempty"`
	HourlyRate              *float64        `json:"hourly_rate,omitempty"`
	MaxStudents             *int            `json:"max_students,omitempty"`
	IsPublished             bool            `json:"is_published"`
	CommunicationStyle      *string         `json:"communication_style,omitempty"`
	AvailabilitySlots       json.RawMessage `json:"availability_slots,omitempty"`
	MainGoals               json.RawMessage `json:"main_goals,omitempty"`
	SuccessMetric           *string         `json:"success_metric,omitempty"`
	TimeLimit               *int            `json:"time_limit,omitempty"`
	NeedsCuratorHelp        bool            `json:"needs_curator_help"`
	MinDurationMonths       *int            `json:"min_duration_months,omitempty"`
	MaxDurationMonths       *int            `json:"max_duration_months,omitempty"`
	CurrentSkills           json.RawMessage `json:"current_skills,omitempty"`
	TargetSkills            json.RawMessage `json:"target_skills,omitempty"`
	HelpTopics              json.RawMessage `json:"help_topics,omitempty"`
	TargetStudentLevels     json.RawMessage `json:"target_student_levels,omitempty"`
	NotificationPreferences json.RawMessage `json:"notification_preferences,omitempty"`
	OnboardingCompleted     bool            `json:"onboarding_completed"`
	VerificationStatus      string          `json:"verification_status"`
	VerificationAssets      json.RawMessage `json:"verification_assets,omitempty"`
	FreeSlots               *int            `json:"free_slots,omitempty"`
}

type profilePageJSON struct {
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Items []profileJSON `json:"items"`
}

type workspaceStatusBody struct {
	Status WorkspaceStatus `json:"status"`
}

type notifyBody struct {
	Target  NotifyTarget `json:"target"`
	Message string       `json:"message"`
}

type commentBody struct {
	Comment string `json:"comment"`
}

type verificationBody struct {
	Status  VerificationDecision `json:"status"`
	Comment string               `json:"comment"`
}

func toWorkspaceRowJSON(r WorkspaceRow) workspaceRowJSON {
	return workspaceRowJSON{
		ID:              r.ID,
		Status:          r.Status,
		Severity:        r.Severity,
		StudentName:     r.StudentName,
		MentorName:      r.MentorName,
		ProgressPercent: r.ProgressPercent,
		StartedAt:       r.StartedAt,
		LastActivityAt:  r.LastActivityAt,
		Signals:         r.Signals,
	}
}

func toWorkspaceSummaryJSON(r WorkspaceRow, now time.Time) workspaceSummaryJSON {
	first, last, _ := cutName(r.StudentName)
	out := workspaceSummaryJSON{
		ID:              r.ID,
		Status:          r.Status,
		Partner:         personRefJSON{ID: r.StudentID, FirstName: first, LastName: last, AvatarURL: r.StudentAvatar},
		ProgressPercent: r.ProgressPercent,
		StartedAt:       r.StartedAt,
	}
	if days := InactiveDays(r.WorkspaceStats, now); days > 0 {
		out.InactiveDays = &days
	}
	return out
}

func toProfileJSON(p Profile) profileJSON {
	return profileJSON(p)
}

func cutName(full string) (first, last string, ok bool) {
	for i, r := range full {
		if r == ' ' {
			return full[:i], full[i+1:], true
		}
	}
	return full, "", false
}
