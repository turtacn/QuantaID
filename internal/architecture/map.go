package architecture

// Layer and Capability definitions
type Layer string
type Capability string

const (
	LayerPresentation Layer = "presentation"
	LayerGateway      Layer = "gateway"
	LayerAppService   Layer = "app_service"
	LayerDomain       Layer = "domain"
	LayerInfra        Layer = "infra"
)

const (
	CapabilityAuthMultiProtocol Capability = "auth.multi_protocol"
	CapabilityAuthMFA           Capability = "auth.mfa.basic"
	CapabilityIdentityLifecycle Capability = "identity.lifecycle.basic"
	CapabilityConnectorLDAP     Capability = "connector.ldap.basic"
	CapabilityAuditLog          Capability = "audit.log.basic"
	CapabilityMetricsPrometheus Capability = "metrics.prometheus.basic"
)

type CapabilityMapping struct {
	Capability Capability
	Layer      Layer
	Packages   []string
	Status     string // "planned" / "partial" / "done"
}

// DefaultMappings is a manually maintained list of capability mappings.
var DefaultMappings = []CapabilityMapping{
	{
		Capability: CapabilityAuthMultiProtocol,
		Layer:      LayerAppService,
		Packages: []string{
			"internal/services/auth",
			"internal/server/http/handlers/auth.go",
		},
		Status: "partial",
	},
	{
		Capability: CapabilityAuthMFA,
		Layer:      LayerDomain,
		Packages: []string{
			"internal/domain/auth/mfa_policy.go",
		},
		Status: "partial",
	},
	{
		Capability: CapabilityIdentityLifecycle,
		Layer:      LayerDomain,
		Packages: []string{
			"internal/domain/identity",
		},
		Status: "done",
	},
	{
		Capability: CapabilityIdentityLifecycle,
		Layer:      LayerInfra,
		Packages: []string{
			"internal/storage/postgresql/identity.go",
		},
		Status: "done",
	},
	{
		Capability: CapabilityAuditLog,
		Layer:      LayerDomain,
		Packages: []string{
			"internal/domain/auth/repository.go", // AuditLogRepository is here
		},
		Status: "partial",
	},
	{
		Capability: CapabilityMetricsPrometheus,
		Layer:      LayerInfra,
		Packages: []string{
			"internal/metrics/prometheus.go",
			"internal/storage/postgresql/prometheus.go",
		},
		Status: "done",
	},
	{
		Capability: CapabilityConnectorLDAP,
		Layer:      LayerInfra,
		Packages: []string{
			"pkg/plugins/connectors/ldap",
		},
		Status: "planned",
	},
}

// FindMappingsByCapability finds all mappings for a given capability.
func FindMappingsByCapability(c Capability) []CapabilityMapping {
	var res []CapabilityMapping
	for _, m := range DefaultMappings {
		if m.Capability == c {
			res = append(res, m)
		}
	}
	return res
}

// FindMappingsByPackage finds all mappings for a given package.
func FindMappingsByPackage(pkg string) []CapabilityMapping {
	var res []CapabilityMapping
	for _, m := range DefaultMappings {
		for _, p := range m.Packages {
			if p == pkg {
				res = append(res, m)
				break
			}
		}
	}
	return res
}
