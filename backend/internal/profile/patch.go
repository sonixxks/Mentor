package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// readOnlyFields приходят в GET и могут вернуться в PATCH целиком — их просто пропускаем.
var readOnlyFields = set("id", "user_id", "role", "avatar_url", "verification_status", "verification_assets", "free_slots")

// patcher применяет к профилю только переданные поля и копит ошибки по ним.
type patcher struct {
	raw     map[string]json.RawMessage
	details []string
}

func (pt *patcher) fail(field, msg string) {
	pt.details = append(pt.details, field+": "+msg)
}

// value декодирует поле. ok = false — поля нет или оно с ошибкой; null — поле передано как null.
func (pt *patcher) value(field string, dst any) (ok, null bool) {
	raw, present := pt.raw[field]
	if !present {
		return false, false
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return true, true
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		pt.fail(field, "неверный тип значения")
		return false, false
	}
	return true, false
}

func (pt *patcher) requiredString(field string, maxLen int, dst *string) {
	var v string
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	v = strings.TrimSpace(v)
	switch {
	case null || v == "":
		pt.fail(field, "обязательное поле")
	case utf8.RuneCountInString(v) > maxLen:
		pt.fail(field, fmt.Sprintf("не длиннее %d символов", maxLen))
	default:
		*dst = v
	}
}

// optString: пустая строка и null очищают поле.
func (pt *patcher) optString(field string, maxLen int, dst **string) {
	var v string
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	v = strings.TrimSpace(v)
	switch {
	case null || v == "":
		*dst = nil
	case maxLen > 0 && utf8.RuneCountInString(v) > maxLen:
		pt.fail(field, fmt.Sprintf("не длиннее %d символов", maxLen))
	default:
		*dst = &v
	}
}

func (pt *patcher) optInt(field string, minVal, maxVal int, dst **int) {
	var v int
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	switch {
	case null:
		*dst = nil
	case v < minVal || v > maxVal:
		pt.fail(field, fmt.Sprintf("от %d до %d", minVal, maxVal))
	default:
		*dst = &v
	}
}

func (pt *patcher) optFloat(field string, minVal float64, dst **float64) {
	var v float64
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	switch {
	case null:
		*dst = nil
	case v < minVal || v > 1e8:
		pt.fail(field, "недопустимое значение")
	default:
		*dst = &v
	}
}

func (pt *patcher) optEnum(field string, allowed map[string]bool, dst **string) {
	var v string
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	switch {
	case null:
		*dst = nil
	case !allowed[v]:
		pt.fail(field, "допустимые значения: "+keys(allowed))
	default:
		*dst = &v
	}
}

func (pt *patcher) boolean(field string, dst *bool) {
	var v bool
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	if null {
		pt.fail(field, "должно быть true или false")
		return
	}
	*dst = v
}

// enumList: список значений из справочника без повторов; null или [] очищают поле.
func (pt *patcher) enumList(field string, allowed map[string]bool, maxItems int, dst *[]string) {
	var v []string
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	if null || len(v) == 0 {
		*dst = nil
		return
	}
	if maxItems > 0 && len(v) > maxItems {
		pt.fail(field, fmt.Sprintf("не больше %d значений", maxItems))
		return
	}
	seen := map[string]bool{}
	for _, item := range v {
		if !allowed[item] {
			pt.fail(field, "допустимые значения: "+keys(allowed))
			return
		}
		if seen[item] {
			pt.fail(field, "значения не должны повторяться")
			return
		}
		seen[item] = true
	}
	*dst = v
}

func (pt *patcher) availability(field string, dst *map[string][]int) {
	var v map[string][]int
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	if null || len(v) == 0 {
		*dst = nil
		return
	}
	for day, hours := range v {
		if !weekDays[day] {
			pt.fail(field, "дни недели: "+keys(weekDays))
			return
		}
		seen := map[int]bool{}
		for _, h := range hours {
			if h < 8 || h > 21 || seen[h] {
				pt.fail(field, "часы от 8 до 21 без повторов")
				return
			}
			seen[h] = true
		}
		sort.Ints(hours)
		if len(hours) == 0 {
			delete(v, day)
		}
	}
	*dst = v
}

func (pt *patcher) notificationPrefs(field string, dst *map[string]Channels) {
	var v map[string]Channels
	ok, null := pt.value(field, &v)
	if !ok {
		return
	}
	if null {
		*dst = nil
		return
	}
	for kind := range v {
		if !notificationKinds[kind] {
			pt.fail(field, "категории: "+keys(notificationKinds))
			return
		}
	}
	*dst = v
}

func keys(m map[string]bool) string {
	list := make([]string, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}
