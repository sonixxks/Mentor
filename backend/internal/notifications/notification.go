// Package notifications хранит уведомления пользователей.
// Пока только запись: другие модули создают уведомления в своих транзакциях.
// Выдача и отметки о прочтении — отдельная задача.
package notifications

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
)

type Type string

const (
	TypeRequestUpdate      Type = "request_update"
	TypeTaskReview         Type = "task_review"
	TypeNewMessage         Type = "new_message"
	TypeNewRecommendations Type = "new_recommendations"
	TypeMeetingUpdate      Type = "meeting_update"
	TypeSystemAlert        Type = "system_alert"
)

type Notification struct {
	UserID   uuid.UUID
	Type     Type
	Title    string
	Message  *string
	EntityID *uuid.UUID
}

const insertSQL = `
	INSERT INTO notifications (user_id, type, title, message, entity_id)
	VALUES ($1, $2::notification_type, $3, $4, $5)`

// Insert записывает уведомления через переданное соединение или транзакцию.
func Insert(ctx context.Context, db postgres.DBTX, list ...Notification) error {
	for _, n := range list {
		if _, err := db.Exec(ctx, insertSQL, n.UserID, string(n.Type), n.Title, n.Message, n.EntityID); err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	return nil
}

func Ptr[T any](v T) *T {
	return &v
}
