-- +goose Up
-- Программы наставника: черновик -> опубликована -> архив (мягкое удаление,
-- потому что на программу ссылаются заявки).
CREATE TYPE package_status AS ENUM ('draft', 'published', 'archived');

ALTER TABLE service_packages
    ALTER COLUMN description DROP NOT NULL,
    ADD COLUMN skills JSONB,
    ADD COLUMN price  NUMERIC(10, 2) CHECK (price >= 0),
    ADD COLUMN result TEXT,
    ADD COLUMN status package_status NOT NULL DEFAULT 'draft',
    ADD CONSTRAINT service_packages_skills_is_array CHECK (skills IS NULL OR jsonb_typeof(skills) = 'array');

UPDATE service_packages
SET status = CASE WHEN is_active THEN 'published'::package_status ELSE 'archived'::package_status END;

ALTER TABLE service_packages DROP COLUMN is_active;

-- Новый статус заявки: наставник одобрил, ждём подтверждения ученика.
-- Используется в следующей миграции: новое значение ENUM нельзя применять
-- в той же транзакции, где оно добавлено.
ALTER TYPE request_status ADD VALUE IF NOT EXISTS 'approved_by_mentor' AFTER 'pending';

-- +goose Down
ALTER TABLE service_packages ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
UPDATE service_packages SET is_active = (status = 'published');

ALTER TABLE service_packages
    DROP CONSTRAINT IF EXISTS service_packages_skills_is_array,
    DROP COLUMN status,
    DROP COLUMN result,
    DROP COLUMN price,
    DROP COLUMN skills;

UPDATE service_packages SET description = '' WHERE description IS NULL;
ALTER TABLE service_packages ALTER COLUMN description SET NOT NULL;

DROP TYPE IF EXISTS package_status;
