package curator

import (
	"sort"
	"time"
)

// Пороги сигналов из контракта (GET /curator/workspaces).
const (
	NoActivityDays    = 7
	OverdueTasksLimit = 2
	BadRatingMax      = 2
)

// Evaluate считает сигналы, светофор и прогресс пары на момент now.
// У завершённых пар сигналов нет.
func Evaluate(s WorkspaceStats, now time.Time) WorkspaceRow {
	row := WorkspaceRow{WorkspaceStats: s, Signals: []Signal{}, Severity: SeverityGreen}

	if s.TasksTotal > 0 {
		row.ProgressPercent = s.TasksCompleted * 100 / s.TasksTotal
	}

	if s.Status == WorkspaceCompleted {
		return row
	}

	if days := InactiveDays(s, now); days >= NoActivityDays {
		row.Signals = append(row.Signals, Signal{Type: SignalNoActivity, Value: days})
	}
	if s.TasksOverdue >= OverdueTasksLimit {
		row.Signals = append(row.Signals, Signal{Type: SignalOverdueTasks, Value: s.TasksOverdue})
	}
	if s.LastRating != nil && *s.LastRating <= BadRatingMax {
		row.Signals = append(row.Signals, Signal{Type: SignalBadReviews, Value: *s.LastRating})
	}

	switch {
	case s.Status == WorkspaceStagnation || len(row.Signals) >= 2:
		row.Severity = SeverityRed
	case len(row.Signals) == 1:
		row.Severity = SeverityYellow
	}
	return row
}

// InactiveDays — полных дней без активности; если активности не было, считаем от создания пары.
func InactiveDays(s WorkspaceStats, now time.Time) int {
	last := s.StartedAt
	if s.LastActivityAt != nil && s.LastActivityAt.After(last) {
		last = *s.LastActivityAt
	}
	if !now.After(last) {
		return 0
	}
	return int(now.Sub(last) / (24 * time.Hour))
}

var severityRank = map[Severity]int{SeverityRed: 2, SeverityYellow: 1, SeverityGreen: 0}

// SortRows: severity — сначала проблемные (красные, жёлтые), внутри — дольше без активности;
// activity — по последней активности, свежие сверху.
func SortRows(rows []WorkspaceRow, by WorkspaceSort, now time.Time) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if by != SortActivity {
			if ra, rb := severityRank[a.Severity], severityRank[b.Severity]; ra != rb {
				return ra > rb
			}
			if da, db := InactiveDays(a.WorkspaceStats, now), InactiveDays(b.WorkspaceStats, now); da != db {
				return da > db
			}
			return a.StartedAt.Before(b.StartedAt)
		}
		return lastActivity(a).After(lastActivity(b))
	})
}

func lastActivity(r WorkspaceRow) time.Time {
	if r.LastActivityAt != nil {
		return *r.LastActivityAt
	}
	return r.StartedAt
}
