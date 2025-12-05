package poll

import (
	"context"
	"errors"
)

var (
	ErrInvalidStatus = errors.New("invalid poll status")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, p *Poll, options []Option) (int64, error) {
	if p.Title == "" {
		return 0, errors.New("title required")
	}
	if len(options) < 2 {
		return 0, errors.New("poll must have at least 2 options")
	}
	p.Status = "draft"
	return s.repo.Create(ctx, p, options)
}

func (s *Service) Get(ctx context.Context, id int64) (*Poll, []Option, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, status *string) ([]Poll, error) {
	return s.repo.List(ctx, status)
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, status string) error {
	if status != "draft" && status != "active" && status != "closed" {
		return ErrInvalidStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}
