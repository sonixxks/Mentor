package requests

import (
	"context"
	"errors"
	"testing"
)

const (
	studentID = "student-1"
	mentorID  = "mentor-1"
	requestID = "request-1"
)

var (
	student      = Actor{UserID: studentID, Role: RoleStudent}
	otherStudent = Actor{UserID: "student-2", Role: RoleStudent}
	mentor       = Actor{UserID: mentorID, Role: RoleMentor}
	otherMentor  = Actor{UserID: "mentor-2", Role: RoleMentor}
	curator      = Actor{UserID: "curator-1", Role: RoleCurator}
)

// fakeRepo хранит одну заявку и запоминает, какие методы вызывались.
type fakeRepo struct {
	req            Request
	acceptCalled   bool
	updatedTo      Status
	curatorComment *string
}

func newFakeRepo(status Status) *fakeRepo {
	return &fakeRepo{req: Request{ID: requestID, StudentID: studentID, MentorID: mentorID, Status: status}}
}

func (f *fakeRepo) Create(_ context.Context, studentID, mentorID string, packageID *string) (Request, error) {
	return Request{StudentID: studentID, MentorID: mentorID, ServicePackageID: packageID, Status: StatusPending}, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id string) (Request, error) {
	if id != f.req.ID {
		return Request{}, ErrNotFound
	}
	return f.req, nil
}

func (f *fakeRepo) ListByStudent(context.Context, string) ([]Request, error) { return nil, nil }
func (f *fakeRepo) ListByMentor(context.Context, string) ([]Request, error)  { return nil, nil }

func (f *fakeRepo) UpdateStatus(_ context.Context, _ string, to Status, curatorComment *string) (Request, error) {
	f.updatedTo = to
	f.curatorComment = curatorComment
	f.req.Status = to
	f.req.CuratorComment = curatorComment
	return f.req, nil
}

func (f *fakeRepo) Accept(context.Context, string) (Request, Workspace, error) {
	f.acceptCalled = true
	f.req.Status = StatusAccepted
	return f.req, Workspace{ID: "ws-1", RequestID: f.req.ID}, nil
}

func TestChangeStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		actor   Actor
		to      Status
		comment string
		want    error
	}{
		{"student cancels own pending", StatusPending, student, StatusCancelled, "", nil},
		{"mentor rejects own pending", StatusPending, mentor, StatusRejectedByMentor, "", nil},
		{"curator rejects with comment", StatusPending, curator, StatusRejectedByCurator, "spam", nil},

		{"other student cannot cancel", StatusPending, otherStudent, StatusCancelled, "", ErrNotFound},
		{"other mentor cannot reject", StatusPending, otherMentor, StatusRejectedByMentor, "", ErrNotFound},
		{"curator without comment", StatusPending, curator, StatusRejectedByCurator, "   ", ErrCuratorCommentRequired},
		{"student cannot cancel accepted", StatusAccepted, student, StatusCancelled, "", ErrInvalidTransition},
		{"student cannot accept", StatusPending, student, StatusAccepted, "", ErrForbidden},
		{"unknown status", StatusPending, mentor, Status("approved"), "", ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo(tt.status)
			svc := NewService(repo)

			_, err := svc.ChangeStatus(context.Background(), tt.actor, ChangeStatusInput{
				RequestID:      requestID,
				To:             tt.to,
				CuratorComment: tt.comment,
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

func TestChangeStatusAcceptCreatesWorkspace(t *testing.T) {
	repo := newFakeRepo(StatusPending)
	svc := NewService(repo)

	res, err := svc.ChangeStatus(context.Background(), mentor, ChangeStatusInput{RequestID: requestID, To: StatusAccepted})
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if !repo.acceptCalled {
		t.Fatal("accept must go through repo.Accept (transaction with workspace)")
	}
	if res.Workspace == nil || res.Workspace.RequestID != requestID {
		t.Fatalf("workspace = %+v, want workspace for %s", res.Workspace, requestID)
	}
}

func TestChangeStatusCuratorCommentTrimmed(t *testing.T) {
	repo := newFakeRepo(StatusPending)
	svc := NewService(repo)

	_, err := svc.ChangeStatus(context.Background(), curator, ChangeStatusInput{
		RequestID:      requestID,
		To:             StatusRejectedByCurator,
		CuratorComment: "  жалоба на спам  ",
	})
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if repo.curatorComment == nil || *repo.curatorComment != "жалоба на спам" {
		t.Fatalf("curator comment = %v, want trimmed text", repo.curatorComment)
	}
}

func TestCreate(t *testing.T) {
	svc := NewService(newFakeRepo(StatusPending))

	if _, err := svc.Create(context.Background(), mentor, "mentor-2", nil); !errors.Is(err, ErrForbidden) {
		t.Fatalf("mentor creates request: error = %v, want %v", err, ErrForbidden)
	}
	if _, err := svc.Create(context.Background(), student, studentID, nil); !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("request to self: error = %v, want %v", err, ErrInvalidTarget)
	}

	req, err := svc.Create(context.Background(), student, mentorID, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if req.StudentID != studentID || req.MentorID != mentorID {
		t.Fatalf("request = %+v, want student %s -> mentor %s", req, studentID, mentorID)
	}
}
