package faq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	bot, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandler(bot).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/faq/ask", strings.NewReader(`{"question":"Как отменить заявку?"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	var resp askResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Matched || resp.QuestionID == nil || *resp.QuestionID != "cancel_request" || resp.Suggestions == nil {
		t.Fatalf("response = %+v", resp)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/faq/ask", strings.NewReader(`{"question":"  "}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty question: status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/faq", nil))
	var list []entryJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != len(bot.Entries()) {
		t.Fatalf("list: %d entries, err = %v", len(list), err)
	}
}
