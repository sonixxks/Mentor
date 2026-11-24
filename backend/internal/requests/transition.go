package requests

import "github.com/sonixxks/Mentor/backend/internal/user"

// allowedTransitions — матрица переходов из контракта API v1.2:
//
//	ученик:    pending -> cancelled; approved_by_mentor -> accepted
//	наставник: pending -> approved_by_mentor / rejected_by_mentor
//	куратор:   любой нетерминальный статус -> rejected_by_curator
var allowedTransitions = map[Status]map[Status]user.Role{
	StatusPending: {
		StatusCancelled:         user.Student,
		StatusApprovedByMentor:  user.Mentor,
		StatusRejectedByMentor:  user.Mentor,
		StatusRejectedByCurator: user.Curator,
	},
	StatusApprovedByMentor: {
		StatusAccepted:          user.Student,
		StatusRejectedByCurator: user.Curator,
	},
}

// CheckTransition проверяет переход и роль. Что ученик — автор заявки,
// а наставник — её адресат, проверяет сервис.
func CheckTransition(from, to Status, role user.Role) error {
	allowedRole, ok := allowedTransitions[from][to]
	if !ok {
		return ErrInvalidTransition
	}
	if allowedRole != role {
		return ErrForbidden
	}
	return nil
}
