-- +goose Up
-- Management by Exception: статус меняет наставник (accepted / rejected_by_mentor)
-- или студент (cancelled); куратор вмешивается только при аномалиях (rejected_by_curator).
CREATE TYPE request_status AS ENUM (
    'pending',
    'accepted',
    'rejected_by_mentor',
    'rejected_by_curator',
    'cancelled'
);

CREATE TABLE service_packages (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    mentor_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title         VARCHAR(255) NOT NULL,
    description   TEXT         NOT NULL,
    session_count INT          CHECK (session_count > 0),
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_service_packages_mentor_id ON service_packages (mentor_id);

CREATE TABLE requests (
    id                 UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id         UUID           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    mentor_id          UUID           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    service_package_id UUID           REFERENCES service_packages (id) ON DELETE SET NULL,
    status             request_status NOT NULL DEFAULT 'pending',
    curator_comment    TEXT,
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT requests_student_not_mentor CHECK (student_id <> mentor_id),
    -- Принудительное отклонение куратором всегда с объяснением.
    CONSTRAINT requests_curator_reject_has_comment CHECK (
        status <> 'rejected_by_curator'
        OR (curator_comment IS NOT NULL AND btrim(curator_comment) <> '')
    )
);

CREATE INDEX idx_requests_student_id ON requests (student_id);
CREATE INDEX idx_requests_mentor_id ON requests (mentor_id);

-- Не больше одной ожидающей заявки от ученика к одному наставнику (защита от спама).
-- После cancelled / rejected_* можно подать заявку этому наставнику снова.
CREATE UNIQUE INDEX uq_requests_pending_pair
    ON requests (student_id, mentor_id)
    WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS requests;
DROP TABLE IF EXISTS service_packages;

DROP TYPE IF EXISTS request_status;
