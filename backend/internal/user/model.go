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
	Admin   Role = "admin"
)

func (r Role) Valid() bool {
	if r != Student && r != Mentor && r != Curator && r != Admin {
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
	ID                uuid.UUID
	Role              Role
	Email             string
	YandexID          *string
	ConsentAt         *time.Time
	PasswordChangedAt *time.Time
	CreatedAt         time.Time
}

type UserRecord struct {
	User User
	Hash *string
}
