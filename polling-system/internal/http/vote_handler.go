package api

import (
    "encoding/json"
    "net/http"

    "polling-system/internal/domain/vote"
    "polling-system/internal/worker"
)

type voteRequest struct {
    OptionID int64 `json:"option_id"`
}

func (h *Handler) handleVote(w http.ResponseWriter, r *http.Request) {
    pollID, err := parseIDParam(r, "id")
    if err != nil {
        http.Error(w, "invalid poll id", http.StatusBadRequest)
        return
    }

    var req voteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    userID := userIDFromCtx(r)

    if err := h.voteSvc.Vote(pollID, req.OptionID, userID); err != nil {
        if err == vote.ErrAlreadyVoted {
            http.Error(w, "already voted", http.StatusBadRequest)
            return
        }
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    select {
    case h.voteCh <- worker.VoteEvent{PollID: pollID, OptionID: req.OptionID}:
    default:
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handlePollResults(w http.ResponseWriter, r *http.Request) {
    pollID, err := parseIDParam(r, "id")
    if err != nil {
        http.Error(w, "invalid poll id", http.StatusBadRequest)
        return
    }

    res, total, err := h.voteSvc.Results(pollID)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "poll_id":     pollID,
        "total_votes": total,
        "options":     res,
    })
}
