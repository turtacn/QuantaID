package oauth

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// OauthTokensIssuedTotal is a counter for the total number of OAuth tokens issued.
	OauthTokensIssuedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oauth_tokens_issued_total",
			Help: "Total number of OAuth tokens issued.",
		},
		[]string{"grant_type", "tenant_id", "client_id"},
	)

	// PKCEValidationsTotal is a counter for the number of PKCE validations.
	PKCEValidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oauth_pkce_validations_total",
			Help: "Total number of PKCE validations.",
		},
		[]string{"result"}, // "success" or "failure" reason
	)

	// DeviceFlowActivationsTotal is a counter for device flow activations.
	DeviceFlowActivationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oauth_device_flow_activations_total",
			Help: "Total number of device flow activations.",
		},
		[]string{"status"}, // "authorized", "denied", "expired"
	)

	// TenantQuotaUsage is a gauge for tenant quota usage.
	TenantQuotaUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_quota_usage",
			Help: "Current tenant quota usage.",
		},
		[]string{"tenant_id", "resource_type"}, // "clients", "users", "tokens_per_hour"
	)
)
