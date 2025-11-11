package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	AuthRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_auth_requests_total",
			Help: "Total number of authentication requests.",
		},
	)
	AuthDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "quantaid_auth_duration_seconds",
			Help: "Authentication request duration in seconds.",
		},
	)
	ActiveSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "quantaid_active_sessions",
			Help: "Current number of active sessions.",
		},
	)
	RedisOperationsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_redis_operations_total",
			Help: "Total number of Redis operations.",
		},
	)
)

func init() {
	prometheus.MustRegister(AuthRequestsTotal)
	prometheus.MustRegister(AuthDurationSeconds)
	prometheus.MustRegister(ActiveSessions)
	prometheus.MustRegister(RedisOperationsTotal)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
