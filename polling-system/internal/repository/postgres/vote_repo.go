package postgres

import (
    "database/sql"

    "polling-system/internal/domain/vote"
)

type VoteRepo struct {
    db *sql.DB
}

func NewVoteRepo(db *sql.DB) *VoteRepo {
    return &VoteRepo{db: db}
}

func (r *VoteRepo) Create(v *vote.Vote) error {
    query := `
        INSERT INTO votes (poll_id, option_id, user_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
    return r.db.QueryRow(query, v.PollID, v.OptionID, v.UserID).
        Scan(&v.ID, &v.CreatedAt)
}

func (r *VoteRepo) HasUserVoted(pollID, userID int64) (bool, error) {
    query := `
        SELECT 1 FROM votes
        WHERE poll_id = $1 AND user_id = $2
        LIMIT 1
    `
    var dummy int
    err := r.db.QueryRow(query, pollID, userID).Scan(&dummy)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *VoteRepo) CountByPoll(pollID int64) (map[int64]int64, int64, error) {
    rows, err := r.db.Query(`
        SELECT option_id, COUNT(*)
        FROM votes
        WHERE poll_id = $1
        GROUP BY option_id
    `, pollID)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    res := make(map[int64]int64)
    var total int64
    for rows.Next() {
        var optID int64
        var c int64
        if err := rows.Scan(&optID, &c); err != nil {
            return nil, 0, err
        }
        res[optID] = c
        total += c
    }

    return res, total, nil
}
