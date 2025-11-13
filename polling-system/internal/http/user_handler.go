package api

import (
    "encoding/json"
    "net/http"
)

type updateRoleRequest struct {
    Role string `json:"role"`
}

func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.userSvc.List()
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, users)
}

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

    if err := h.userSvc.UpdateRole(id, req.Role); err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
