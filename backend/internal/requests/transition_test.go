package requests

import (
	"errors"
	"testing"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

func TestCheckTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		role user.Role
		want error
	}{
		{"student cancels pending", StatusPending, StatusCancelled, user.Student, nil},
		{"mentor approves pending", StatusPending, StatusApprovedByMentor, user.Mentor, nil},
		{"mentor rejects pending", StatusPending, StatusRejectedByMentor, user.Mentor, nil},
		{"curator rejects pending", StatusPending, StatusRejectedByCurator, user.Curator, nil},
		{"student accepts approved", StatusApprovedByMentor, StatusAccepted, user.Student, nil},
		{"curator rejects approved", StatusApprovedByMentor, StatusRejectedByCurator, user.Curator, nil},

		{"mentor cannot accept pending", StatusPending, StatusAccepted, user.Mentor, ErrInvalidTransition},
		{"student cannot approve", StatusPending, StatusApprovedByMentor, user.Student, ErrForbidden},
		{"mentor cannot cancel", StatusPending, StatusCancelled, user.Mentor, ErrForbidden},
		{"mentor cannot accept approved", StatusApprovedByMentor, StatusAccepted, user.Mentor, ErrForbidden},
		{"curator cannot reject as mentor", StatusPending, StatusRejectedByMentor, user.Curator, ErrForbidden},
		{"student cannot cancel approved", StatusApprovedByMentor, StatusCancelled, user.Student, ErrInvalidTransition},
		{"mentor cannot reject approved", StatusApprovedByMentor, StatusRejectedByMentor, user.Mentor, ErrInvalidTransition},

		{"accepted is final", StatusAccepted, StatusRejectedByCurator, user.Curator, ErrInvalidTransition},
		{"rejected is final", StatusRejectedByMentor, StatusApprovedByMentor, user.Mentor, ErrInvalidTransition},
		{"cancelled is final", StatusCancelled, StatusPending, user.Student, ErrInvalidTransition},
		{"pending to pending", StatusPending, StatusPending, user.Student, ErrInvalidTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckTransition(tt.from, tt.to, tt.role)
			if !errors.Is(got, tt.want) {
				t.Fatalf("CheckTransition(%s -> %s, %s) = %v, want %v", tt.from, tt.to, tt.role, got, tt.want)
			}
		})
	}
}
