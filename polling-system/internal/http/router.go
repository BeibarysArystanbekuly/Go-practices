package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"

    "polling-system/internal/domain/poll"
    "polling-system/internal/domain/user"
    "polling-system/internal/domain/vote"
    jwtpkg "polling-system/internal/platform/jwt"
    "polling-system/internal/worker"
)

type Handler struct {
    userSvc *user.Service
    pollSvc *poll.Service
    voteSvc *vote.Service
    jwtMgr  *jwtpkg.Manager
    voteCh  chan<- worker.VoteEvent
}

func NewRouter(
    userSvc *user.Service,
    pollSvc *poll.Service,
    voteSvc *vote.Service,
    jwtMgr *jwtpkg.Manager,
    voteCh chan<- worker.VoteEvent,
) http.Handler {
    h := &Handler{
        userSvc: userSvc,
        pollSvc: pollSvc,
        voteSvc: voteSvc,
        jwtMgr:  jwtMgr,
        voteCh:  voteCh,
    }

    r := chi.NewRouter()

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    })

    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/auth/register", h.handleRegister)
        r.Post("/auth/login", h.handleLogin)

        r.Group(func(r chi.Router) {
            r.Use(AuthMiddleware(jwtMgr))

            r.Get("/polls", h.handleListPolls)
            r.Get("/polls/{id}", h.handleGetPoll)
            r.Post("/polls/{id}/vote", h.handleVote)
            r.Get("/polls/{id}/results", h.handlePollResults)

            r.Group(func(r chi.Router) {
                r.Use(RequireRole("admin"))
                r.Post("/polls", h.handleCreatePoll)
                r.Patch("/polls/{id}/status", h.handleUpdatePollStatus)
                r.Get("/users", h.handleListUsers)
                r.Patch("/users/{id}/role", h.handleUpdateUserRole)
            })
        })
    })

    return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func parseIDParam(r *http.Request, name string) (int64, error) {
    idStr := chi.URLParam(r, name)
    return strconv.ParseInt(idStr, 10, 64)
}

func parseTimePtr(s *string) *time.Time {
    if s == nil || *s == "" {
        return nil
    }
    t, err := time.Parse(time.RFC3339, *s)
    if err != nil {
        return nil
    }
    return &t
}
