-- +goose Up
-- Матрица переходов v1.2:
--   ученик:    pending -> cancelled; approved_by_mentor -> accepted
--   наставник: pending -> approved_by_mentor / rejected_by_mentor
--   куратор:   любой нетерминальный -> rejected_by_curator (с причиной)
ALTER TABLE requests RENAME COLUMN curator_comment TO rejection_reason;

ALTER TABLE requests
    ADD COLUMN cover_letter TEXT CHECK (char_length(cover_letter) <= 1000),
    DROP CONSTRAINT requests_curator_reject_has_comment,
    ADD CONSTRAINT requests_curator_reject_has_reason CHECK (
        status <> 'rejected_by_curator'
        OR (rejection_reason IS NOT NULL AND btrim(rejection_reason) <> '')
    );

-- Не больше одной активной заявки ученика к одному наставнику (409 DUPLICATE_REQUEST).
DROP INDEX uq_requests_pending_pair;
CREATE UNIQUE INDEX uq_requests_active_pair
    ON requests (student_id, mentor_id)
    WHERE status IN ('pending', 'approved_by_mentor');

-- Очередь заявок куратора с фильтром по статусу.
CREATE INDEX idx_requests_status_created_at ON requests (status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_requests_status_created_at;
DROP INDEX IF EXISTS uq_requests_active_pair;

-- approved_by_mentor в старой схеме не было: возвращаем такие заявки в pending.
UPDATE requests SET status = 'pending' WHERE status = 'approved_by_mentor';

CREATE UNIQUE INDEX uq_requests_pending_pair
    ON requests (student_id, mentor_id)
    WHERE status = 'pending';

ALTER TABLE requests
    DROP CONSTRAINT requests_curator_reject_has_reason,
    DROP COLUMN cover_letter;

ALTER TABLE requests RENAME COLUMN rejection_reason TO curator_comment;

ALTER TABLE requests ADD CONSTRAINT requests_curator_reject_has_comment CHECK (
    status <> 'rejected_by_curator'
    OR (curator_comment IS NOT NULL AND btrim(curator_comment) <> '')
);
