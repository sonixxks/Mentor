package curator

import (
	"testing"
	"time"
)

func TestEvaluate(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	daysAgo := func(d int) *time.Time {
		v := now.Add(-time.Duration(d) * 24 * time.Hour)
		return &v
	}
	rating := func(r int) *int { return &r }

	tests := []struct {
		name     string
		stats    WorkspaceStats
		severity Severity
		signals  []SignalType
		progress int
	}{
		{
			name:     "healthy pair",
			stats:    WorkspaceStats{Status: WorkspaceActive, StartedAt: *daysAgo(30), LastActivityAt: daysAgo(1), TasksTotal: 4, TasksCompleted: 3},
			severity: SeverityGreen,
			progress: 75,
		},
		{
			name:     "no activity for a week",
			stats:    WorkspaceStats{Status: WorkspaceActive, StartedAt: *daysAgo(30), LastActivityAt: daysAgo(7)},
			severity: SeverityYellow,
			signals:  []SignalType{SignalNoActivity},
		},
		{
			name:     "never active counts from start",
			stats:    WorkspaceStats{Status: WorkspaceActive, StartedAt: *daysAgo(10)},
			severity: SeverityYellow,
			signals:  []SignalType{SignalNoActivity},
		},
		{
			name:     "one overdue task is fine",
			stats:    WorkspaceStats{Status: WorkspaceActive, StartedAt: *daysAgo(3), TasksOverdue: 1, TasksTotal: 2},
			severity: SeverityGreen,
		},
		{
			name:     "two signals are red",
			stats:    WorkspaceStats{Status: WorkspaceActive, StartedAt: *daysAgo(3), TasksOverdue: 2, LastRating: rating(2)},
			severity: SeverityRed,
			signals:  []SignalType{SignalOverdueTasks, SignalBadReviews},
		},
		{
			name:     "stagnation is red without signals",
			stats:    WorkspaceStats{Status: WorkspaceStagnation, StartedAt: *daysAgo(1)},
			severity: SeverityRed,
		},
		{
			name:     "good last review",
			stats:    WorkspaceStats{Status: WorkspacePaused, StartedAt: *daysAgo(1), LastRating: rating(3)},
			severity: SeverityGreen,
		},
		{
			name:     "completed pair has no signals",
			stats:    WorkspaceStats{Status: WorkspaceCompleted, StartedAt: *daysAgo(90), TasksOverdue: 5, TasksTotal: 5, TasksCompleted: 5},
			severity: SeverityGreen,
			progress: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := Evaluate(tt.stats, now)
			if row.Severity != tt.severity || row.ProgressPercent != tt.progress {
				t.Fatalf("severity = %s, progress = %d, want %s, %d", row.Severity, row.ProgressPercent, tt.severity, tt.progress)
			}
			if len(row.Signals) != len(tt.signals) {
				t.Fatalf("signals = %+v, want %v", row.Signals, tt.signals)
			}
			for i, s := range row.Signals {
				if s.Type != tt.signals[i] {
					t.Fatalf("signals = %+v, want %v", row.Signals, tt.signals)
				}
			}
		})
	}
}

func TestSortRows(t *testing.T) {
	now := time.Now()
	at := func(d int) time.Time { return now.Add(-time.Duration(d) * 24 * time.Hour) }
	ptr := func(t time.Time) *time.Time { return &t }

	rows := []WorkspaceRow{
		{WorkspaceStats: WorkspaceStats{StudentName: "green", StartedAt: at(5), LastActivityAt: ptr(at(0))}, Severity: SeverityGreen},
		{WorkspaceStats: WorkspaceStats{StudentName: "yellow-old", StartedAt: at(40), LastActivityAt: ptr(at(20))}, Severity: SeverityYellow},
		{WorkspaceStats: WorkspaceStats{StudentName: "red", StartedAt: at(9), LastActivityAt: ptr(at(1))}, Severity: SeverityRed},
		{WorkspaceStats: WorkspaceStats{StudentName: "yellow-new", StartedAt: at(30), LastActivityAt: ptr(at(8))}, Severity: SeverityYellow},
	}

	SortRows(rows, SortSeverity, now)
	want := []string{"red", "yellow-old", "yellow-new", "green"}
	for i, r := range rows {
		if r.StudentName != want[i] {
			t.Fatalf("severity order[%d] = %s, want %s", i, r.StudentName, want[i])
		}
	}

	SortRows(rows, SortActivity, now)
	want = []string{"green", "red", "yellow-new", "yellow-old"}
	for i, r := range rows {
		if r.StudentName != want[i] {
			t.Fatalf("activity order[%d] = %s, want %s", i, r.StudentName, want[i])
		}
	}
}
