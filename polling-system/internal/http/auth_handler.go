package api

import (
	"encoding/json"
	"net/http"
	"time"

	"polling-system/internal/domain/user"
	"polling-system/internal/platform/apperr"
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
// @Failure     400      {object}  map[string]string  "invalid body or email taken"
// @Failure     500      {object}  map[string]string  "server error"
// @Router      /api/v1/auth/register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid body", err))
		return
	}
	if req.Email == "" || req.Password == "" {
		errorResponse(w, apperr.BadRequest("invalid_input", "email and password required", nil))
		return
	}

	u, err := h.userSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		errorResponse(w, err)
		return
	}

	token, err := h.jwtMgr.Generate(u.ID, u.Role, 24*time.Hour)
	if err != nil {
		errorResponse(w, apperr.Internal("internal_error", "failed to generate token", err))
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
// @Failure     400      {object}  map[string]string  "invalid body"
// @Failure     401      {object}  map[string]string  "invalid credentials"
// @Failure     500      {object}  map[string]string  "server error"
// @Router      /api/v1/auth/login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, apperr.BadRequest("invalid_input", "invalid body", err))
		return
	}
	if req.Email == "" || req.Password == "" {
		errorResponse(w, apperr.BadRequest("invalid_input", "email and password required", nil))
		return
	}

	u, err := h.userSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		errorResponse(w, err)
		return
	}

	token, err := h.jwtMgr.Generate(u.ID, u.Role, 24*time.Hour)
	if err != nil {
		errorResponse(w, apperr.Internal("internal_error", "failed to generate token", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  u,
		"token": token,
	})
}
