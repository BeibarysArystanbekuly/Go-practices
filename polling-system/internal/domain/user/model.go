package user

import "time"

type User struct {
    ID           int64     `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    Role         string    `json:"role"`
    CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
    Create(u *User) error
    GetByEmail(email string) (*User, error)
    GetByID(id int64) (*User, error)
    List() ([]User, error)
    UpdateRole(id int64, role string) error
}
