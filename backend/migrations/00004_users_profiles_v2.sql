-- +goose Up
-- Приводим users и profiles к ER-диаграмме от 16.11.

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'admin';

ALTER TYPE experience_level ADD VALUE IF NOT EXISTS '3_to_5';
ALTER TYPE experience_level ADD VALUE IF NOT EXISTS '5_to_10';
ALTER TYPE experience_level ADD VALUE IF NOT EXISTS 'over_10';

CREATE TYPE verification_status AS ENUM ('none', 'pending', 'verified', 'rejected');

-- Пароля может не быть: вход через Яндекс ID, куратор до принятия приглашения.
ALTER TABLE users
    ALTER COLUMN password_hash DROP NOT NULL,
    ADD COLUMN yandex_id           VARCHAR(255) UNIQUE,
    ADD COLUMN consent_at          TIMESTAMPTZ,
    ADD COLUMN password_changed_at TIMESTAMPTZ;

-- У профиля свой id (по нему работают ручки куратора), user_id остаётся уникальным.
ALTER TABLE profiles DROP CONSTRAINT profiles_pkey;
ALTER TABLE profiles
    ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD CONSTRAINT profiles_pkey PRIMARY KEY (id),
    ADD CONSTRAINT profiles_user_id_key UNIQUE (user_id);

-- Анкета заполняется по шагам, поэтому почти все поля необязательные.
ALTER TABLE profiles
    ALTER COLUMN age DROP NOT NULL,
    ALTER COLUMN current_status DROP NOT NULL,
    ALTER COLUMN university DROP NOT NULL,
    ALTER COLUMN major DROP NOT NULL,
    ALTER COLUMN current_skills DROP NOT NULL,
    ALTER COLUMN main_goals DROP NOT NULL,
    ALTER COLUMN time_limit DROP NOT NULL,
    ALTER COLUMN availability_slots DROP NOT NULL,
    ALTER COLUMN communication_style DROP NOT NULL,
    DROP COLUMN duration_months;

ALTER TABLE profiles
    ADD COLUMN middle_name              VARCHAR(255),
    ADD COLUMN course                   INT CHECK (course BETWEEN 1 AND 6),
    ADD COLUMN min_duration_months      INT CHECK (min_duration_months > 0),
    ADD COLUMN max_duration_months      INT CHECK (max_duration_months BETWEEN 1 AND 12),
    ADD COLUMN job_title                VARCHAR(255),
    ADD COLUMN company                  VARCHAR(255),
    ADD COLUMN verification_assets      JSONB,
    ADD COLUMN hourly_rate              NUMERIC(10, 2) CHECK (hourly_rate >= 0),
    ADD COLUMN max_students             INT CHECK (max_students BETWEEN 1 AND 10),
    ADD COLUMN target_student_levels    JSONB,
    ADD COLUMN help_topics              JSONB,
    ADD COLUMN notification_preferences JSONB,
    ADD COLUMN is_published             BOOLEAN             NOT NULL DEFAULT FALSE,
    ADD COLUMN needs_curator_help       BOOLEAN             NOT NULL DEFAULT FALSE,
    ADD COLUMN onboarding_completed     BOOLEAN             NOT NULL DEFAULT FALSE,
    ADD COLUMN verification_status      verification_status NOT NULL DEFAULT 'none',
    ADD CONSTRAINT profiles_duration_range CHECK (
        min_duration_months IS NULL OR max_duration_months IS NULL
        OR min_duration_months <= max_duration_months
    ),
    ADD CONSTRAINT profiles_time_limit_range CHECK (time_limit IS NULL OR time_limit <= 10);

-- Каталог показывает только опубликованных верифицированных наставников.
CREATE INDEX idx_profiles_catalog ON profiles (user_id)
    WHERE is_published AND verification_status = 'verified';

-- Фильтр по навыкам: current_skills @> '[{"skill": "Go"}]'.
CREATE INDEX idx_profiles_current_skills ON profiles USING GIN (current_skills jsonb_path_ops);

-- Очереди куратора: запросы помощи и верификация.
CREATE INDEX idx_profiles_needs_help ON profiles (user_id) WHERE needs_curator_help;
CREATE INDEX idx_profiles_verification_pending ON profiles (user_id) WHERE verification_status = 'pending';

-- +goose Down
-- Значения, добавленные в ENUM (admin, 3_to_5, ...), PostgreSQL удалить не позволяет — они остаются.
DROP INDEX IF EXISTS idx_profiles_verification_pending;
DROP INDEX IF EXISTS idx_profiles_needs_help;
DROP INDEX IF EXISTS idx_profiles_current_skills;
DROP INDEX IF EXISTS idx_profiles_catalog;

ALTER TABLE profiles
    DROP CONSTRAINT IF EXISTS profiles_time_limit_range,
    DROP CONSTRAINT IF EXISTS profiles_duration_range,
    DROP COLUMN verification_status,
    DROP COLUMN onboarding_completed,
    DROP COLUMN needs_curator_help,
    DROP COLUMN is_published,
    DROP COLUMN notification_preferences,
    DROP COLUMN help_topics,
    DROP COLUMN target_student_levels,
    DROP COLUMN max_students,
    DROP COLUMN hourly_rate,
    DROP COLUMN verification_assets,
    DROP COLUMN company,
    DROP COLUMN job_title,
    DROP COLUMN max_duration_months,
    DROP COLUMN min_duration_months,
    DROP COLUMN course,
    DROP COLUMN middle_name,
    ADD COLUMN duration_months INT CHECK (duration_months > 0);

ALTER TABLE profiles DROP CONSTRAINT profiles_pkey;
ALTER TABLE profiles DROP CONSTRAINT profiles_user_id_key;
ALTER TABLE profiles DROP COLUMN id;
ALTER TABLE profiles ADD CONSTRAINT profiles_pkey PRIMARY KEY (user_id);

ALTER TABLE users
    DROP COLUMN password_changed_at,
    DROP COLUMN consent_at,
    DROP COLUMN yandex_id;

DROP TYPE IF EXISTS verification_status;
