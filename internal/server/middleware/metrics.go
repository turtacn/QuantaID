package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsMiddleware tracks HTTP request metrics.
type MetricsMiddleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewMetricsMiddleware creates a new MetricsMiddleware.
func NewMetricsMiddleware(reg prometheus.Registerer) *MetricsMiddleware {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"code", "method", "path"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "method", "path"},
	)

	reg.MustRegister(requestsTotal, requestDuration)

	return &MetricsMiddleware{
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
	}
}

// Execute wraps the next handler with metrics tracking.
func (m *MetricsMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		statusCode := strconv.Itoa(rw.StatusCode())
		m.requestsTotal.WithLabelValues(statusCode, r.Method, path).Inc()
		m.requestDuration.WithLabelValues(statusCode, r.Method, path).Observe(duration.Seconds())
	})
}
