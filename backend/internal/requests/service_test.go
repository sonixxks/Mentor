package requests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/notifications"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

var (
	studentID = uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	mentorID  = uuid.MustParse("00000000-0000-0000-0000-00000000000b")
	requestID = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

	student      = Actor{UserID: studentID, Role: user.Student}
	otherStudent = Actor{UserID: uuid.New(), Role: user.Student}
	mentor       = Actor{UserID: mentorID, Role: user.Mentor}
	otherMentor  = Actor{UserID: uuid.New(), Role: user.Mentor}
	curator      = Actor{UserID: uuid.New(), Role: user.Curator}
)

// fakeRepo хранит одну заявку и запоминает, что с ней делали.
type fakeRepo struct {
	req          Request
	created      *CreateParams
	createNotify []notifications.Notification
	acceptCalled bool
	updatedTo    Status
	reason       *string
	notified     []notifications.Notification
	listFilter   ListFilter
}

func newFakeRepo(status Status) *fakeRepo {
	return &fakeRepo{req: Request{ID: requestID, StudentID: studentID, MentorID: mentorID, Status: status}}
}

func (f *fakeRepo) Create(_ context.Context, p CreateParams, notify func(Request) []notifications.Notification) (Request, error) {
	f.created = &p
	f.req = Request{ID: requestID, StudentID: p.StudentID, MentorID: p.MentorID, Status: StatusPending}
	f.createNotify = notify(f.req)
	return f.req, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (Request, error) {
	if id != f.req.ID {
		return Request{}, ErrNotFound
	}
	return f.req, nil
}

func (f *fakeRepo) GetView(_ context.Context, id uuid.UUID) (View, error) {
	if id != f.req.ID {
		return View{}, ErrNotFound
	}
	return View{Request: f.req}, nil
}

func (f *fakeRepo) List(_ context.Context, filter ListFilter) ([]View, int, error) {
	f.listFilter = filter
	return nil, 0, nil
}

func (f *fakeRepo) UpdateStatus(_ context.Context, _ uuid.UUID, from, to Status, reason *string, notify []notifications.Notification) (Request, error) {
	if from != f.req.Status {
		return Request{}, ErrStatusConflict
	}
	f.updatedTo, f.reason, f.notified = to, reason, notify
	f.req.Status, f.req.RejectionReason = to, reason
	return f.req, nil
}

func (f *fakeRepo) Accept(_ context.Context, _ uuid.UUID, notify []notifications.Notification) (Request, uuid.UUID, error) {
	f.acceptCalled, f.notified = true, notify
	f.req.Status = StatusAccepted
	return f.req, uuid.New(), nil
}

func TestChangeStatus(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		actor  Actor
		to     Status
		reason string
		want   error
	}{
		{"student cancels own pending", StatusPending, student, StatusCancelled, "", nil},
		{"mentor approves own pending", StatusPending, mentor, StatusApprovedByMentor, "", nil},
		{"mentor rejects without reason", StatusPending, mentor, StatusRejectedByMentor, "", nil},
		{"curator rejects with reason", StatusApprovedByMentor, curator, StatusRejectedByCurator, "spam", nil},
		{"student accepts approved", StatusApprovedByMentor, student, StatusAccepted, "", nil},

		{"other student cannot cancel", StatusPending, otherStudent, StatusCancelled, "", ErrNotFound},
		{"other mentor cannot approve", StatusPending, otherMentor, StatusApprovedByMentor, "", ErrNotFound},
		{"curator without reason", StatusPending, curator, StatusRejectedByCurator, "   ", ErrReasonRequired},
		{"student cannot cancel accepted", StatusAccepted, student, StatusCancelled, "", ErrInvalidTransition},
		{"mentor cannot accept for student", StatusApprovedByMentor, mentor, StatusAccepted, "", ErrForbidden},
		{"unknown status", StatusPending, mentor, Status("approved"), "", ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo(tt.status)
			svc := NewService(repo, nil)

			_, err := svc.ChangeStatus(context.Background(), tt.actor, ChangeStatusInput{
				RequestID:       requestID,
				To:              tt.to,
				RejectionReason: tt.reason,
			})
			if !errors.Is(err, tt.want) {
				t.Fatalf("ChangeStatus() error = %v, want %v", err, tt.want)
			}
			if tt.want != nil && (repo.updatedTo != "" || repo.acceptCalled) {
				t.Fatal("repository must not be changed on error")
			}
		})
	}
}

func TestChangeStatusAcceptGoesThroughTransaction(t *testing.T) {
	repo := newFakeRepo(StatusApprovedByMentor)

	view, err := NewService(repo, nil).ChangeStatus(context.Background(), student, ChangeStatusInput{RequestID: requestID, To: StatusAccepted})
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if !repo.acceptCalled || repo.updatedTo != "" {
		t.Fatal("accept must go through repo.Accept (transaction with workspace)")
	}
	if view.Status != StatusAccepted {
		t.Fatalf("status = %s, want accepted", view.Status)
	}
	if len(repo.notified) != 1 || repo.notified[0].UserID != mentorID {
		t.Fatalf("notifications = %+v, want one for mentor", repo.notified)
	}
}

func TestChangeStatusReason(t *testing.T) {
	repo := newFakeRepo(StatusPending)
	_, err := NewService(repo, nil).ChangeStatus(context.Background(), curator, ChangeStatusInput{
		RequestID:       requestID,
		To:              StatusRejectedByCurator,
		RejectionReason: "  жалоба на спам  ",
	})
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if repo.reason == nil || *repo.reason != "жалоба на спам" {
		t.Fatalf("reason = %v, want trimmed text", repo.reason)
	}
	if len(repo.notified) != 2 {
		t.Fatalf("curator decision must notify both sides, got %d", len(repo.notified))
	}

	// Причина при одобрении не сохраняется.
	repo = newFakeRepo(StatusPending)
	_, err = NewService(repo, nil).ChangeStatus(context.Background(), mentor, ChangeStatusInput{
		RequestID: requestID, To: StatusApprovedByMentor, RejectionReason: "лишнее",
	})
	if err != nil || repo.reason != nil {
		t.Fatalf("approve: err = %v, reason = %v, want nil reason", err, repo.reason)
	}
	if len(repo.notified) != 1 || repo.notified[0].UserID != studentID {
		t.Fatalf("approve must notify the student, got %+v", repo.notified)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	if _, err := NewService(newFakeRepo(StatusPending), nil).Create(ctx, mentor, CreateInput{MentorID: uuid.New()}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("mentor creates request: error = %v, want %v", err, ErrForbidden)
	}

	long := strings.Repeat("я", MaxCoverLetterLen+1)
	if _, err := NewService(newFakeRepo(StatusPending), nil).Create(ctx, student, CreateInput{MentorID: mentorID, CoverLetter: long}); !errors.Is(err, ErrCoverLetterLength) {
		t.Fatalf("long cover letter: error = %v, want %v", err, ErrCoverLetterLength)
	}

	repo := newFakeRepo(StatusPending)
	view, err := NewService(repo, nil).Create(ctx, student, CreateInput{MentorID: mentorID, CoverLetter: "  Хочу в бэкенд  "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if view.StudentID != studentID || view.MentorID != mentorID {
		t.Fatalf("request = %+v, want student -> mentor", view.Request)
	}
	if repo.created.CoverLetter == nil || *repo.created.CoverLetter != "Хочу в бэкенд" || repo.created.SuggestedBy != nil {
		t.Fatalf("create params = %+v", repo.created)
	}
	if len(repo.createNotify) != 1 || repo.createNotify[0].UserID != mentorID {
		t.Fatalf("new request must notify the mentor, got %+v", repo.createNotify)
	}
}

func TestCreateByCurator(t *testing.T) {
	ctx := context.Background()

	if _, err := NewService(newFakeRepo(StatusPending), nil).CreateByCurator(ctx, student, CuratorCreateInput{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("student uses curator route: error = %v, want %v", err, ErrForbidden)
	}

	repo := newFakeRepo(StatusPending)
	if _, err := NewService(repo, nil).CreateByCurator(ctx, curator, CuratorCreateInput{StudentID: studentID, MentorID: mentorID}); err != nil {
		t.Fatalf("CreateByCurator() error = %v", err)
	}
	if repo.created.SuggestedBy == nil || *repo.created.SuggestedBy != curator.UserID {
		t.Fatalf("suggested_by = %v, want curator", repo.created.SuggestedBy)
	}
	if len(repo.createNotify) != 2 {
		t.Fatalf("both sides must be notified, got %d", len(repo.createNotify))
	}
}

func TestListMine(t *testing.T) {
	ctx := context.Background()

	repo := newFakeRepo(StatusPending)
	if _, _, err := NewService(repo, nil).ListMine(ctx, student, GroupActive, 10, 0); err != nil {
		t.Fatal(err)
	}
	if repo.listFilter.StudentID == nil || *repo.listFilter.StudentID != studentID || repo.listFilter.MentorID != nil {
		t.Fatalf("student filter = %+v", repo.listFilter)
	}
	if len(repo.listFilter.Statuses) != len(ActiveStatuses) {
		t.Fatalf("active group statuses = %v", repo.listFilter.Statuses)
	}

	if _, _, err := NewService(repo, nil).ListMine(ctx, mentor, GroupArchive, 10, 0); err != nil {
		t.Fatal(err)
	}
	if repo.listFilter.MentorID == nil || *repo.listFilter.MentorID != mentorID || len(repo.listFilter.Statuses) != len(ArchiveStatuses) {
		t.Fatalf("mentor filter = %+v", repo.listFilter)
	}

	if _, _, err := NewService(repo, nil).ListMine(ctx, curator, GroupAll, 10, 0); !errors.Is(err, ErrForbidden) {
		t.Fatalf("curator ListMine: error = %v, want %v", err, ErrForbidden)
	}

	bad := Status("nope")
	if _, _, err := NewService(repo, nil).ListAll(ctx, curator, &bad, 10, 0); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("ListAll bad status: error = %v", err)
	}
}
