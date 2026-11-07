-- +goose Up
CREATE TYPE user_role AS ENUM ('student', 'mentor', 'curator');
CREATE TYPE current_status AS ENUM ('student', 'graduate', 'junior', 'switcher');
CREATE TYPE experience_level AS ENUM ('none', 'under_1', '1_to_3', 'over_3');
CREATE TYPE communication_style AS ENUM ('friendly', 'structured', 'free');

CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          user_role    NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Профиль 1:1 с пользователем, поэтому user_id сразу и PK.
CREATE TABLE profiles (
    user_id             UUID                PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    first_name          VARCHAR(255)        NOT NULL,
    last_name           VARCHAR(255)        NOT NULL,
    avatar_url          VARCHAR(255),
    age                 INT                 NOT NULL CHECK (age > 0),
    bio                 TEXT,
    current_status      current_status      NOT NULL,
    university          VARCHAR(255)        NOT NULL,
    major               VARCHAR(255)        NOT NULL,
    experience          experience_level,
    target_skills       JSONB,
    current_skills      JSONB               NOT NULL DEFAULT '[]',
    main_goals          JSONB               NOT NULL DEFAULT '[]',
    duration_months     INT                 NOT NULL CHECK (duration_months > 0),
    success_metric      TEXT,
    time_limit          INT                 NOT NULL CHECK (time_limit > 0),
    availability_slots  JSONB               NOT NULL DEFAULT '[]',
    communication_style communication_style NOT NULL,

    -- Навыки и цели — массивы слагов из закрытого справочника; сами слаги валидирует код.
    CONSTRAINT profiles_target_skills_is_array CHECK (target_skills IS NULL OR jsonb_typeof(target_skills) = 'array'),
    CONSTRAINT profiles_current_skills_is_array CHECK (jsonb_typeof(current_skills) = 'array'),
    CONSTRAINT profiles_main_goals_is_array CHECK (jsonb_typeof(main_goals) = 'array')
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS communication_style;
DROP TYPE IF EXISTS experience_level;
DROP TYPE IF EXISTS current_status;
DROP TYPE IF EXISTS user_role;
