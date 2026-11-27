package httpx

import (
	"net/url"
	"testing"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    Page
		wantErr bool
	}{
		{"defaults", "", Page{Number: 1, Limit: 10}, false},
		{"explicit", "page=3&limit=20", Page{Number: 3, Limit: 20}, false},
		{"zero page", "page=0", Page{}, true},
		{"limit over max", "limit=51", Page{}, true},
		{"not a number", "limit=ten", Page{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _ := url.ParseQuery(tt.query)
			got, err := ParsePage(q, 10, 50)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("ParsePage() = %+v, want %+v", got, tt.want)
			}
		})
	}

	if off := (Page{Number: 3, Limit: 20}).Offset(); off != 40 {
		t.Fatalf("Offset() = %d, want 40", off)
	}
}

func TestQueryParser(t *testing.T) {
	q, _ := url.ParseQuery("skills=Go,%20SQL,,&has_free_slots=yes&min_rating=4.5&sort=price_asc&hours_min=11")
	p := NewQueryParser(q)

	if got := p.List("skills"); len(got) != 2 || got[0] != "Go" || got[1] != "SQL" {
		t.Fatalf("List() = %v", got)
	}
	if got := p.Float("min_rating", 0, 5); got == nil || *got != 4.5 {
		t.Fatalf("Float() = %v", got)
	}
	if got := p.Enum("sort", "match", "match", "price_asc"); got != "price_asc" {
		t.Fatalf("Enum() = %q", got)
	}
	if p.Bool("has_free_slots") != nil || p.Int("hours_min", 1, 10) != nil {
		t.Fatal("invalid values must be reported, not returned")
	}
	if len(p.Errors()) != 2 {
		t.Fatalf("Errors() = %v, want 2 errors", p.Errors())
	}
}
