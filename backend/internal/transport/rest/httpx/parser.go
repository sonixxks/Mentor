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
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return false
		}

		SendError(w, http.StatusBadRequest, ErrorResponse{Error: "parsing request body failed"})
		return false
	}
	return true
}
