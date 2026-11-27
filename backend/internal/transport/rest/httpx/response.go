package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sonixxks/Mentor/backend/internal/platform/validation"
)

type ErrCode string

const (
	ErrValidation          ErrCode = "VALIDATION_ERROR"
	ErrUnauthorized        ErrCode = "UNAUTHORIZED"
	ErrForbidden           ErrCode = "FORBIDDEN"
	ErrNotFound            ErrCode = "NOT_FOUND"
	ErrConflict            ErrCode = "CONFLICT"
	ErrInvalidCode         ErrCode = "INVALID_CODE"
	ErrTooManyRequests     ErrCode = "TOO_MANY_REQUESTS"
	ErrRoleRequired        ErrCode = "ROLE_REQUIRED"
	ErrProfileIncomplete   ErrCode = "PROFILE_INCOMPLETE"
	ErrMentorLimitReached  ErrCode = "MENTOR_LIMIT_REACHED"
	ErrRequestLimitReached ErrCode = "REQUEST_LIMIT_REACHED"
	ErrDuplicateRequest    ErrCode = "DUPLICATE_REQUEST"
	ErrReviewExists        ErrCode = "REVIEW_EXISTS"
	ErrInvalidToken        ErrCode = "INVALID_TOKEN"
	ErrInvalidTransition   ErrCode = "INVALID_TRANSITION"
	ErrInternal            ErrCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	ErrorCode ErrCode  `json:"error_code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
}

func SendValidationErrors(w http.ResponseWriter, err error, statusCode int) {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		slog.Error("Invalid request",
			"error", err)
		SendError(w, http.StatusInternalServerError, ErrorResponse{ErrorCode: ErrInternal, Message: "Internal server error"})
		return
	}

	details := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		details = append(details, fe.Field()+validation.ValidationMessage(fe))
	}

	resp := ErrorResponse{ErrorCode: ErrValidation, Message: "Validation failed", Details: details}

	SendError(w, statusCode, resp)
}

func SendError(w http.ResponseWriter, statusCode int, resp ErrorResponse) {
	SendJSON(w, statusCode, resp)
}

func SendJSON(w http.ResponseWriter, statusCode int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode response",
			"error", err)
		return
	}
}

func SendInternalError(w http.ResponseWriter) {
	SendError(w, http.StatusInternalServerError, ErrorResponse{ErrorCode: ErrInternal, Message: "internal server error"})
}

func SendBadRequest(w http.ResponseWriter, message string, details ...string) {
	SendError(w, http.StatusBadRequest, ErrorResponse{ErrorCode: ErrValidation, Message: message, Details: details})
}
