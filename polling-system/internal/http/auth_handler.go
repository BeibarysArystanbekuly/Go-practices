package api

import (
    "encoding/json"
    "net/http"
    "time"

    "polling-system/internal/domain/user"
)

type authRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
    var req authRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    u, err := h.userSvc.Register(req.Email, req.Password)
    if err != nil {
        if err == user.ErrEmailTaken {
            http.Error(w, "email taken", http.StatusBadRequest)
            return
        }
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    token, err := h.jwtMgr.Generate(u.ID, u.Role, 24*time.Hour)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusCreated, map[string]any{
        "user":  u,
        "token": token,
    })
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
    var req authRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    u, err := h.userSvc.Login(req.Email, req.Password)
    if err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    token, err := h.jwtMgr.Generate(u.ID, u.Role, 24*time.Hour)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "user":  u,
        "token": token,
    })
}
