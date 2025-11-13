package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// HTTPMetricsMiddleware is a middleware for recording HTTP-level Prometheus metrics.
type HTTPMetricsMiddleware struct {
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec
}

// NewHTTPMetricsMiddleware creates a new HTTPMetricsMiddleware.
func NewHTTPMetricsMiddleware(reg prometheus.Registerer) *HTTPMetricsMiddleware {
	return &HTTPMetricsMiddleware{
		requestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quantaid_http_request_duration_seconds",
				Help:    "Histogram of HTTP request latencies.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		requestTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "quantaid_http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
	}
}

// Execute is the middleware handler function.
func (m *HTTPMetricsMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		path, _ := mux.CurrentRoute(r).GetPathTemplate()
		if path == "" {
			path = "not_found"
		}
		method := r.Method
		status := strconv.Itoa(rw.statusCode)
		duration := time.Since(start).Seconds()

		m.requestDuration.WithLabelValues(method, path, status).Observe(duration)
		m.requestTotal.WithLabelValues(method, path, status).Inc()
	})
}
