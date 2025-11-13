package poll

import "errors"

var (
    ErrInvalidStatus = errors.New("invalid poll status")
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Create(p *Poll, options []Option) (int64, error) {
    if p.Title == "" {
        return 0, errors.New("title required")
    }
    if len(options) < 2 {
        return 0, errors.New("poll must have at least 2 options")
    }
    p.Status = "draft"
    return s.repo.Create(p, options)
}

func (s *Service) Get(id int64) (*Poll, []Option, error) {
    return s.repo.GetByID(id)
}

func (s *Service) List(status *string) ([]Poll, error) {
    return s.repo.List(status)
}

func (s *Service) UpdateStatus(id int64, status string) error {
    if status != "draft" && status != "active" && status != "closed" {
        return ErrInvalidStatus
    }
    return s.repo.UpdateStatus(id, status)
}
