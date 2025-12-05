package postgres

import (
	"context"
	"database/sql"

	"polling-system/internal/domain/poll"
)

type PollRepo struct {
	db *sql.DB
}

func NewPollRepo(db *sql.DB) *PollRepo {
	return &PollRepo{db: db}
}

func (r *PollRepo) Create(ctx context.Context, p *poll.Poll, options []poll.Option) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queryPoll := `
        INSERT INTO polls (title, description, status, starts_at, ends_at, creator_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRowContext(ctx, queryPoll,
		p.Title,
		p.Description,
		p.Status,
		p.StartsAt,
		p.EndsAt,
		p.CreatorID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return 0, err
	}

	queryOpt := `
        INSERT INTO options (poll_id, text)
        VALUES ($1, $2)
        RETURNING id, created_at
    `

	for i := range options {
		options[i].PollID = p.ID
		if err := tx.QueryRowContext(ctx, queryOpt, options[i].PollID, options[i].Text).
			Scan(&options[i].ID, &options[i].CreatedAt); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return p.ID, nil
}

func (r *PollRepo) GetByID(ctx context.Context, id int64) (*poll.Poll, []poll.Option, error) {
	p := &poll.Poll{}
	err := r.db.QueryRowContext(ctx, `
        SELECT id, title, description, status, starts_at, ends_at, creator_id, created_at, updated_at
        FROM polls WHERE id = $1
    `, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.Status,
		&p.StartsAt, &p.EndsAt, &p.CreatorID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, poll_id, text, created_at
        FROM options WHERE poll_id = $1
    `, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var opts []poll.Option
	for rows.Next() {
		var o poll.Option
		if err := rows.Scan(&o.ID, &o.PollID, &o.Text, &o.CreatedAt); err != nil {
			return nil, nil, err
		}
		opts = append(opts, o)
	}

	return p, opts, nil
}

func (r *PollRepo) List(ctx context.Context, status *string) ([]poll.Poll, error) {
	query := `
        SELECT id, title, description, status, starts_at, ends_at, creator_id, created_at, updated_at
        FROM polls
    `
	var rows *sql.Rows
	var err error

	if status != nil {
		query += " WHERE status = $1 ORDER BY created_at DESC"
		rows, err = r.db.QueryContext(ctx, query, *status)
	} else {
		query += " ORDER BY created_at DESC"
		rows, err = r.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []poll.Poll
	for rows.Next() {
		var p poll.Poll
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Status,
			&p.StartsAt, &p.EndsAt, &p.CreatorID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *PollRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE polls SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}
