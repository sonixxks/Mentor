-- +goose Up
-- Встречи пары. Ссылка на видеозвонок постоянная и хранится в workspaces.jitsi_link.
CREATE TABLE meetings (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID         NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL DEFAULT 'Встреча',
    start_time   TIMESTAMPTZ  NOT NULL,
    end_time     TIMESTAMPTZ,

    CONSTRAINT meetings_end_after_start CHECK (end_time IS NULL OR end_time > start_time)
);

CREATE INDEX idx_meetings_workspace_start ON meetings (workspace_id, start_time);

-- Общие markdown-заметки пары (побеждает последнее сохранение).
CREATE TABLE meeting_notes (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID         NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    author_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL,
    content      TEXT         NOT NULL DEFAULT '',
    meeting_date TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_meeting_notes_workspace_id ON meeting_notes (workspace_id, meeting_date DESC);

-- Оценка встречи: каждый участник пары оценивает партнёра один раз за встречу.
-- Отзывов в старом формате (без автора) нет — API их не создавал.
DELETE FROM session_reviews;

ALTER TABLE session_reviews
    ADD COLUMN author_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    ADD COLUMN target_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    ADD COLUMN meeting_id UUID NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    ADD CONSTRAINT session_reviews_author_not_target CHECK (author_id <> target_id),
    -- 409 REVIEW_EXISTS
    ADD CONSTRAINT uq_session_reviews_meeting_author UNIQUE (meeting_id, author_id);

-- Рейтинг наставника в каталоге и последняя оценка для сигнала bad_reviews.
CREATE INDEX idx_session_reviews_target_id ON session_reviews (target_id);
CREATE INDEX idx_session_reviews_workspace_created ON session_reviews (workspace_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_session_reviews_workspace_created;
DROP INDEX IF EXISTS idx_session_reviews_target_id;

ALTER TABLE session_reviews
    DROP CONSTRAINT uq_session_reviews_meeting_author,
    DROP CONSTRAINT session_reviews_author_not_target,
    DROP COLUMN meeting_id,
    DROP COLUMN target_id,
    DROP COLUMN author_id;

DROP TABLE IF EXISTS meeting_notes;
DROP TABLE IF EXISTS meetings;
