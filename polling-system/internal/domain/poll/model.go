package poll

import "time"

type Poll struct {
    ID          int64      `json:"id"`
    Title       string     `json:"title"`
    Description *string    `json:"description,omitempty"`
    Status      string     `json:"status"`
    StartsAt    *time.Time `json:"starts_at,omitempty"`
    EndsAt      *time.Time `json:"ends_at,omitempty"`
    CreatorID   int64      `json:"creator_id"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type Option struct {
    ID        int64     `json:"id"`
    PollID    int64     `json:"poll_id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
    Create(p *Poll, options []Option) (int64, error)
    GetByID(id int64) (*Poll, []Option, error)
    List(status *string) ([]Poll, error)
    UpdateStatus(id int64, status string) error
}
