package profile

import "github.com/google/uuid"

type Profile struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	FirstName  string
	LastName   string
	MiddleName *string
}
