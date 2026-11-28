package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// Store — хранилище анкет (реализует Repository).
type Store interface {
	Get(ctx context.Context, userID uuid.UUID) (Profile, error)
	Update(ctx context.Context, userID uuid.UUID, change func(*Profile) error) (Profile, error)
	CanonicalSkill(ctx context.Context, name string) (string, error)
}

type Storage interface {
	Save(ctx context.Context, folder, originalName string, r io.Reader) (string, error)
	Delete(ctx context.Context, url string) error
}

type Service struct {
	repo    Store
	storage Storage
}

func NewService(repo Store, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID) (Profile, error) {
	return s.repo.Get(ctx, userID)
}

// Patch обновляет только переданные поля (анкету можно сохранять по шагам).
// onboarding_completed: true проверяет обязательные для роли поля.
func (s *Service) Patch(ctx context.Context, userID uuid.UUID, raw map[string]json.RawMessage) (Profile, error) {
	return s.repo.Update(ctx, userID, func(p *Profile) error {
		pt := &patcher{raw: raw}
		if err := s.apply(ctx, pt, p); err != nil {
			return err
		}
		if len(pt.details) > 0 {
			return &ValidationError{Details: pt.details}
		}

		// Первая публикация с материалами ставит наставника в очередь верификации.
		if p.Role == user.Mentor && p.IsPublished && p.VerificationStatus == "none" && len(p.VerificationAssets) > 0 {
			p.VerificationStatus = "pending"
		}
		if p.OnboardingCompleted {
			if _, sent := raw["onboarding_completed"]; sent {
				if missing := missingRequired(p); len(missing) > 0 {
					return &ValidationError{Details: missing}
				}
			}
		}
		return nil
	})
}

func (s *Service) apply(ctx context.Context, pt *patcher, p *Profile) error {
	for field := range pt.raw {
		if !knownFields[field] && !readOnlyFields[field] {
			pt.fail(field, "неизвестное поле")
		}
	}

	pt.requiredString("first_name", 255, &p.FirstName)
	pt.requiredString("last_name", 255, &p.LastName)
	pt.optString("middle_name", 255, &p.MiddleName)

	pt.optInt("age", 14, 100, &p.Age)
	pt.optEnum("current_status", currentStatuses, &p.CurrentStatus)
	pt.optString("university", 255, &p.University)
	pt.optString("major", 255, &p.Major)
	pt.optInt("course", 1, 6, &p.Course)

	experience := studentExperience
	maxHours := 8
	minLevel := 1
	if p.Role == user.Mentor {
		experience, maxHours, minLevel = mentorExperience, 10, 2
	}
	pt.optEnum("experience", experience, &p.Experience)

	pt.optString("job_title", 255, &p.JobTitle)
	pt.optString("company", 255, &p.Company)
	pt.optString("bio", 300, &p.Bio)
	pt.optFloat("hourly_rate", 0, &p.HourlyRate)
	pt.optInt("max_students", 1, 10, &p.MaxStudents)
	if _, ok := pt.raw["is_published"]; ok && p.Role != user.Mentor {
		pt.fail("is_published", "только для наставника")
	} else {
		pt.boolean("is_published", &p.IsPublished)
	}

	pt.optEnum("communication_style", communicationStyles, &p.CommunicationStyle)
	pt.availability("availability_slots", &p.AvailabilitySlots)
	pt.enumList("main_goals", mainGoals, 2, &p.MainGoals)
	pt.optString("success_metric", 1000, &p.SuccessMetric)
	pt.optInt("time_limit", 1, maxHours, &p.TimeLimit)
	pt.boolean("needs_curator_help", &p.NeedsCuratorHelp)
	pt.optInt("min_duration_months", 1, 12, &p.MinDurationMonths)
	pt.optInt("max_duration_months", 1, 12, &p.MaxDurationMonths)
	if p.MinDurationMonths != nil && p.MaxDurationMonths != nil && *p.MinDurationMonths > *p.MaxDurationMonths {
		pt.fail("min_duration_months", "не больше max_duration_months")
	}

	pt.enumList("help_topics", helpTopics, 0, &p.HelpTopics)
	pt.enumList("target_student_levels", studentLevels, 0, &p.TargetStudentLevels)
	pt.notificationPrefs("notification_preferences", &p.NotificationPreferences)
	pt.boolean("onboarding_completed", &p.OnboardingCompleted)

	if err := s.applySkills(ctx, pt, p, minLevel); err != nil {
		return err
	}
	return nil
}

var knownFields = set(
	"first_name", "last_name", "middle_name", "age", "current_status", "university", "major", "course",
	"experience", "job_title", "company", "bio", "hourly_rate", "max_students", "is_published",
	"communication_style", "availability_slots", "main_goals", "success_metric", "time_limit",
	"needs_curator_help", "min_duration_months", "max_duration_months", "current_skills", "target_skills",
	"help_topics", "target_student_levels", "notification_preferences", "onboarding_completed",
)

// applySkills: навыки только из словаря (GET /skills), название приводится к словарному.
// Уровни: ученик 1–3, наставник 2–3.
func (s *Service) applySkills(ctx context.Context, pt *patcher, p *Profile, minLevel int) error {
	var current []Skill
	if ok, null := pt.value("current_skills", &current); ok {
		if null {
			current = nil
		}
		seen := map[string]bool{}
		for i, sk := range current {
			name, err := s.repo.CanonicalSkill(ctx, sk.Skill)
			if err != nil {
				return err
			}
			switch {
			case name == "":
				pt.fail("current_skills", fmt.Sprintf("навыка «%s» нет в словаре", sk.Skill))
			case sk.Level < minLevel || sk.Level > 3:
				pt.fail("current_skills", fmt.Sprintf("уровень навыка от %d до 3", minLevel))
			case seen[name]:
				pt.fail("current_skills", fmt.Sprintf("навык «%s» указан дважды", name))
			}
			seen[name] = true
			current[i].Skill = name
		}
		p.CurrentSkills = current
	}

	var target []string
	if ok, null := pt.value("target_skills", &target); ok {
		if null {
			target = nil
		}
		seen := map[string]bool{}
		for i, sk := range target {
			name, err := s.repo.CanonicalSkill(ctx, sk)
			if err != nil {
				return err
			}
			switch {
			case name == "":
				pt.fail("target_skills", fmt.Sprintf("навыка «%s» нет в словаре", sk))
			case seen[name]:
				pt.fail("target_skills", fmt.Sprintf("навык «%s» указан дважды", name))
			}
			seen[name] = true
			target[i] = name
		}
		p.TargetSkills = target
	}
	return nil
}

// missingRequired — незаполненные обязательные поля анкеты для роли.
func missingRequired(p *Profile) []string {
	type check struct {
		field  string
		filled bool
	}
	var checks []check
	switch p.Role {
	case user.Student:
		checks = []check{
			{"age", p.Age != nil},
			{"current_status", p.CurrentStatus != nil},
			{"main_goals", len(p.MainGoals) > 0},
			{"target_skills", len(p.TargetSkills) > 0},
			{"time_limit", p.TimeLimit != nil},
			{"availability_slots", len(p.AvailabilitySlots) > 0},
			{"communication_style", p.CommunicationStyle != nil},
		}
	case user.Mentor:
		checks = []check{
			{"job_title", p.JobTitle != nil},
			{"experience", p.Experience != nil},
			{"bio", p.Bio != nil},
			{"current_skills", len(p.CurrentSkills) > 0},
			{"help_topics", len(p.HelpTopics) > 0},
			{"target_student_levels", len(p.TargetStudentLevels) > 0},
			{"hourly_rate", p.HourlyRate != nil},
			{"time_limit", p.TimeLimit != nil},
			{"max_students", p.MaxStudents != nil},
			{"availability_slots", len(p.AvailabilitySlots) > 0},
			{"communication_style", p.CommunicationStyle != nil},
		}
	}

	var missing []string
	for _, c := range checks {
		if !c.filled {
			missing = append(missing, c.field+": обязательное поле для завершения анкеты")
		}
	}
	return missing
}

// UploadedFile — файл из multipart-запроса.
type UploadedFile struct {
	Name string
	Size int64
	Open func() (io.ReadCloser, error)
}

var allowedAssetExt = set(".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx")

// AddAssets добавляет документы и ссылки для верификации наставника. Если наставник уже
// проверен или ему отказали, статус возвращается в pending и он скрывается из каталога.
func (s *Service) AddAssets(ctx context.Context, actor uuid.UUID, role user.Role, files []UploadedFile, links []string) (Profile, error) {
	if role != user.Mentor {
		return Profile{}, ErrMentorOnly
	}

	var details []string
	if len(files) == 0 && len(links) == 0 {
		details = append(details, "files, links: нужен хотя бы один файл или ссылка")
	}
	if len(files) > MaxFilesPerCall {
		details = append(details, fmt.Sprintf("files: не больше %d за раз", MaxFilesPerCall))
	}
	if len(links) > MaxLinksPerCall {
		details = append(details, fmt.Sprintf("links: не больше %d за раз", MaxLinksPerCall))
	}
	for _, f := range files {
		if !allowedAssetExt[strings.ToLower(filepath.Ext(f.Name))] {
			details = append(details, fmt.Sprintf("files: «%s» — допустимы PDF, JPG, PNG, DOC, DOCX", f.Name))
		}
		if f.Size > MaxAssetFileSize {
			details = append(details, fmt.Sprintf("files: «%s» больше 10 МБ", f.Name))
		}
	}
	for _, l := range links {
		if u, err := url.Parse(l); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			details = append(details, fmt.Sprintf("links: «%s» — нужна ссылка http(s)", l))
		}
	}
	if len(details) > 0 {
		return Profile{}, &ValidationError{Details: details}
	}

	var added []Asset
	for _, f := range files {
		u, err := s.saveFile(ctx, actor, f)
		if err != nil {
			s.cleanup(ctx, added)
			return Profile{}, err
		}
		added = append(added, Asset{ID: uuid.New(), Type: "file", Name: f.Name, URL: u})
	}
	for _, l := range links {
		added = append(added, Asset{ID: uuid.New(), Type: "link", Name: l, URL: l})
	}

	p, err := s.repo.Update(ctx, actor, func(p *Profile) error {
		if len(p.VerificationAssets)+len(added) > MaxAssetsTotal {
			return &ValidationError{Details: []string{fmt.Sprintf("в профиле может быть не больше %d материалов", MaxAssetsTotal)}}
		}
		p.VerificationAssets = append(p.VerificationAssets, added...)
		switch {
		case p.VerificationStatus == "verified" || p.VerificationStatus == "rejected":
			p.VerificationStatus = "pending"
		case p.VerificationStatus == "none" && p.IsPublished:
			p.VerificationStatus = "pending"
		}
		return nil
	})
	if err != nil {
		s.cleanup(ctx, added)
		return Profile{}, err
	}
	return p, nil
}

func (s *Service) DeleteAsset(ctx context.Context, actor uuid.UUID, role user.Role, assetID uuid.UUID) (Profile, error) {
	if role != user.Mentor {
		return Profile{}, ErrMentorOnly
	}

	var removed *Asset
	p, err := s.repo.Update(ctx, actor, func(p *Profile) error {
		for i, a := range p.VerificationAssets {
			if a.ID == assetID {
				removed = &a
				p.VerificationAssets = append(p.VerificationAssets[:i], p.VerificationAssets[i+1:]...)
				return nil
			}
		}
		return ErrAssetNotFound
	})
	if err != nil {
		return Profile{}, err
	}
	if removed != nil && removed.Type == "file" {
		_ = s.storage.Delete(ctx, removed.URL)
	}
	return p, nil
}

func (s *Service) saveFile(ctx context.Context, owner uuid.UUID, f UploadedFile) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("open upload: %w", err)
	}
	defer func() { _ = rc.Close() }()
	return s.storage.Save(ctx, "verification/"+owner.String(), f.Name, io.LimitReader(rc, MaxAssetFileSize+1))
}

func (s *Service) cleanup(ctx context.Context, assets []Asset) {
	for _, a := range assets {
		if a.Type == "file" {
			_ = s.storage.Delete(ctx, a.URL)
		}
	}
}

// IsValidation — ошибка в данных запроса.
func IsValidation(err error) (*ValidationError, bool) {
	var v *ValidationError
	ok := errors.As(err, &v)
	return v, ok
}
