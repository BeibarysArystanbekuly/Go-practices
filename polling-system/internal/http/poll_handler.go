package api

import (
    "encoding/json"
    "net/http"

    "polling-system/internal/domain/poll"
)

type createPollRequest struct {
    Title       string   `json:"title"`
    Description *string  `json:"description"`
    StartsAt    *string  `json:"starts_at"`
    EndsAt      *string  `json:"ends_at"`
    Options     []string `json:"options"`
}

type updateStatusRequest struct {
    Status string `json:"status"`
}

func (h *Handler) handleCreatePoll(w http.ResponseWriter, r *http.Request) {
    var req createPollRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    userID := userIDFromCtx(r)

    p := &poll.Poll{
        Title:       req.Title,
        Description: req.Description,
        StartsAt:    parseTimePtr(req.StartsAt),
        EndsAt:      parseTimePtr(req.EndsAt),
        CreatorID:   userID,
    }

    opts := make([]poll.Option, 0, len(req.Options))
    for _, text := range req.Options {
        opts = append(opts, poll.Option{Text: text})
    }

    id, err := h.pollSvc.Create(p, opts)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) handleListPolls(w http.ResponseWriter, r *http.Request) {
    statusParam := r.URL.Query().Get("status")
    var status *string
    if statusParam != "" {
        status = &statusParam
    }
    polls, err := h.pollSvc.List(status)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, polls)
}

func (h *Handler) handleGetPoll(w http.ResponseWriter, r *http.Request) {
    id, err := parseIDParam(r, "id")
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    p, opts, err := h.pollSvc.Get(id)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "poll":    p,
        "options": opts,
    })
}

func (h *Handler) handleUpdatePollStatus(w http.ResponseWriter, r *http.Request) {
    id, err := parseIDParam(r, "id")
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    var req updateStatusRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    if err := h.pollSvc.UpdateStatus(id, req.Status); err != nil {
        http.Error(w, "bad status", http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
