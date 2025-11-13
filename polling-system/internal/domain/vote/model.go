package vote

import "time"

type Vote struct {
    ID        int64     `json:"id"`
    PollID    int64     `json:"poll_id"`
    OptionID  int64     `json:"option_id"`
    UserID    int64     `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
    Create(v *Vote) error
    HasUserVoted(pollID, userID int64) (bool, error)
    CountByPoll(pollID int64) (map[int64]int64, int64, error)
}
