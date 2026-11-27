package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type fakeRepo struct {
	members  Members
	text     *string
	notice   notifications.Notification
	gotLimit int
}

func (f *fakeRepo) Members(context.Context, uuid.UUID) (Members, error) { return f.members, nil }

func (f *fakeRepo) List(_ context.Context, _ uuid.UUID, _ *time.Time, limit int) ([]Message, error) {
	f.gotLimit = limit
	return nil, nil
}

func (f *fakeRepo) Create(_ context.Context, ws, sender uuid.UUID, text *string, _ *uuid.UUID,
	notify func(Message) notifications.Notification,
) (Message, error) {
	f.text = text
	m := Message{ID: uuid.New(), WorkspaceID: ws, SenderID: &sender, Text: text}
	f.notice = notify(m)
	return m, nil
}

func (f *fakeRepo) MarkRead(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func TestSend(t *testing.T) {
	studentID, mentorID := uuid.New(), uuid.New()
	repo := &fakeRepo{members: Members{StudentID: studentID, MentorID: mentorID}}
	svc := NewService(repo)
	ctx := context.Background()
	ws := uuid.New()

	student := Actor{UserID: studentID, Role: user.Student}
	if _, err := svc.Send(ctx, student, ws, SendInput{Text: "   "}); !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("empty: error = %v", err)
	}
	if _, err := svc.Send(ctx, student, ws, SendInput{Text: strings.Repeat("я", MaxTextLen+1)}); !errors.Is(err, ErrTextTooLong) {
		t.Fatalf("too long: error = %v", err)
	}
	if _, err := svc.Send(ctx, Actor{UserID: uuid.New(), Role: user.Mentor}, ws, SendInput{Text: "hi"}); !errors.Is(err, ErrNotMember) {
		t.Fatalf("stranger: error = %v", err)
	}
	if _, err := svc.Send(ctx, Actor{UserID: uuid.New(), Role: user.Curator}, ws, SendInput{Text: "hi"}); !errors.Is(err, ErrNotMember) {
		t.Fatalf("curator: error = %v", err)
	}

	long := strings.Repeat("а", 150)
	if _, err := svc.Send(ctx, student, ws, SendInput{Text: "  " + long + "  "}); err != nil {
		t.Fatal(err)
	}
	if repo.text == nil || *repo.text != long {
		t.Fatal("text must be trimmed")
	}
	if repo.notice.UserID != mentorID || repo.notice.Type != notifications.TypeNewMessage {
		t.Fatalf("notice = %+v, want new_message for mentor", repo.notice)
	}
	if n := len([]rune(*repo.notice.Message)); n != previewLen+1 {
		t.Fatalf("preview length = %d", n)
	}

	if _, err := svc.History(ctx, Actor{UserID: mentorID, Role: user.Mentor}, ws, nil, 0); err != nil || repo.gotLimit != DefaultPageLimit {
		t.Fatalf("history: err = %v, limit = %d", err, repo.gotLimit)
	}
}
