package requests

import (
	"context"
	"strings"
)

type Repository interface {
	Create(ctx context.Context, studentID, mentorID string, packageID *string) (Request, error)
	GetByID(ctx context.Context, id string) (Request, error)
	ListByStudent(ctx context.Context, studentID string) ([]Request, error)
	ListByMentor(ctx context.Context, mentorID string) ([]Request, error)
	UpdateStatus(ctx context.Context, id string, to Status, curatorComment *string) (Request, error)
	Accept(ctx context.Context, id string) (Request, Workspace, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, actor Actor, mentorID string, packageID *string) (Request, error) {
	if actor.Role != RoleStudent {
		return Request{}, ErrForbidden
	}
	if mentorID == actor.UserID {
		return Request{}, ErrInvalidTarget
	}
	return s.repo.Create(ctx, actor.UserID, mentorID, packageID)
}

// Get возвращает ErrNotFound и для чужих заявок, чтобы не раскрывать их существование.
func (s *Service) Get(ctx context.Context, actor Actor, id string) (Request, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Request{}, err
	}
	if !canAccess(actor, req) {
		return Request{}, ErrNotFound
	}
	return req, nil
}

// ListMine — входящие заявки для ментора, исходящие для студента.
func (s *Service) ListMine(ctx context.Context, actor Actor) ([]Request, error) {
	switch actor.Role {
	case RoleStudent:
		return s.repo.ListByStudent(ctx, actor.UserID)
	case RoleMentor:
		return s.repo.ListByMentor(ctx, actor.UserID)
	default:
		return nil, ErrForbidden
	}
}

type ChangeStatusInput struct {
	RequestID      string
	To             Status
	CuratorComment string
}

type ChangeStatusResult struct {
	Request Request
	// Workspace заполнен только при переходе в accepted.
	Workspace *Workspace
}

func (s *Service) ChangeStatus(ctx context.Context, actor Actor, in ChangeStatusInput) (ChangeStatusResult, error) {
	if !in.To.Valid() {
		return ChangeStatusResult{}, ErrInvalidStatus
	}

	req, err := s.Get(ctx, actor, in.RequestID)
	if err != nil {
		return ChangeStatusResult{}, err
	}

	if err := CheckTransition(req.Status, in.To, actor.Role); err != nil {
		return ChangeStatusResult{}, err
	}

	switch in.To {
	case StatusAccepted:
		updated, ws, err := s.repo.Accept(ctx, req.ID)
		if err != nil {
			return ChangeStatusResult{}, err
		}
		return ChangeStatusResult{Request: updated, Workspace: &ws}, nil

	case StatusRejectedByCurator:
		comment := strings.TrimSpace(in.CuratorComment)
		if comment == "" {
			return ChangeStatusResult{}, ErrCuratorCommentRequired
		}
		updated, err := s.repo.UpdateStatus(ctx, req.ID, in.To, &comment)
		if err != nil {
			return ChangeStatusResult{}, err
		}
		return ChangeStatusResult{Request: updated}, nil

	default:
		updated, err := s.repo.UpdateStatus(ctx, req.ID, in.To, nil)
		if err != nil {
			return ChangeStatusResult{}, err
		}
		return ChangeStatusResult{Request: updated}, nil
	}
}

func canAccess(actor Actor, req Request) bool {
	switch actor.Role {
	case RoleCurator:
		return true
	case RoleStudent:
		return req.StudentID == actor.UserID
	case RoleMentor:
		return req.MentorID == actor.UserID
	default:
		return false
	}
}
