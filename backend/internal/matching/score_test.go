package matching

import "testing"

func ptr[T any](v T) *T { return &v }

func TestCalculate(t *testing.T) {
	student := Student{
		TargetSkills:       []string{"Go", "SQL", "Docker", "Kubernetes"},
		CurrentSkills:      []Skill{{"Python", 2}, {"Git", 1}}, // среднее 1,5 — beginner
		MainGoals:          []string{"find_job_or_internship", "develop_soft_skills"},
		AvailabilitySlots:  map[string][]int{"mon": {19, 20, 21}, "sat": {10, 11}},
		TimeLimit:          ptr(4),
		CommunicationStyle: ptr("friendly"),
	}
	mentor := Mentor{
		CurrentSkills:       []Skill{{"go", 3}, {"SQL", 2}, {"Docker", 3}},
		TargetStudentLevels: []string{"beginner", "has_basics"},
		HelpTopics:          []string{"interview_prep", "code_review"},
		AvailabilitySlots:   map[string][]int{"mon": {19, 20}, "sat": {10, 11, 12}},
		CommunicationStyle:  ptr("free"),
	}

	got := Calculate(student, mentor)
	want := Info{Percent: 73, Breakdown: Breakdown{Skills: 75, Goals: 50, Time: 100, Style: 50}}
	if got != want {
		t.Fatalf("Calculate() = %+v, want %+v", got, want)
	}

	// Уровень ученика не подходит наставнику — навыки вдвое меньше.
	mentor.TargetStudentLevels = []string{"advanced"}
	if got := Calculate(student, mentor).Breakdown.Skills; got != 38 {
		t.Fatalf("skills with wrong level = %d, want 38", got)
	}

	// Совпало меньше часов, чем ученик готов заниматься.
	student.TimeLimit = ptr(8)
	if got := Calculate(student, mentor).Breakdown.Time; got != 80 {
		t.Fatalf("time = %d, want 80 (4 of 5 hours)", got)
	}

	mentor.CommunicationStyle = ptr("structured")
	if got := Calculate(student, mentor).Breakdown.Style; got != 0 {
		t.Fatalf("style = %d, want 0", got)
	}

	if got := Calculate(Student{}, mentor); got.Percent != 0 {
		t.Fatalf("empty student = %+v", got)
	}
}

func TestStudentLevel(t *testing.T) {
	tests := []struct {
		skills []Skill
		want   string
	}{
		{nil, "beginner"},
		{[]Skill{{"a", 1}, {"b", 2}}, "beginner"},
		{[]Skill{{"a", 2}, {"b", 3}}, "has_basics"},
		{[]Skill{{"a", 3}, {"b", 2}, {"c", 3}}, "advanced"},
	}
	for _, tt := range tests {
		if got := StudentLevel(tt.skills); got != tt.want {
			t.Fatalf("StudentLevel(%v) = %s, want %s", tt.skills, got, tt.want)
		}
	}
}
