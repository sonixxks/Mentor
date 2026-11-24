-- +goose Up
-- Чат внутри пары. system — служебные строки («Встреча назначена»), у них нет отправителя.
CREATE TYPE message_type AS ENUM ('text', 'system');

CREATE TABLE messages (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID         NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    sender_id    UUID         REFERENCES users (id) ON DELETE SET NULL,
    type         message_type NOT NULL DEFAULT 'text',
    text         TEXT         CHECK (char_length(text) <= 4000),
    file_id      UUID         REFERENCES knowledge_base (id) ON DELETE SET NULL,
    is_read      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT messages_has_content CHECK (text IS NOT NULL OR file_id IS NOT NULL),
    CONSTRAINT messages_text_has_sender CHECK (type = 'system' OR sender_id IS NOT NULL)
);

-- История чата постранично назад по времени.
CREATE INDEX idx_messages_workspace_created ON messages (workspace_id, created_at DESC);
-- Счётчик непрочитанных в списке пар.
CREATE INDEX idx_messages_unread ON messages (workspace_id, sender_id) WHERE NOT is_read;

CREATE TYPE notification_type AS ENUM (
    'request_update',
    'task_review',
    'new_message',
    'new_recommendations',
    'meeting_update',
    'system_alert'
);

CREATE TABLE notifications (
    id         UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID              NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       notification_type NOT NULL,
    title      VARCHAR(255)      NOT NULL,
    message    TEXT,
    -- Полиморфная ссылка: заявка, задача, сообщение или воркспейс в зависимости от type.
    entity_id  UUID,
    is_read    BOOLEAN           NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications (user_id, type) WHERE NOT is_read;

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;

DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS message_type;
