package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	jwtpkg "polling-system/internal/platform/jwt"
)

type ctxKey string

const (
	ctxKeyUserID ctxKey = "user_id"
	ctxKeyRole   ctxKey = "role"
)

var metrics = newMetricsCollector()

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
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RateLimitPerUser(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if id := userIDFromCtx(r); id > 0 {
				key = fmt.Sprintf("user:%d", id)
			}
			if !limiter.allow(key) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func MetricsMiddleware(next http.Handler) http.Handler {
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

		metrics.record(route, r.Method, status, time.Since(start))
	})
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	snapshot := metrics.snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for key, m := range snapshot {
		fmt.Fprintf(w, "polling_requests_total{path=\"%s\",method=\"%s\",status=\"%d\"} %d\n",
			key.Path, key.Method, key.Status, m.Count)
		fmt.Fprintf(w, "polling_request_duration_ms_sum{path=\"%s\",method=\"%s\",status=\"%d\"} %d\n",
			key.Path, key.Method, key.Status, m.Latency/time.Millisecond)
		fmt.Fprintf(w, "polling_request_duration_ms_count{path=\"%s\",method=\"%s\",status=\"%d\"} %d\n",
			key.Path, key.Method, key.Status, m.Count)
	}
}

type metricKey struct {
	Path   string
	Method string
	Status int
}

type metricEntry struct {
	Count   uint64
	Latency time.Duration
}

type metricsCollector struct {
	mu      sync.Mutex
	entries map[metricKey]metricEntry
}

func newMetricsCollector() *metricsCollector {
	return &metricsCollector{
		entries: make(map[metricKey]metricEntry),
	}
}

func (m *metricsCollector) record(path, method string, status int, dur time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := metricKey{Path: path, Method: method, Status: status}
	entry := m.entries[key]
	entry.Count++
	entry.Latency += dur
	m.entries[key] = entry
}

func (m *metricsCollector) snapshot() map[metricKey]metricEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	res := make(map[metricKey]metricEntry, len(m.entries))
	for k, v := range m.entries {
		res[k] = v
	}
	return res
}

type rateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	hits   map[string]rateState
}

type rateState struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		window: window,
		limit:  limit,
		hits:   make(map[string]rateState),
	}
}

func (r *rateLimiter) allow(key string) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.hits[key]
	if now.After(state.reset) {
		state = rateState{count: 0, reset: now.Add(r.window)}
	}
	if state.count >= r.limit {
		return false
	}
	state.count++
	r.hits[key] = state
	return true
}
