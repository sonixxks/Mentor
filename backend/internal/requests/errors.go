package requests

import "errors"

var (
	ErrNotFound               = errors.New("request not found")
	ErrForbidden              = errors.New("action is not allowed for this role")
	ErrInvalidStatus          = errors.New("unknown request status")
	ErrInvalidTransition      = errors.New("status transition is not allowed")
	ErrStatusConflict         = errors.New("request status was changed concurrently")
	ErrCuratorCommentRequired = errors.New("curator comment is required")
	ErrAlreadyPending         = errors.New("pending request to this mentor already exists")
	ErrInvalidTarget          = errors.New("mentor or service package not found")
)
