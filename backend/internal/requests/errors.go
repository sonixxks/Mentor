package requests

import "errors"

var (
	ErrNotFound          = errors.New("заявка не найдена")
	ErrForbidden         = errors.New("действие недоступно для вашей роли")
	ErrInvalidStatus     = errors.New("неизвестный статус заявки")
	ErrInvalidTransition = errors.New("такой переход статуса недопустим")
	ErrStatusConflict    = errors.New("статус заявки уже изменился, обновите страницу")
	ErrReasonRequired    = errors.New("укажите причину отклонения")
	ErrCoverLetterLength = errors.New("сопроводительное письмо длиннее 1000 символов")

	ErrMentorNotFound  = errors.New("наставник не найден")
	ErrPackageNotFound = errors.New("программа не найдена")
	ErrStudentNotFound = errors.New("ученик не найден")
	ErrNoHelpRequest   = errors.New("ученик не запрашивал помощь куратора")

	ErrRequestLimit = errors.New("у вас уже 3 активные заявки")
	ErrDuplicate    = errors.New("у вас уже есть активная заявка к этому наставнику")
	ErrMentorLimit  = errors.New("у наставника нет свободных мест")
)
