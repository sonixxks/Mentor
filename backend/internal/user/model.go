package user

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	Student Role = "student"
	Mentor  Role = "mentor"
	Curator Role = "curator"
)

func (r Role) Valid() bool {
	if r != Student && r != Mentor && r != Curator {
		return false
	}
	return true
}

func (r Role) Registrable() bool {
	if r != Student && r != Mentor {
		return false
	}
	return true
}

type User struct {
	UserID    uuid.UUID
	Role      Role
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRecord struct {
	User User
	Hash string
}
