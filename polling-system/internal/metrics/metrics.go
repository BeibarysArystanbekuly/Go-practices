package metrics

import (
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal *prometheus.CounterVec
	registerOnce      sync.Once
)

// Register initializes Prometheus metrics on the default registry.
func Register() {
	registerOnce.Do(func() {
		httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "polling",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests processed by the polling API.",
		}, []string{"method", "path", "status"})
	})
}

// IncRequest increments the http_requests_total counter with the given labels.
func IncRequest(method, path string, status int) {
	if httpRequestsTotal == nil {
		return
	}
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}
