// Package matching считает совместимость ученика и наставника по анкетам.
// Формула из контракта (GET /mentors/recommendations): навыки 50%, цели 20%, время 20%, стиль 10%.
package matching

import (
	"math"
	"strings"
)

type Info struct {
	Percent   int       `json:"percent"`
	Breakdown Breakdown `json:"breakdown"`
}

type Breakdown struct {
	Skills int `json:"skills"`
	Goals  int `json:"goals"`
	Time   int `json:"time"`
	Style  int `json:"style"`
}

type Skill struct {
	Skill string `json:"skill"`
	Level int    `json:"level"`
}

// Student и Mentor — поля анкет, которые участвуют в расчёте.
type Student struct {
	TargetSkills       []string
	CurrentSkills      []Skill
	MainGoals          []string
	AvailabilitySlots  map[string][]int
	TimeLimit          *int
	CommunicationStyle *string
}

type Mentor struct {
	CurrentSkills       []Skill
	TargetStudentLevels []string
	HelpTopics          []string
	AvailabilitySlots   map[string][]int
	CommunicationStyle  *string
}

// goalTopics — какие темы наставника закрывают цель ученика.
var goalTopics = map[string][]string{
	"find_job_or_internship": {"first_job_in_it", "interview_prep", "portfolio"},
	"learn_new_skill":        {"code_review", "career_growth"},
	"build_project":          {"portfolio", "code_review"},
	"interview_prep":         {"interview_prep"},
	"change_direction":       {"change_direction", "career_growth"},
	"develop_soft_skills":    {"soft_skills"},
}

func Calculate(s Student, m Mentor) Info {
	b := Breakdown{
		Skills: skillsScore(s, m),
		Goals:  goalsScore(s, m),
		Time:   timeScore(s, m),
		Style:  styleScore(s, m),
	}
	total := 0.5*float64(b.Skills) + 0.2*float64(b.Goals) + 0.2*float64(b.Time) + 0.1*float64(b.Style)
	return Info{Percent: int(math.Round(total)), Breakdown: b}
}

// skillsScore — доля целевых навыков ученика, которые есть у наставника.
// Если уровень ученика не входит в target_student_levels наставника — вдвое меньше.
func skillsScore(s Student, m Mentor) int {
	if len(s.TargetSkills) == 0 {
		return 0
	}
	has := map[string]bool{}
	for _, sk := range m.CurrentSkills {
		has[strings.ToLower(sk.Skill)] = true
	}
	matched := 0
	for _, sk := range s.TargetSkills {
		if has[strings.ToLower(sk)] {
			matched++
		}
	}
	score := float64(matched) / float64(len(s.TargetSkills)) * 100

	if !contains(m.TargetStudentLevels, StudentLevel(s.CurrentSkills)) {
		score *= 0.5
	}
	return int(math.Round(score))
}

// StudentLevel — средний уровень навыков ученика: до 1,5 — beginner, до 2,5 — has_basics, иначе advanced.
// Без навыков ученик считается новичком.
func StudentLevel(skills []Skill) string {
	if len(skills) == 0 {
		return "beginner"
	}
	sum := 0
	for _, sk := range skills {
		sum += sk.Level
	}
	avg := float64(sum) / float64(len(skills))
	switch {
	case avg <= 1.5:
		return "beginner"
	case avg <= 2.5:
		return "has_basics"
	default:
		return "advanced"
	}
}

func goalsScore(s Student, m Mentor) int {
	if len(s.MainGoals) == 0 {
		return 0
	}
	matched := 0
	for _, goal := range s.MainGoals {
		for _, topic := range goalTopics[goal] {
			if contains(m.HelpTopics, topic) {
				matched++
				break
			}
		}
	}
	return int(math.Round(float64(matched) / float64(len(s.MainGoals)) * 100))
}

// timeScore — сколько нужных ученику часов совпадает со слотами наставника.
// Нужно не больше time_limit часов: если совпало столько, сколько ученик готов заниматься, — 100.
func timeScore(s Student, m Mentor) int {
	studentHours, overlap := 0, 0
	for day, hours := range s.AvailabilitySlots {
		mentorHours := map[int]bool{}
		for _, h := range m.AvailabilitySlots[day] {
			mentorHours[h] = true
		}
		for _, h := range hours {
			studentHours++
			if mentorHours[h] {
				overlap++
			}
		}
	}
	if studentHours == 0 {
		return 0
	}

	need := studentHours
	if s.TimeLimit != nil && *s.TimeLimit > 0 && *s.TimeLimit < need {
		need = *s.TimeLimit
	}
	if overlap > need {
		overlap = need
	}
	return int(math.Round(float64(overlap) / float64(need) * 100))
}

func styleScore(s Student, m Mentor) int {
	if s.CommunicationStyle == nil || m.CommunicationStyle == nil {
		return 0
	}
	switch {
	case *s.CommunicationStyle == *m.CommunicationStyle:
		return 100
	case *s.CommunicationStyle == "free" || *m.CommunicationStyle == "free":
		return 50
	default:
		return 0
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
