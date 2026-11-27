package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxRequestBytes = 16 << 10

func ReadBodyWithLimit(w http.ResponseWriter, r *http.Request, req any) bool {
	return readJSON(w, r, req, false, false)
}

// DecodeJSON возвращает клиенту подробности ошибки декодирования.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, allowEmpty bool) bool {
	return readJSON(w, r, dst, allowEmpty, true)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any, allowEmpty, includeDetails bool) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	decodeErr := decoder.Decode(dst)
	if decodeErr != nil {
		if allowEmpty && errors.Is(decodeErr, io.EOF) {
			return true
		}
		var maxBytesErr *http.MaxBytesError

		if errors.As(decodeErr, &maxBytesErr) {
			SendError(w, http.StatusRequestEntityTooLarge, ErrorResponse{ErrorCode: ErrValidation, Message: "request body too large"})
			return false
		}

		if includeDetails {
			SendBadRequest(w, "invalid JSON body", decodeErr.Error())
		} else {
			SendBadRequest(w, "parsing request body failed")
		}
		return false
	}
	return true
}
