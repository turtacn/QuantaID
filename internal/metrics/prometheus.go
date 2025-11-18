package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

var (
	// HTTPRequestsTotal is a counter for total HTTP requests.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantaid_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration is a histogram for HTTP request latencies.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantaid_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// DBQueriesTotal is a counter for total database queries.
	DBQueriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_db_queries_total",
			Help: "Total number of database queries executed",
		},
	)

	// CacheHitsTotal is a counter for total cache hits.
	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	// CacheMissesTotal is a counter for total cache misses.
	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	// OauthTokensIssuedTotal is a counter for the total number of OAuth tokens issued.
	OauthTokensIssuedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "quantaid_oauth_tokens_issued_total",
			Help: "Total number of OAuth tokens issued",
		},
	)

	// MFAVerificationsTotal is a counter for MFA verifications.
	MFAVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantaid_mfa_verifications_total",
			Help: "Total number of MFA verifications",
		},
		[]string{"status"}, // "success" or "failure"
	)

	// AuthRiskScore is a histogram for the risk scores of authentication attempts.
	AuthRiskScore = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "quantaid_auth_risk_score",
			Help:    "Distribution of risk scores for authentication attempts",
			Buckets: prometheus.LinearBuckets(0, 0.1, 10),
		},
	)

	// AuthRiskLevelTotal is a counter for the total number of authentication attempts per risk level.
	AuthRiskLevelTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantaid_auth_risk_level_total",
			Help: "Total number of authentication attempts per risk level",
		},
		[]string{"level"}, // "low", "medium", "high"
	)

	// MFAChallengeTotal is a counter for the total number of MFA challenges issued.
	MFAChallengeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantaid_auth_mfa_challenge_total",
			Help: "Total number of MFA challenges issued",
		},
		[]string{"provider"}, // "totp", "webauthn", etc.
	)

	// MFAFailureTotal is a counter for the total number of failed MFA verifications.
	MFAFailureTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantaid_auth_mfa_failure_total",
			Help: "Total number of failed MFA verifications",
		},
		[]string{"provider", "reason"}, // e.g., "invalid_code", "timeout"
	)
)

// Middleware returns a Gin middleware for recording Prometheus metrics.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath() // Use the route path template to avoid high cardinality

		HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// Handler returns a Gin handler for serving Prometheus metrics.
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// NewRedisMetrics creates a new set of metrics for Redis operations.
func NewRedisMetrics(namespace string) *redis.Metrics {
	return redis.NewMetrics(namespace, prometheus.DefaultRegisterer)
}
