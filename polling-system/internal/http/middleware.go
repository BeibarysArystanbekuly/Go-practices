package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"

	"polling-system/internal/metrics"
	jwtpkg "polling-system/internal/platform/jwt"
)

type ctxKey string

const (
	ctxKeyUserID ctxKey = "user_id"
	ctxKeyRole   ctxKey = "role"
)

var slogLogger = slog.Default()

func SetLogger(l *slog.Logger) {
	if l != nil {
		slogLogger = l
	}
}

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

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RateLimitVotes(r rate.Limit, burst int) func(http.Handler) http.Handler {
	limiter := newIPRateLimiter(r, burst, 10*time.Minute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !limiter.allow(ip) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(rw, r)

		status := rw.Status()
		if status == 0 {
			status = http.StatusOK
		}
		route := r.URL.Path
		if rc := chi.RouteContext(r.Context()); rc != nil && rc.RoutePattern() != "" {
			route = rc.RoutePattern()
		}

		metrics.IncRequest(r.Method, route, status)

		slogLogger.Info("request",
			"method", r.Method,
			"path", route,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	lastSeen map[string]time.Time
	limit    rate.Limit
	burst    int
	entryTTL time.Duration
}

func newIPRateLimiter(limit rate.Limit, burst int, entryTTL time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		lastSeen: make(map[string]time.Time),
		limit:    limit,
		burst:    burst,
		entryTTL: entryTTL,
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, ts := range l.lastSeen {
		if now.Sub(ts) > l.entryTTL {
			delete(l.limiters, key)
			delete(l.lastSeen, key)
		}
	}

	if limiter, ok := l.limiters[ip]; ok {
		l.lastSeen[ip] = now
		return limiter
	}
	limiter := rate.NewLimiter(l.limit, l.burst)
	l.limiters[ip] = limiter
	l.lastSeen[ip] = now
	return limiter
}

func (l *ipRateLimiter) allow(ip string) bool {
	limiter := l.getLimiter(ip)
	return limiter.Allow()
}

func clientIP(r *http.Request) string {
	if xfwd := r.Header.Get("X-Forwarded-For"); xfwd != "" {
		parts := strings.Split(xfwd, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
