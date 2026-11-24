package faq

import "testing"

func TestDefaultBase(t *testing.T) {
	bot, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(bot.Entries()) < 10 {
		t.Fatalf("faq base has only %d entries", len(bot.Entries()))
	}

	seen := map[string]bool{}
	for _, e := range bot.Entries() {
		if seen[e.ID] {
			t.Fatalf("duplicate id %q", e.ID)
		}
		seen[e.ID] = true
	}
}

func TestAsk(t *testing.T) {
	bot, err := New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		question string
		want     string
	}{
		{"Сколько заявок я могу отправить одновременно?", "request_limit"},
		{"как ОТМЕНИТЬ заявку???", "cancel_request"},
		{"Забыл пароль, как войти", "password"},
		{"Что значит статус «одобрена наставником»?", "request_statuses"},
		{"Как проходит созвон с ментором", "meetings"},
		{"Нужна помощь куратора, не могу найти наставника", "curator_help"},
		{"Сколько стоит занятие?", "price"},
		{"как пройти верификацию, какие документы нужны", "mentor_verification"},
		{"Как считается процент совместимости", "match_percent"},
		{"в каком часовом поясе время встреч", "time_zone"},
	}

	for _, tt := range tests {
		t.Run(tt.question, func(t *testing.T) {
			ans := bot.Ask(tt.question)
			if !ans.Matched || ans.Entry.ID != tt.want {
				got := "<none>"
				if ans.Entry != nil {
					got = ans.Entry.ID
				}
				t.Fatalf("Ask(%q) = %s, want %s", tt.question, got, tt.want)
			}
			if ans.Text == "" || len(ans.Suggestions) > maxSuggestions {
				t.Fatalf("answer = %+v", ans)
			}
			for _, s := range ans.Suggestions {
				if s.ID == tt.want {
					t.Fatal("best answer must not repeat in suggestions")
				}
			}
		})
	}
}

func TestAskFallback(t *testing.T) {
	bot, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ans := bot.Ask("какая сегодня погода?")
	if ans.Matched || ans.Text != fallbackText || len(ans.Suggestions) != maxSuggestions {
		t.Fatalf("fallback = %+v", ans)
	}
	if ans := bot.Ask("   "); ans.Matched {
		t.Fatal("empty question must not match")
	}
}

func TestNewFromJSON(t *testing.T) {
	if _, err := NewFromJSON([]byte(`[{"id":"x","answer":"a","keywords":[]}]`)); err == nil {
		t.Fatal("entry without keywords must be rejected")
	}
	bot, err := NewFromJSON([]byte(`[{"id":"x","question":"q","answer":"a","keywords":["Ёлк", "не могу"]}]`))
	if err != nil {
		t.Fatal(err)
	}
	if !bot.Ask("Ёлки зелёные").Matched || !bot.Ask("Я не могу!").Matched || bot.Ask("могу").Matched {
		t.Fatal("stem and phrase matching do not work")
	}
}
