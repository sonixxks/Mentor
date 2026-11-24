package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
)

const messageColumns = `m.id, m.workspace_id, m.sender_id, m.type::text, m.text, m.is_read, m.created_at,
	f.id, f.type::text, f.file_name, f.file_url, f.file_size_kb, f.uploaded_by, f.created_at`

const (
	membersSQL = `
		SELECT r.student_id, r.mentor_id
		FROM workspaces w JOIN requests r ON r.id = w.request_id
		WHERE w.id = $1`

	// Берём последние limit сообщений до before и разворачиваем по возрастанию времени.
	listSQL = `
		SELECT * FROM (
			SELECT ` + messageColumns + `
			FROM messages m
			LEFT JOIN knowledge_base f ON f.id = m.file_id
			WHERE m.workspace_id = $1 AND ($2::timestamptz IS NULL OR m.created_at < $2)
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $3
		) page
		ORDER BY 7, 1`

	fileInWorkspaceSQL = `SELECT EXISTS (SELECT 1 FROM knowledge_base WHERE id = $1 AND workspace_id = $2)`

	insertSQL = `
		WITH m AS (
			INSERT INTO messages (workspace_id, sender_id, type, text, file_id)
			VALUES ($1, $2, $3::message_type, $4, $5)
			RETURNING *
		)
		SELECT ` + messageColumns + ` FROM m LEFT JOIN knowledge_base f ON f.id = m.file_id`

	// Одно непрочитанное уведомление на пару: пока партнёр не открыл чат, новых не создаём.
	hasUnreadNoticeSQL = `
		SELECT EXISTS (
			SELECT 1 FROM notifications n
			JOIN messages msg ON msg.id = n.entity_id
			WHERE n.user_id = $1 AND n.type = 'new_message' AND NOT n.is_read AND msg.workspace_id = $2
		)`

	// Прочитанными становятся сообщения партнёра и системные.
	markReadSQL = `
		UPDATE messages SET is_read = TRUE
		WHERE workspace_id = $1 AND sender_id IS DISTINCT FROM $2 AND NOT is_read`

	readNoticesSQL = `
		UPDATE notifications n SET is_read = TRUE
		FROM messages msg
		WHERE msg.id = n.entity_id AND msg.workspace_id = $2
		  AND n.user_id = $1 AND n.type = 'new_message' AND NOT n.is_read`
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Members(ctx context.Context, workspaceID uuid.UUID) (Members, error) {
	var m Members
	err := r.pool.QueryRow(ctx, membersSQL, workspaceID).Scan(&m.StudentID, &m.MentorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Members{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Members{}, fmt.Errorf("workspace members: %w", err)
	}
	return m, nil
}

func (r *PostgresRepository) List(ctx context.Context, workspaceID uuid.UUID, before *time.Time, limit int) ([]Message, error) {
	rows, err := r.pool.Query(ctx, listSQL, workspaceID, before, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Message, error) {
		return scanMessage(row)
	})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return list, nil
}

// Create сохраняет сообщение и, если у партнёра ещё нет непрочитанного
// уведомления по этой паре, создаёт уведомление new_message.
func (r *PostgresRepository) Create(
	ctx context.Context,
	workspaceID, senderID uuid.UUID,
	text *string,
	fileID *uuid.UUID,
	notify func(Message) notifications.Notification,
) (Message, error) {
	var msg Message

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if fileID != nil {
			var ok bool
			if err := tx.QueryRow(ctx, fileInWorkspaceSQL, *fileID, workspaceID).Scan(&ok); err != nil {
				return fmt.Errorf("check file: %w", err)
			}
			if !ok {
				return ErrFileNotFound
			}
		}

		var err error
		msg, err = scanMessage(tx.QueryRow(ctx, insertSQL, workspaceID, senderID, string(TypeText), text, fileID))
		if err != nil {
			return fmt.Errorf("insert message: %w", err)
		}

		n := notify(msg)
		var hasUnread bool
		if err := tx.QueryRow(ctx, hasUnreadNoticeSQL, n.UserID, workspaceID).Scan(&hasUnread); err != nil {
			return fmt.Errorf("check unread notification: %w", err)
		}
		if hasUnread {
			return nil
		}
		return notifications.Insert(ctx, tx, n)
	})
	if err != nil {
		return Message{}, err
	}
	return msg, nil
}

// MarkRead отмечает прочитанными сообщения партнёра и уведомления о них.
func (r *PostgresRepository) MarkRead(ctx context.Context, workspaceID, readerID uuid.UUID) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, markReadSQL, workspaceID, readerID); err != nil {
			return fmt.Errorf("mark messages read: %w", err)
		}
		if _, err := tx.Exec(ctx, readNoticesSQL, readerID, workspaceID); err != nil {
			return fmt.Errorf("mark notifications read: %w", err)
		}
		return nil
	})
}

const insertSystemSQL = `INSERT INTO messages (workspace_id, type, text) VALUES ($1, 'system', $2)`

// InsertSystemMessage пишет служебную строку в чат пары (например, «Встреча назначена»).
// Вызывается другими модулями внутри их транзакций.
func InsertSystemMessage(ctx context.Context, db postgres.DBTX, workspaceID uuid.UUID, text string) error {
	if _, err := db.Exec(ctx, insertSystemSQL, workspaceID, text); err != nil {
		return fmt.Errorf("insert system message: %w", err)
	}
	return nil
}

func scanMessage(row pgx.Row) (Message, error) {
	var (
		m       Message
		msgType string
		fileID  *uuid.UUID
		fType   *string
		fName   *string
		fURL    *string
		fSize   *int
		fBy     *uuid.UUID
		fAt     *time.Time
	)
	err := row.Scan(&m.ID, &m.WorkspaceID, &m.SenderID, &msgType, &m.Text, &m.IsRead, &m.CreatedAt,
		&fileID, &fType, &fName, &fURL, &fSize, &fBy, &fAt)
	if err != nil {
		return Message{}, err
	}
	m.Type = MessageType(msgType)
	if fileID != nil {
		m.File = &File{ID: *fileID, Type: *fType, FileName: *fName, FileURL: *fURL, FileSizeKB: fSize, UploadedBy: *fBy, CreatedAt: *fAt}
	}
	return m, nil
}
