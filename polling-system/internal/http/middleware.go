package api

import (
    "context"
    "net/http"
    "strings"

    jwtpkg "polling-system/internal/platform/jwt"
)

type ctxKey string

const (
    ctxKeyUserID ctxKey = "user_id"
    ctxKeyRole   ctxKey = "role"
)

func AuthMiddleware(jm *jwtpkg.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            h := r.Header.Get("Authorization")
            if h == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }

            parts := strings.SplitN(h, " ", 2)
            if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
                http.Error(w, "invalid authorization header", http.StatusUnauthorized)
                return
            }

            claims, err := jm.Parse(parts[1])
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
            ctx = context.WithValue(ctx, ctxKeyRole, claims.Role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctxRole, ok := r.Context().Value(ctxKeyRole).(string)
            if !ok || ctxRole != role {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func userIDFromCtx(r *http.Request) int64 {
    if v := r.Context().Value(ctxKeyUserID); v != nil {
        if id, ok := v.(int64); ok {
            return id
        }
    }
    return 0
}
