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

type pollDetailsResponse struct {
	Poll    *poll.Poll    `json:"poll"`
	Options []poll.Option `json:"options"`
}

// @Summary     Create poll
// @Description Admin only
// @Tags        polls
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       request  body      createPollRequest  true  "Poll payload"
// @Success     201      {object}  map[string]int64
// @Failure     400      {string}  string  "invalid body"
// @Failure     401      {string}  string  "unauthorized"
// @Failure     403      {string}  string  "forbidden"
// @Failure     500      {string}  string  "server error"
// @Router      /api/v1/polls [post]
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

	id, err := h.pollSvc.Create(r.Context(), p, opts)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// @Summary     List polls
// @Tags        polls
// @Security    BearerAuth
// @Produce     json
// @Param       status  query     string  false  "Filter by status"  Enums(draft,active,closed)
// @Success     200     {array}   poll.Poll
// @Failure     401     {string}  string  "unauthorized"
// @Failure     500     {string}  string  "server error"
// @Router      /api/v1/polls [get]
func (h *Handler) handleListPolls(w http.ResponseWriter, r *http.Request) {
	statusParam := r.URL.Query().Get("status")
	var status *string
	if statusParam != "" {
		status = &statusParam
	}
	polls, err := h.pollSvc.List(r.Context(), status)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, polls)
}

// @Summary     Get poll with options
// @Tags        polls
// @Security    BearerAuth
// @Produce     json
// @Param       id   path     int64  true  "Poll ID"
// @Success     200  {object} pollDetailsResponse
// @Failure     400  {string}  string  "invalid id"
// @Failure     401  {string}  string  "unauthorized"
// @Failure     404  {string}  string  "not found"
// @Failure     500  {string}  string  "server error"
// @Router      /api/v1/polls/{id} [get]
func (h *Handler) handleGetPoll(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, opts, err := h.pollSvc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"poll":    p,
		"options": opts,
	})
}

// @Summary     Update poll status
// @Description Admin only
// @Tags        polls
// @Security    BearerAuth
// @Accept      json
// @Param       id       path      int64                true  "Poll ID"
// @Param       request  body      updateStatusRequest  true  "New status"
// @Success     204
// @Failure     400      {string}  string  "invalid id or status"
// @Failure     401      {string}  string  "unauthorized"
// @Failure     403      {string}  string  "forbidden"
// @Failure     500      {string}  string  "server error"
// @Router      /api/v1/polls/{id}/status [patch]
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

	if err := h.pollSvc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, "bad status", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
