package requests

import "time"

type Status string

const (
	StatusPending           Status = "pending"
	StatusAccepted          Status = "accepted"
	StatusRejectedByMentor  Status = "rejected_by_mentor"
	StatusRejectedByCurator Status = "rejected_by_curator"
	StatusCancelled         Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusAccepted, StatusRejectedByMentor, StatusRejectedByCurator, StatusCancelled:
		return true
	}
	return false
}

// Role совпадает с ENUM user_role в БД.
// TODO: перенести в общий пакет, когда появится auth middleware.
type Role string

const (
	RoleStudent Role = "student"
	RoleMentor  Role = "mentor"
	RoleCurator Role = "curator"
)

// Actor — тот, кто выполняет действие (берётся из JWT).
type Actor struct {
	UserID string
	Role   Role
}

type Request struct {
	ID               string
	StudentID        string
	MentorID         string
	ServicePackageID *string
	Status           Status
	CuratorComment   *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Workspace struct {
	ID        string
	RequestID string
	JitsiLink string
	CreatedAt time.Time
}
