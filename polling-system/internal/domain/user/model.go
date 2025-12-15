package user

import (
	"context"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	IsActive     bool      `json:"is_active"`
}

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	UpdateRole(ctx context.Context, id int64, role string) error
	Deactivate(ctx context.Context, id int64) error
}
