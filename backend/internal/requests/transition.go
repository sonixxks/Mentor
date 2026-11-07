package requests

// allowedTransitions: из какого статуса, в какой и какой роли это разрешено.
// В MVP все статусы, кроме pending, финальные.
var allowedTransitions = map[Status]map[Status]Role{
	StatusPending: {
		StatusCancelled:         RoleStudent,
		StatusAccepted:          RoleMentor,
		StatusRejectedByMentor:  RoleMentor,
		StatusRejectedByCurator: RoleCurator,
	},
}

// CheckTransition проверяет только роль. Что студент — автор заявки,
// а ментор — её адресат, проверяет сервис.
func CheckTransition(from, to Status, role Role) error {
	allowedRole, ok := allowedTransitions[from][to]
	if !ok {
		return ErrInvalidTransition
	}
	if allowedRole != role {
		return ErrForbidden
	}
	return nil
}
