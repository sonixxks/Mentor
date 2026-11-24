package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// mentorsCTE — наставники с рейтингом и свободными местами.
// $1 — id наставника, которого надо показать независимо от публикации (владелец карточки), или NULL.
const mentorsCTE = `
	WITH mentors AS (
		SELECT p.user_id AS id, p.first_name, p.last_name, p.avatar_url, p.job_title, p.company,
		       p.experience::text AS experience, p.bio, p.hourly_rate::float8 AS hourly_rate,
		       p.verification_status::text AS verification_status,
		       COALESCE(p.current_skills, '[]') AS current_skills,
		       p.time_limit, p.communication_style::text AS communication_style,
		       p.help_topics, p.target_student_levels, p.availability_slots, p.max_students,
		       p.is_published,
		       rv.rating_avg, rv.reviews_count,
		       GREATEST(COALESCE(p.max_students, 0) - occ.occupied, 0)::int AS free_slots
		FROM profiles p
		JOIN users u ON u.id = p.user_id AND u.role = 'mentor'
		LEFT JOIN LATERAL (
			SELECT round(avg(sr.rating), 2)::float8 AS rating_avg, count(*)::int AS reviews_count
			FROM session_reviews sr
			WHERE sr.target_id = p.user_id
		) rv ON TRUE
		LEFT JOIN LATERAL (
			SELECT count(*)::int AS occupied
			FROM workspaces w
			JOIN requests r ON r.id = w.request_id
			WHERE r.mentor_id = p.user_id AND w.status <> 'completed'
		) occ ON TRUE
		WHERE (p.is_published AND p.verification_status = 'verified') OR p.user_id = $1
	)`

const summaryColumns = `m.id, m.first_name, m.last_name, m.avatar_url, m.job_title, m.company,
	m.experience, m.bio, m.hourly_rate, m.verification_status, m.current_skills,
	m.rating_avg, m.reviews_count, m.free_slots, m.time_limit, m.communication_style`

// В списке владелец не подмешивается ($1 = NULL), поэтому каталог — только опубликованные и верифицированные.
const listWhere = `
	WHERE ($2::text IS NULL
	       OR m.first_name ILIKE $2 OR m.last_name ILIKE $2
	       OR m.company ILIKE $2 OR m.job_title ILIKE $2
	       OR EXISTS (SELECT 1 FROM jsonb_array_elements(m.current_skills) e WHERE e->>'skill' ILIKE $2))
	  AND ($3::text[] IS NULL
	       OR EXISTS (SELECT 1 FROM jsonb_array_elements(m.current_skills) e WHERE lower(e->>'skill') = ANY ($3)))
	  AND ($4::float8 IS NULL OR m.hourly_rate >= $4)
	  AND ($5::float8 IS NULL OR m.hourly_rate <= $5)
	  AND ($6::int IS NULL OR m.time_limit >= $6)
	  AND ($7::int IS NULL OR m.time_limit <= $7)
	  AND ($8::float8 IS NULL OR m.rating_avg >= $8)
	  AND (NOT $9::bool OR m.free_slots > 0)`

var orderBy = map[Sort]string{
	SortRating:    `m.rating_avg DESC NULLS LAST, m.reviews_count DESC, m.id`,
	SortPriceAsc:  `m.hourly_rate ASC NULLS LAST, m.rating_avg DESC NULLS LAST, m.id`,
	SortPriceDesc: `m.hourly_rate DESC NULLS LAST, m.rating_avg DESC NULLS LAST, m.id`,
	SortExperience: `CASE m.experience
		WHEN 'over_10' THEN 4 WHEN '5_to_10' THEN 3 WHEN '3_to_5' THEN 2 WHEN '1_to_3' THEN 1 ELSE 0
	END DESC, m.rating_avg DESC NULLS LAST, m.id`,
}

const (
	detailSQL = mentorsCTE + `
		SELECT ` + summaryColumns + `,
		       m.help_topics, m.target_student_levels, m.availability_slots, m.max_students, m.is_published
		FROM mentors m WHERE m.id = $2`

	packagesSQL = `
		SELECT sp.id, sp.title, sp.description, sp.result, sp.price::float8, sp.session_count,
		       sp.skills, sp.status::text,
		       (SELECT count(*) FROM requests r JOIN workspaces w ON w.request_id = r.id
		        WHERE r.service_package_id = sp.id)::int
		FROM service_packages sp
		WHERE sp.mentor_id = $1 AND sp.status = 'published'
		ORDER BY sp.created_at`

	mentorExistsSQL = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND role = 'mentor')`

	reviewsCountSQL = `
		SELECT count(*) FROM session_reviews sr
		JOIN users a ON a.id = sr.author_id AND a.role = 'student'
		WHERE sr.target_id = $1`

	reviewsSQL = `
		SELECT sr.id, sr.rating, sr.comment, COALESCE(ap.first_name, ''), mt.start_time, sr.created_at
		FROM session_reviews sr
		JOIN users a ON a.id = sr.author_id AND a.role = 'student'
		LEFT JOIN profiles ap ON ap.user_id = sr.author_id
		LEFT JOIN meetings mt ON mt.id = sr.meeting_id
		WHERE sr.target_id = $1
		ORDER BY sr.created_at DESC, sr.id
		LIMIT $2 OFFSET $3`

	// Сначала совпадения с начала названия, затем популярные.
	skillsSQL = `
		SELECT name FROM skills
		WHERE ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
		  AND (NOT $2::bool OR is_popular)
		ORDER BY ($1::text IS NOT NULL AND name ILIKE $1 || '%') DESC, is_popular DESC, name
		LIMIT $3`
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListMentors(ctx context.Context, f Filter) ([]MentorSummary, int, error) {
	order, ok := orderBy[f.Sort]
	if !ok {
		order = orderBy[SortRating]
	}

	var query *string
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + escapeLike(q) + "%"
		query = &like
	}

	var skills []string
	for _, s := range f.Skills {
		skills = append(skills, strings.ToLower(s))
	}

	args := []any{nil, query, skills, f.PriceMin, f.PriceMax, f.HoursMin, f.HoursMax, f.MinRating, f.HasFreeSlots}

	var total int
	if err := r.pool.QueryRow(ctx, mentorsCTE+` SELECT count(*) FROM mentors m `+listWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count mentors: %w", err)
	}

	var limit *int
	if f.Limit > 0 {
		limit = &f.Limit
	}
	listSQL := mentorsCTE + ` SELECT ` + summaryColumns + ` FROM mentors m ` + listWhere +
		` ORDER BY ` + order + ` LIMIT $10 OFFSET $11`

	rows, err := r.pool.Query(ctx, listSQL, append(args, limit, f.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list mentors: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (MentorSummary, error) {
		var s MentorSummary
		var skillsRaw []byte
		if err := row.Scan(summaryDest(&s, &skillsRaw)...); err != nil {
			return MentorSummary{}, err
		}
		return s, decodeJSON(skillsRaw, &s.CurrentSkills)
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list mentors: %w", err)
	}
	return list, total, nil
}

// GetMentor возвращает карточку: опубликованную и верифицированную — всем,
// любую свою — владельцу (кнопка «Как видят ученики»).
func (r *PostgresRepository) GetMentor(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (MentorDetail, error) {
	// Владелец передаётся в CTE как $1: его профиль виден, даже если скрыт.
	var owner *uuid.UUID
	if viewerID != nil && *viewerID == id {
		owner = &id
	}

	var (
		d                                         MentorDetail
		skillsRaw, topicsRaw, levelsRaw, slotsRaw []byte
	)
	dest := append(summaryDest(&d.MentorSummary, &skillsRaw), &topicsRaw, &levelsRaw, &slotsRaw, &d.MaxStudents, &d.IsPublished)

	err := r.pool.QueryRow(ctx, detailSQL, owner, id).Scan(dest...)
	if errors.Is(err, pgx.ErrNoRows) {
		return MentorDetail{}, ErrMentorNotFound
	}
	if err != nil {
		return MentorDetail{}, fmt.Errorf("get mentor: %w", err)
	}

	for _, item := range []struct {
		raw []byte
		dst any
	}{{skillsRaw, &d.CurrentSkills}, {topicsRaw, &d.HelpTopics}, {levelsRaw, &d.TargetStudentLevels}} {
		if err := decodeJSON(item.raw, item.dst); err != nil {
			return MentorDetail{}, fmt.Errorf("decode mentor profile: %w", err)
		}
	}
	if len(slotsRaw) > 0 {
		d.AvailabilitySlots = json.RawMessage(slotsRaw)
	}

	d.ServicePackages, err = r.packages(ctx, id)
	if err != nil {
		return MentorDetail{}, err
	}
	return d, nil
}

func (r *PostgresRepository) packages(ctx context.Context, mentorID uuid.UUID) ([]ServicePackage, error) {
	rows, err := r.pool.Query(ctx, packagesSQL, mentorID)
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ServicePackage, error) {
		var (
			p         ServicePackage
			skillsRaw []byte
		)
		err := row.Scan(&p.ID, &p.Title, &p.Description, &p.Result, &p.Price, &p.SessionCount,
			&skillsRaw, &p.Status, &p.PairsCount)
		if err != nil {
			return ServicePackage{}, err
		}
		return p, decodeJSON(skillsRaw, &p.Skills)
	})
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	return list, nil
}

func (r *PostgresRepository) ListReviews(ctx context.Context, mentorID uuid.UUID, limit, offset int) ([]Review, int, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, mentorExistsSQL, mentorID).Scan(&exists); err != nil {
		return nil, 0, fmt.Errorf("check mentor: %w", err)
	}
	if !exists {
		return nil, 0, ErrMentorNotFound
	}

	var total int
	if err := r.pool.QueryRow(ctx, reviewsCountSQL, mentorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	rows, err := r.pool.Query(ctx, reviewsSQL, mentorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Review, error) {
		var rv Review
		err := row.Scan(&rv.ID, &rv.Rating, &rv.Comment, &rv.AuthorFirstName, &rv.MeetingDate, &rv.CreatedAt)
		return rv, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	return list, total, nil
}

func (r *PostgresRepository) ListSkills(ctx context.Context, q SkillQuery) ([]string, error) {
	var query *string
	if s := strings.TrimSpace(q.Query); s != "" {
		escaped := escapeLike(s)
		query = &escaped
	}

	rows, err := r.pool.Query(ctx, skillsSQL, query, q.Popular, q.Limit)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return names, nil
}

func summaryDest(s *MentorSummary, skillsRaw *[]byte) []any {
	return []any{
		&s.ID, &s.FirstName, &s.LastName, &s.AvatarURL, &s.JobTitle, &s.Company,
		&s.Experience, &s.Bio, &s.HourlyRate, &s.VerificationStatus, skillsRaw,
		&s.RatingAvg, &s.ReviewsCount, &s.FreeSlots, &s.TimeLimit, &s.CommunicationStyle,
	}
}

func decodeJSON(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

// escapeLike экранирует спецсимволы LIKE, чтобы «100%» искалось буквально.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
