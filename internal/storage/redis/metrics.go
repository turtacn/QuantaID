package redis

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var registerMetricsOnce sync.Once

// Metrics holds the Prometheus metrics for Redis operations.
type Metrics struct {
	Commands       *prometheus.CounterVec
	CommandLatency *prometheus.HistogramVec
	Errors         *prometheus.CounterVec
	PoolHits       prometheus.Counter
	PoolMisses     prometheus.Counter
	PoolTimeouts   prometheus.Counter
	PoolTotalConns prometheus.Gauge
	PoolIdleConns  prometheus.Gauge
	PoolStaleConns prometheus.Gauge
}

// NewMetrics creates a new Metrics struct and registers the Prometheus collectors.
func NewMetrics(namespace string, reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		Commands: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "commands_total",
				Help:      "Total number of Redis commands executed.",
			},
			[]string{"command"},
		),
		CommandLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "command_latency_seconds",
				Help:      "Latency of Redis commands in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"command"},
		),
		Errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "errors_total",
				Help:      "Total number of Redis errors.",
			},
			[]string{"command"},
		),
		PoolHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "hits_total",
				Help:      "Total number of connection pool hits.",
			},
		),
		PoolMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "misses_total",
				Help:      "Total number of connection pool misses.",
			},
		),
		PoolTimeouts: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "timeouts_total",
				Help:      "Total number of connection pool timeouts.",
			},
		),
		PoolTotalConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "total_connections",
				Help:      "Total number of connections in the pool.",
			},
		),
		PoolIdleConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "idle_connections",
				Help:      "Number of idle connections in the pool.",
			},
		),
		PoolStaleConns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "redis_pool",
				Name:      "stale_connections",
				Help:      "Number of stale connections in the pool.",
			},
		),
	}

	registerMetricsOnce.Do(func() {
		reg.MustRegister(
			m.Commands,
			m.CommandLatency,
			m.Errors,
			m.PoolHits,
			m.PoolMisses,
			m.PoolTimeouts,
			m.PoolTotalConns,
			m.PoolIdleConns,
			m.PoolStaleConns,
		)
	})

	return m
}
