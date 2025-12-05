package postgres

import (
	"context"
	"database/sql"

	"polling-system/internal/domain/user"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	query := `
        INSERT INTO users (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
	return r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
        SELECT id, email, password_hash, role, created_at
        FROM users WHERE email = $1
    `
	u := &user.User{}
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
        SELECT id, email, password_hash, role, created_at
        FROM users WHERE id = $1
    `
	u := &user.User{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, email, password_hash, role, created_at
        FROM users ORDER BY id
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersList []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		usersList = append(usersList, u)
	}
	return usersList, nil
}

func (r *UserRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, id)
	return err
}
