package api

import (
	"encoding/json"
	"net/http"

	"polling-system/internal/platform/apperr"
)

type updateRoleRequest struct {
	Role string `json:"role"`
}

// @Summary     List users
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200  {array}   user.User
// @Failure     500  {object}  map[string]string  "server error"
// @Router      /api/v1/users [get]
func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userSvc.List(r.Context())
	if err != nil {
		errorResponse(w, err)
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
// @Failure     400      {object}  map[string]string  "invalid id or body"
// @Failure     404      {object}  map[string]string  "not found"
// @Failure     500      {object}  map[string]string  "server error"
// @Router      /api/v1/users/{id}/role [patch]
func (h *Handler) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid id", err))
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid body", err))
		return
	}
	if req.Role != "admin" && req.Role != "user" {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid role", nil))
		return
	}

	if err := h.userSvc.UpdateRole(r.Context(), id, req.Role); err != nil {
		errorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary     Deactivate user
// @Tags        users
// @Security    BearerAuth
// @Param       id   path  int64  true  "User ID"
// @Success     204
// @Failure     400  {object}  map[string]string  "invalid id"
// @Failure     401  {object}  map[string]string  "unauthorized"
// @Failure     403  {object}  map[string]string  "forbidden"
// @Failure     404  {object}  map[string]string  "not found"
// @Failure     500  {object}  map[string]string  "server error"
// @Router      /api/v1/users/{id}/deactivate [patch]
func (h *Handler) handleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid id", err))
		return
	}

	if err := h.userSvc.Deactivate(r.Context(), id); err != nil {
		errorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
