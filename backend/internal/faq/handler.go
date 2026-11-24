package faq

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
)

const maxQuestionLen = 500

type Handler struct {
	bot *Bot
}

func NewHandler(bot *Bot) *Handler {
	return &Handler{bot: bot}
}

// Register: ручки публичные — FAQ доступен и до входа.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /faq", h.list)
	mux.HandleFunc("POST /faq/ask", h.ask)
}

type questionJSON struct {
	ID       string `json:"id"`
	Question string `json:"question"`
}

type entryJSON struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type askBody struct {
	Question string `json:"question"`
}

type askResponse struct {
	Matched     bool           `json:"matched"`
	QuestionID  *string        `json:"question_id"`
	Answer      string         `json:"answer"`
	Suggestions []questionJSON `json:"suggestions"`
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request) {
	entries := h.bot.Entries()
	out := make([]entryJSON, len(entries))
	for i, e := range entries {
		out[i] = entryJSON{ID: e.ID, Question: e.Question, Answer: e.Answer}
	}
	httpx.SendJSON(w, http.StatusOK, out)
}

func (h *Handler) ask(w http.ResponseWriter, r *http.Request) {
	var body askBody
	if !httpx.DecodeJSON(w, r, &body, false) {
		return
	}
	q := strings.TrimSpace(body.Question)
	if q == "" || utf8.RuneCountInString(q) > maxQuestionLen {
		httpx.SendBadRequest(w, "validation error", "question: must be 1-500 characters")
		return
	}

	ans := h.bot.Ask(q)
	resp := askResponse{Matched: ans.Matched, Answer: ans.Text, Suggestions: make([]questionJSON, len(ans.Suggestions))}
	if ans.Entry != nil {
		resp.QuestionID = &ans.Entry.ID
	}
	for i, s := range ans.Suggestions {
		resp.Suggestions[i] = questionJSON{ID: s.ID, Question: s.Question}
	}
	httpx.SendJSON(w, http.StatusOK, resp)
}
