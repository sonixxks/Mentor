package httpx

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Page — номер страницы и размер, как в параметрах page/limit контракта.
type Page struct {
	Number int
	Limit  int
}

func (p Page) Offset() int {
	return (p.Number - 1) * p.Limit
}

// ParsePage читает page (>= 1) и limit (1..maxLimit) из query.
func ParsePage(q url.Values, defaultLimit, maxLimit int) (Page, error) {
	page := Page{Number: 1, Limit: defaultLimit}

	if raw := q.Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return Page{}, fmt.Errorf("page must be a positive integer")
		}
		page.Number = n
	}

	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > maxLimit {
			return Page{}, fmt.Errorf("limit must be between 1 and %d", maxLimit)
		}
		page.Limit = n
	}

	return page, nil
}

// PathUUID разбирает UUID из сегмента пути (r.PathValue).
func PathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a valid uuid", name)
	}
	return id, nil
}

// QueryParser собирает ошибки разбора query-параметров, чтобы вернуть их все разом в details.
type QueryParser struct {
	q    url.Values
	errs []string
}

func NewQueryParser(q url.Values) *QueryParser {
	return &QueryParser{q: q}
}

func (p *QueryParser) Errors() []string {
	return p.errs
}

func (p *QueryParser) String(name string) string {
	return strings.TrimSpace(p.q.Get(name))
}

func (p *QueryParser) Bool(name string) *bool {
	raw := p.String(name)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		p.errs = append(p.errs, name+": must be true or false")
		return nil
	}
	return &v
}

func (p *QueryParser) Int(name string, minVal, maxVal int) *int {
	raw := p.String(name)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < minVal || v > maxVal {
		p.errs = append(p.errs, fmt.Sprintf("%s: must be an integer between %d and %d", name, minVal, maxVal))
		return nil
	}
	return &v
}

func (p *QueryParser) Float(name string, minVal, maxVal float64) *float64 {
	raw := p.String(name)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < minVal || v > maxVal {
		p.errs = append(p.errs, fmt.Sprintf("%s: must be a number between %g and %g", name, minVal, maxVal))
		return nil
	}
	return &v
}

// Enum возвращает значение, если оно входит в allowed, иначе def (для пустого) или ошибку.
func (p *QueryParser) Enum(name, def string, allowed ...string) string {
	raw := p.String(name)
	if raw == "" {
		return def
	}
	for _, a := range allowed {
		if raw == a {
			return raw
		}
	}
	p.errs = append(p.errs, name+": must be one of "+strings.Join(allowed, ", "))
	return def
}

// List — значения через запятую без пустых элементов.
func (p *QueryParser) List(name string) []string {
	raw := p.String(name)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}
