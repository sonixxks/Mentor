package httpx

import (
	"net/http"
)

// Коды ошибок из контракта API (components.schemas.ErrorResponse).
const (
	CodeValidation          = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeProfileIncomplete   = "PROFILE_INCOMPLETE"
	CodeMentorLimitReached  = "MENTOR_LIMIT_REACHED"
	CodeRequestLimitReached = "REQUEST_LIMIT_REACHED"
	CodeDuplicateRequest    = "DUPLICATE_REQUEST"
	CodeInvalidTransition   = "INVALID_TRANSITION"
	CodeInternal            = "INTERNAL_ERROR"
)

// APIError — тело ошибки по контракту: {"error_code": "...", "message": "...", "details": [...]}.
type APIError struct {
	ErrorCode string   `json:"error_code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
}

func SendAPIError(w http.ResponseWriter, status int, code, message string, details ...string) {
	SendJSON(w, status, APIError{ErrorCode: code, Message: message, Details: details})
}

func SendInternalError(w http.ResponseWriter) {
	SendAPIError(w, http.StatusInternalServerError, CodeInternal, "internal server error")
}

func SendBadRequest(w http.ResponseWriter, message string, details ...string) {
	SendAPIError(w, http.StatusBadRequest, CodeValidation, message, details...)
}
