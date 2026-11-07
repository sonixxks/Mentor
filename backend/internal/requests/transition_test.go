package requests

import (
	"errors"
	"testing"
)

func TestCheckTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		role Role
		want error
	}{
		{"student cancels pending", StatusPending, StatusCancelled, RoleStudent, nil},
		{"mentor accepts pending", StatusPending, StatusAccepted, RoleMentor, nil},
		{"mentor rejects pending", StatusPending, StatusRejectedByMentor, RoleMentor, nil},
		{"curator rejects pending", StatusPending, StatusRejectedByCurator, RoleCurator, nil},

		{"student cannot accept", StatusPending, StatusAccepted, RoleStudent, ErrForbidden},
		{"mentor cannot cancel", StatusPending, StatusCancelled, RoleMentor, ErrForbidden},
		{"curator cannot reject as mentor", StatusPending, StatusRejectedByMentor, RoleCurator, ErrForbidden},
		{"mentor cannot reject as curator", StatusPending, StatusRejectedByCurator, RoleMentor, ErrForbidden},

		{"student cannot cancel accepted", StatusAccepted, StatusCancelled, RoleStudent, ErrInvalidTransition},
		{"curator cannot reject accepted", StatusAccepted, StatusRejectedByCurator, RoleCurator, ErrInvalidTransition},
		{"rejected is final", StatusRejectedByMentor, StatusAccepted, RoleMentor, ErrInvalidTransition},
		{"cancelled is final", StatusCancelled, StatusPending, RoleStudent, ErrInvalidTransition},
		{"pending to pending", StatusPending, StatusPending, RoleStudent, ErrInvalidTransition},
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
