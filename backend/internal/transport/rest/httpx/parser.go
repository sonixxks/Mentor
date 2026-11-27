package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxRequestBytes = 16 << 10

func ReadBodyWithLimit(w http.ResponseWriter, r *http.Request, req any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	decodeErr := decoder.Decode(req)
	if decodeErr != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(decodeErr, &maxBytesErr) {
			SendError(w, http.StatusRequestEntityTooLarge, ErrorResponse{ErrorCode: ErrValidation, Message: "request body too large"})
			return false
		}

		SendError(w, http.StatusBadRequest, ErrorResponse{ErrorCode: ErrValidation, Message: "parsing request body failed"})
		return false
	}
	return true
}
