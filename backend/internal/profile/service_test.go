package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres/pgtest"
	"github.com/sonixxks/Mentor/backend/internal/platform/storage"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

func newService(t *testing.T) (*Service, *pgtest.Fixtures, string) {
	t.Helper()
	pool := pgtest.Pool(t)
	dir := t.TempDir()
	files, err := storage.NewLocal(dir, "http://test/uploads")
	if err != nil {
		t.Fatal(err)
	}
	return NewService(NewRepository(postgres.NewTransactionManager(pool)), files), pgtest.NewFixtures(t, pool), dir
}

func newUser(fx *pgtest.Fixtures, role string) uuid.UUID {
	id := uuid.New()
	fx.Exec(`INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, 'x', $3::user_role)`, id, id.String()+"@t.local", role)
	fx.Exec(`INSERT INTO profiles (user_id, first_name, last_name) VALUES ($1, 'Анна', 'Петрова')`, id)
	return id
}

func body(t *testing.T, s string) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func details(t *testing.T, err error) string {
	t.Helper()
	v, ok := IsValidation(err)
	if !ok {
		t.Fatalf("want validation error, got %v", err)
	}
	return strings.Join(v.Details, "\n")
}

func TestStudentOnboarding(t *testing.T) {
	svc, fx, _ := newService(t)
	ctx := context.Background()
	st := newUser(fx, "student")

	// Шаг 1 — сохраняем часть анкеты.
	p, err := svc.Patch(ctx, st, body(t, `{"age": 20, "current_status": "student", "university": " УрФУ ", "course": 3}`))
	if err != nil {
		t.Fatal(err)
	}
	if *p.Age != 20 || *p.University != "УрФУ" || p.OnboardingCompleted {
		t.Fatalf("after step 1: %+v", p)
	}

	// Завершить анкету рано нельзя: в details перечислены пустые обязательные поля.
	_, err = svc.Patch(ctx, st, body(t, `{"onboarding_completed": true}`))
	d := details(t, err)
	for _, f := range []string{"main_goals", "target_skills", "time_limit", "availability_slots", "communication_style"} {
		if !strings.Contains(d, f) {
			t.Fatalf("details %q has no %s", d, f)
		}
	}

	_, err = svc.Patch(ctx, st, body(t, `{"target_skills": ["Кобол"], "main_goals": ["a","b","c"], "time_limit": 9, "is_published": true, "nickname": "x"}`))
	d = details(t, err)
	for _, want := range []string{"Кобол", "main_goals", "time_limit", "is_published", "nickname"} {
		if !strings.Contains(d, want) {
			t.Fatalf("details %q has no %s", d, want)
		}
	}

	p, err = svc.Patch(ctx, st, body(t, `{
		"target_skills": ["go", "SQL"], "current_skills": [{"skill": "python", "level": 1}],
		"main_goals": ["interview_prep"], "time_limit": 4,
		"availability_slots": {"mon": [21, 19, 20], "sat": [10]}, "communication_style": "friendly",
		"min_duration_months": 2, "max_duration_months": 6, "onboarding_completed": true,
		"id": "ignored", "role": "admin"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !p.OnboardingCompleted || p.TargetSkills[0] != "Go" || p.CurrentSkills[0].Skill != "Python" || p.AvailabilitySlots["mon"][0] != 19 || p.Role != user.Student {
		t.Fatalf("completed profile: %+v", p)
	}

	if _, err := svc.Patch(ctx, st, body(t, `{"min_duration_months": 8}`)); !strings.Contains(details(t, err), "min_duration_months") {
		t.Fatal("min must not exceed max")
	}
	if _, err := svc.Patch(ctx, st, body(t, `{"first_name": "  "}`)); !strings.Contains(details(t, err), "first_name") {
		t.Fatal("first name is required")
	}
}

type file struct{ name, content string }

func uploads(files ...file) []UploadedFile {
	out := make([]UploadedFile, len(files))
	for i, f := range files {
		content := f.content
		out[i] = UploadedFile{Name: f.name, Size: int64(len(content)), Open: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBufferString(content)), nil
		}}
	}
	return out
}

func TestMentorVerification(t *testing.T) {
	svc, fx, dir := newService(t)
	ctx := context.Background()
	m := newUser(fx, "mentor")

	_, err := svc.Patch(ctx, m, body(t, `{"current_skills": [{"skill": "Go", "level": 1}], "experience": "none"}`))
	if d := details(t, err); !strings.Contains(d, "уровень") || !strings.Contains(d, "experience") {
		t.Fatalf("mentor rules: %s", d)
	}

	// Публикация без материалов: статус остаётся none.
	p, err := svc.Patch(ctx, m, body(t, `{"is_published": true, "max_students": 3}`))
	if err != nil || p.VerificationStatus != "none" || p.FreeSlots == nil || *p.FreeSlots != 3 {
		t.Fatalf("publish without assets: %+v, %v", p, err)
	}

	if _, err := svc.AddAssets(ctx, m, user.Student, uploads(file{"a.pdf", "x"}), nil); !errors.Is(err, ErrMentorOnly) {
		t.Fatalf("student uploads: %v", err)
	}
	_, err = svc.AddAssets(ctx, m, user.Mentor, uploads(file{"virus.exe", "x"}), []string{"ftp://x"})
	if d := details(t, err); !strings.Contains(d, "virus.exe") || !strings.Contains(d, "ftp://x") {
		t.Fatalf("asset validation: %s", d)
	}

	p, err = svc.AddAssets(ctx, m, user.Mentor, uploads(file{"Диплом.PDF", "diploma"}), []string{"https://github.com/ivan"})
	if err != nil {
		t.Fatal(err)
	}
	if p.VerificationStatus != "pending" || len(p.VerificationAssets) != 2 {
		t.Fatalf("after upload: %+v", p)
	}
	fileAsset := p.VerificationAssets[0]
	saved := filepath.Join(dir, strings.TrimPrefix(fileAsset.URL, "http://test/uploads/"))
	if b, err := os.ReadFile(saved); /* #nosec G304 -- путь из теста */ err != nil || string(b) != "diploma" || !strings.HasSuffix(saved, ".pdf") {
		t.Fatalf("saved file %s: %q, %v", saved, b, err)
	}

	// Куратор проверил, наставник добавил ещё ссылку — снова на проверку.
	fx.Exec(`UPDATE profiles SET verification_status = 'verified' WHERE user_id = $1`, m)
	p, err = svc.AddAssets(ctx, m, user.Mentor, nil, []string{"https://t.me/ivan"})
	if err != nil || p.VerificationStatus != "pending" {
		t.Fatalf("re-verification: %+v, %v", p, err)
	}

	// В профиле 3 материала: ещё 5 можно, а следующие 3 уже превысят лимит 10.
	links := func(n int, prefix string) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = "https://example.com/" + prefix + string(rune('a'+i))
		}
		return out
	}
	if _, err := svc.AddAssets(ctx, m, user.Mentor, nil, links(5, "x")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddAssets(ctx, m, user.Mentor, nil, links(3, "y")); !strings.Contains(details(t, err), "10") {
		t.Fatalf("assets limit: %v", err)
	}

	p, err = svc.DeleteAsset(ctx, m, user.Mentor, fileAsset.ID)
	if err != nil || len(p.VerificationAssets) != 7 {
		t.Fatalf("delete: %+v, %v", p.VerificationAssets, err)
	}
	if _, err := os.Stat(saved); !os.IsNotExist(err) {
		t.Fatal("file must be removed from disk")
	}
	if _, err := svc.DeleteAsset(ctx, m, user.Mentor, fileAsset.ID); !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("delete twice: %v", err)
	}
}
