-- +goose Up
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'review', 'done');

CREATE TABLE workspaces (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id            UUID         NOT NULL UNIQUE REFERENCES requests (id) ON DELETE CASCADE,
    next_session_time     TIMESTAMPTZ,
    next_session_topic    VARCHAR(255),
    next_session_duration INT          CHECK (next_session_duration > 0),
    jitsi_link            VARCHAR(255) NOT NULL,
    whiteboard_url        VARCHAR(255),
    meeting_notes         TEXT,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE tasks (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id   UUID         NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    status         task_status  NOT NULL DEFAULT 'todo',
    artifacts_url  VARCHAR(255),
    mentor_comment TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deadline_at    TIMESTAMPTZ  NOT NULL,
    completed_at   TIMESTAMPTZ
);

CREATE INDEX idx_tasks_workspace_id ON tasks (workspace_id);

CREATE TABLE task_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    UUID        NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    author_id  UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    text       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_comments_task_id ON task_comments (task_id);

CREATE TABLE knowledge_base (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID         NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    file_name    VARCHAR(255) NOT NULL,
    file_url     VARCHAR(255) NOT NULL,
    file_size_kb INT          NOT NULL CHECK (file_size_kb >= 0),
    uploaded_by  UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_knowledge_base_workspace_id ON knowledge_base (workspace_id);

CREATE TABLE session_reviews (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID        NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    rating       INT         NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_session_reviews_workspace_id ON session_reviews (workspace_id);

-- +goose Down
DROP TABLE IF EXISTS session_reviews;
DROP TABLE IF EXISTS knowledge_base;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS workspaces;

DROP TYPE IF EXISTS task_status;
