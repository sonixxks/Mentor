// Package mailer — отправка писем.
package mailer

import (
	"context"
	"log/slog"

	"github.com/sonixxks/Mentor/backend/internal/auth"
)

var (
	_ auth.Mailer              = (*LogMailer)(nil)
	_ auth.PasswordResetMailer = (*LogMailer)(nil)
)

// LogMailer вместо отправки письма пишет код в лог.
// ТОЛЬКО для разработки: код подтверждения в логах — это уязвимость.
// В app.go подключать только при APP_ENV=development, позже заменить на SMTP (Mailpit / реальный сервер).
type LogMailer struct {
	log *slog.Logger
}

func NewLogMailer(log *slog.Logger) *LogMailer {
	return &LogMailer{log: log}
}

func (m *LogMailer) SendVerificationCode(ctx context.Context, email, code string) error {
	m.log.InfoContext(ctx, "DEV MAILER: verification code",
		"email", email,
		"code", code,
	)
	return nil
}

func (m *LogMailer) SendPasswordResetCode(ctx context.Context, email, code string) error {
	m.log.InfoContext(ctx, "DEV MAILER: password reset code",
		"email", email,
		"code", code,
	)
	return nil
}
