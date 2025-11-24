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

type authResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

// @Summary     Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request  body      authRequest  true  "User credentials"
// @Success     201      {object}  authResponse
// @Failure     400      {string}  string  "invalid body or email taken"
// @Failure     500      {string}  string  "server error"
// @Router      /api/v1/auth/register [post]
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

// @Summary     Login user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request  body      authRequest  true  "User credentials"
// @Success     200      {object}  authResponse
// @Failure     400      {string}  string  "invalid body"
// @Failure     401      {string}  string  "invalid credentials"
// @Failure     500      {string}  string  "server error"
// @Router      /api/v1/auth/login [post]
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
