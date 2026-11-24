-- +goose Up
-- Статус пары ведёт куратор. active / paused / stagnation занимают место
-- наставника, completed освобождает.
CREATE TYPE workspace_status AS ENUM ('active', 'paused', 'stagnation', 'completed');

ALTER TABLE workspaces
    ADD COLUMN status workspace_status NOT NULL DEFAULT 'active',
    -- Расписание и заметки переехали в meetings и meeting_notes.
    DROP COLUMN next_session_time,
    DROP COLUMN next_session_topic,
    DROP COLUMN next_session_duration,
    DROP COLUMN whiteboard_url,
    DROP COLUMN meeting_notes;

CREATE INDEX idx_workspaces_status ON workspaces (status);

-- Новый жизненный цикл задачи: todo -> in_progress -> on_review -> completed,
-- с возвратом на доработку (needs_revision).
CREATE TYPE task_status_v2 AS ENUM ('todo', 'in_progress', 'on_review', 'needs_revision', 'completed');

ALTER TABLE tasks ALTER COLUMN status DROP DEFAULT;
ALTER TABLE tasks ALTER COLUMN status TYPE task_status_v2 USING (
    CASE status::text
        WHEN 'review' THEN 'on_review'
        WHEN 'done' THEN 'completed'
        ELSE status::text
    END
)::task_status_v2;
ALTER TABLE tasks ALTER COLUMN status SET DEFAULT 'todo';

DROP TYPE task_status;
ALTER TYPE task_status_v2 RENAME TO task_status;

-- Просроченные задачи для «светофора» куратора.
CREATE INDEX idx_tasks_workspace_deadline ON tasks (workspace_id, deadline_at) WHERE status <> 'completed';

-- В базе знаний теперь и файлы, и ссылки.
CREATE TYPE knowledge_item_type AS ENUM ('file', 'link');

ALTER TABLE knowledge_base
    ADD COLUMN type knowledge_item_type NOT NULL DEFAULT 'file',
    ALTER COLUMN file_size_kb DROP NOT NULL,
    ADD CONSTRAINT knowledge_base_file_has_size CHECK (type = 'link' OR file_size_kb IS NOT NULL);

-- +goose Down
ALTER TABLE knowledge_base DROP CONSTRAINT knowledge_base_file_has_size;
DELETE FROM knowledge_base WHERE type = 'link';
ALTER TABLE knowledge_base
    ALTER COLUMN file_size_kb SET NOT NULL,
    DROP COLUMN type;
DROP TYPE IF EXISTS knowledge_item_type;

DROP INDEX IF EXISTS idx_tasks_workspace_deadline;

CREATE TYPE task_status_v1 AS ENUM ('todo', 'in_progress', 'review', 'done');
ALTER TABLE tasks ALTER COLUMN status DROP DEFAULT;
ALTER TABLE tasks ALTER COLUMN status TYPE task_status_v1 USING (
    CASE status::text
        WHEN 'on_review' THEN 'review'
        WHEN 'needs_revision' THEN 'in_progress'
        WHEN 'completed' THEN 'done'
        ELSE status::text
    END
)::task_status_v1;
ALTER TABLE tasks ALTER COLUMN status SET DEFAULT 'todo';
DROP TYPE task_status;
ALTER TYPE task_status_v1 RENAME TO task_status;

DROP INDEX IF EXISTS idx_workspaces_status;
ALTER TABLE workspaces
    DROP COLUMN status,
    ADD COLUMN next_session_time     TIMESTAMPTZ,
    ADD COLUMN next_session_topic    VARCHAR(255),
    ADD COLUMN next_session_duration INT CHECK (next_session_duration > 0),
    ADD COLUMN whiteboard_url        VARCHAR(255),
    ADD COLUMN meeting_notes         TEXT;
DROP TYPE IF EXISTS workspace_status;
