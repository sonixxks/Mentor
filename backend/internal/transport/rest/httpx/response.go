package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sonixxks/Mentor/backend/internal/platform/validation"
)

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func SendValidationErrors(w http.ResponseWriter, err error, statusCode int) {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		slog.Error("Invalid request",
			"error", err)
		SendError(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	fields := make(map[string]string, len(validationErrs))
	for _, fe := range validationErrs {
		fields[fe.Field()] = validation.ValidationMessage(fe)
	}

	resp := ErrorResponse{Error: "validation error", Fields: fields}

	SendError(w, statusCode, resp)
}

func SendError(w http.ResponseWriter, statusCode int, resp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode Error response",
			"error", err)
		return
	}
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
