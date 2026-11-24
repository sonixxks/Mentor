package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSON читает тело запроса (до 16 КБ, без неизвестных полей) и при ошибке
// сам отвечает клиентом в формате контракта. Пустое тело — ошибка, если allowEmpty = false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, allowEmpty bool) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)
	if errors.Is(err, io.EOF) && allowEmpty {
		return true
	}
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			SendAPIError(w, http.StatusRequestEntityTooLarge, CodeValidation, "request body too large")
			return false
		}
		SendBadRequest(w, "invalid JSON body", err.Error())
		return false
	}
	return true
}
