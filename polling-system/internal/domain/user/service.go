package user

import (
    "errors"

    "golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrEmailTaken         = errors.New("email already taken")
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Register(email, password string) (*User, error) {
    if email == "" || password == "" {
        return nil, errors.New("email and password required")
    }

    if _, err := s.repo.GetByEmail(email); err == nil {
        return nil, ErrEmailTaken
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    u := &User{
        Email:        email,
        PasswordHash: string(hash),
        Role:         "user",
    }

    if err := s.repo.Create(u); err != nil {
        return nil, err
    }

    return u, nil
}

func (s *Service) Login(email, password string) (*User, error) {
    u, err := s.repo.GetByEmail(email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
        return nil, ErrInvalidCredentials
    }

    return u, nil
}

func (s *Service) List() ([]User, error) {
    return s.repo.List()
}

func (s *Service) UpdateRole(id int64, role string) error {
    if role != "admin" && role != "user" {
        return errors.New("invalid role")
    }
    return s.repo.UpdateRole(id, role)
}

func (s *Service) GetByID(id int64) (*User, error) {
    return s.repo.GetByID(id)
}
