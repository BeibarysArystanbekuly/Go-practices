package api

import (
	"encoding/json"
	"net/http"
)

type updateRoleRequest struct {
	Role string `json:"role"`
}

// @Summary     List users
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200  {array}   user.User
// @Failure     500  {string}  string  "server error"
// @Router      /api/v1/users [get]
func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userSvc.List(r.Context())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// @Summary     Update user role
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Param       id       path     int64              true  "User ID"
// @Param       request  body     updateRoleRequest  true  "New role"
// @Success     204
// @Failure     400      {string}  string  "invalid id or body"
// @Failure     500      {string}  string  "server error"
// @Router      /api/v1/users/{id}/role [patch]
func (h *Handler) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.userSvc.UpdateRole(r.Context(), id, req.Role); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
