package requests

import "errors"

var (
	ErrNotFound          = errors.New("request not found")
	ErrForbidden         = errors.New("action is not allowed for this role")
	ErrInvalidStatus     = errors.New("unknown request status")
	ErrInvalidTransition = errors.New("status transition is not allowed")
	ErrStatusConflict    = errors.New("request status was changed concurrently")
	ErrReasonRequired    = errors.New("rejection reason is required")
	ErrCoverLetterLength = errors.New("cover letter is too long")

	ErrMentorNotFound  = errors.New("mentor not found")
	ErrPackageNotFound = errors.New("service package not found")
	ErrStudentNotFound = errors.New("student not found")
	ErrNoHelpRequest   = errors.New("student did not ask curator for help")

	ErrRequestLimit = errors.New("student already has the maximum number of active requests")
	ErrDuplicate    = errors.New("active request to this mentor already exists")
	ErrMentorLimit  = errors.New("mentor has no free slots")
)
