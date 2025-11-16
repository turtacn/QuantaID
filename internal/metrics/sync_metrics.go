package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	ldapSyncDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ldap_sync_duration_seconds",
			Help:    "LDAP sync duration in seconds",
			Buckets: []float64{10, 30, 60, 120, 300, 600},
		},
		[]string{"source_id", "sync_type"},
	)

	ldapSyncErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ldap_sync_errors_total",
			Help: "LDAP sync errors total",
		},
		[]string{"source_id", "error_type"},
	)

	identityDuplicates = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_duplicates_detected_total",
			Help: "Identity duplicates detected total",
		},
		[]string{"source_id", "match_field"},
	)

	syncLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ldap_sync_lag_seconds",
			Help: "Incremental sync lag in seconds",
		},
		[]string{"source_id"},
	)
)

func init() {
	prometheus.MustRegister(ldapSyncDuration)
	prometheus.MustRegister(ldapSyncErrors)
	prometheus.MustRegister(identityDuplicates)
	prometheus.MustRegister(syncLag)
}

type SyncMetrics struct {
	sourceID string
}

func NewSyncMetrics(sourceID string) *SyncMetrics {
	return &SyncMetrics{sourceID: sourceID}
}

func (m *SyncMetrics) RecordFullSyncDuration(duration time.Duration) {
	ldapSyncDuration.WithLabelValues(m.sourceID, "full").Observe(duration.Seconds())
}

func (m *SyncMetrics) RecordSyncError(errType string) {
	ldapSyncErrors.WithLabelValues(m.sourceID, errType).Inc()
}
