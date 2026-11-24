// Package faq — простой FAQ-бот: ищет ответ в базе вопросов по ключевым словам.
//
// База лежит в faq.json рядом с кодом (встраивается в бинарник): у каждого вопроса
// есть список ключевых основ слов («заявк» найдёт «заявка», «заявку», «заявки»).
package faq

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

//go:embed faq.json
var defaultBase []byte

type Entry struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Keywords []string `json:"keywords"`
}

type Answer struct {
	Matched     bool
	Entry       *Entry
	Text        string
	Suggestions []Entry
}

const (
	maxSuggestions = 3
	fallbackText   = "Не нашёл ответа на этот вопрос. Попробуйте переформулировать или выберите похожий вопрос ниже. " +
		"Если нужна помощь с подбором наставника, отметьте в профиле «Нужна помощь куратора»."
)

type Bot struct {
	entries []Entry
	// weight — вес ключа: 1 / число вопросов, где он встречается.
	// Общие слова («заявк») весят меньше уникальных («отмен»).
	weight map[string]float64
}

// New загружает встроенную базу вопросов.
func New() (*Bot, error) {
	return NewFromJSON(defaultBase)
}

func NewFromJSON(data []byte) (*Bot, error) {
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse faq base: %w", err)
	}
	freq := map[string]int{}
	for i := range entries {
		if entries[i].ID == "" || entries[i].Answer == "" || len(entries[i].Keywords) == 0 {
			return nil, fmt.Errorf("faq entry %d: id, answer and keywords are required", i)
		}
		for j, k := range entries[i].Keywords {
			entries[i].Keywords[j] = normalize(k)
			freq[entries[i].Keywords[j]]++
		}
	}

	weight := make(map[string]float64, len(freq))
	for k, n := range freq {
		weight[k] = 1 / float64(n)
	}
	return &Bot{entries: entries, weight: weight}, nil
}

func (b *Bot) Entries() []Entry {
	return b.entries
}

// Ask подбирает ответ: побеждает вопрос с наибольшим суммарным весом совпавших
// ключевых слов (при равенстве — тот, что выше в базе). Остальные совпавшие — подсказки.
func (b *Bot) Ask(question string) Answer {
	tokens := tokenize(question)
	text := " " + strings.Join(tokens, " ") + " "

	type scored struct {
		idx   int
		score float64
	}
	var hits []scored
	for i, e := range b.entries {
		if s := b.score(e, tokens, text); s > 0 {
			hits = append(hits, scored{i, s})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })

	if len(hits) == 0 {
		return Answer{Text: fallbackText, Suggestions: b.firstEntries(maxSuggestions)}
	}

	best := b.entries[hits[0].idx]
	ans := Answer{Matched: true, Entry: &best, Text: best.Answer, Suggestions: []Entry{}}
	for _, h := range hits[1:] {
		if len(ans.Suggestions) == maxSuggestions {
			break
		}
		ans.Suggestions = append(ans.Suggestions, b.entries[h.idx])
	}
	return ans
}

// score — суммарный вес найденных ключевых слов. Ключ из одного слова совпадает
// с началом любого слова; ключ из нескольких слов ищется как фраза.
func (b *Bot) score(e Entry, tokens []string, text string) float64 {
	var sum float64
	for _, k := range e.Keywords {
		if matches(k, tokens, text) {
			sum += b.weight[k]
		}
	}
	return sum
}

func matches(keyword string, tokens []string, text string) bool {
	if strings.Contains(keyword, " ") {
		return strings.Contains(text, " "+keyword)
	}
	for _, t := range tokens {
		if strings.HasPrefix(t, keyword) {
			return true
		}
	}
	return false
}

func (b *Bot) firstEntries(n int) []Entry {
	return b.entries[:min(n, len(b.entries))]
}

func normalize(s string) string {
	return strings.Join(tokenize(s), " ")
}

// tokenize: нижний регистр, ё -> е, слова из букв и цифр.
func tokenize(s string) []string {
	s = strings.ReplaceAll(strings.ToLower(s), "ё", "е")
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
