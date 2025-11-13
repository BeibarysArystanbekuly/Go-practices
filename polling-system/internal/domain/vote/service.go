package vote

import "errors"

var (
    ErrAlreadyVoted = errors.New("user already voted in this poll")
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Vote(pollID, optionID, userID int64) error {
    ok, err := s.repo.HasUserVoted(pollID, userID)
    if err != nil {
        return err
    }
    if ok {
        return ErrAlreadyVoted
    }

    v := &Vote{
        PollID:   pollID,
        OptionID: optionID,
        UserID:   userID,
    }

    return s.repo.Create(v)
}

type Result struct {
    OptionID   int64   `json:"option_id"`
    Votes      int64   `json:"votes"`
    Percentage float64 `json:"percentage"`
}

func (s *Service) Results(pollID int64) ([]Result, int64, error) {
    counts, total, err := s.repo.CountByPoll(pollID)
    if err != nil {
        return nil, 0, err
    }

    results := make([]Result, 0, len(counts))
    for optionID, c := range counts {
        var p float64
        if total > 0 {
            p = float64(c) * 100.0 / float64(total)
        }
        results = append(results, Result{
            OptionID:   optionID,
            Votes:      c,
            Percentage: p,
        })
    }

    return results, total, nil
}
