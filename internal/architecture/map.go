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
	CapabilityAuthMultiProtocol   Capability = "auth.multi_protocol"
	CapabilityAuthEngineCore      Capability = "auth.engine.core"
	CapabilityAuthMFACore         Capability = "auth.mfa.core"
	CapabilityAuthzPolicyEngine   Capability = "authz.policy.engine"
	CapabilityIdentityLifecycleCore Capability = "identity.lifecycle.core"
	CapabilityIdentitySyncLDAP      Capability = "identity.sync.ldap"
	CapabilityAuditCore           Capability = "audit.core"
	CapabilityMetricsHTTP         Capability = "metrics.http"
	CapabilityPlatformDevCenter   Capability = "platform.devcenter"
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
		Capability: CapabilityAuthEngineCore,
		Layer:      LayerDomain,
		Packages: []string{
			"pkg/auth/engine.go",
			"internal/services/auth/service.go",
		},
		Status: "partial",
	},
	{
		Capability: CapabilityAuthMFACore,
		Layer:      LayerAppService,
		Packages: []string{
			"pkg/auth/mfa/manager.go",
			"pkg/plugins/mfa/totp/",
		},
		Status: "partial",
	},
	{
		Capability: CapabilityIdentityLifecycleCore,
		Layer:      LayerDomain,
		Packages: []string{
			"internal/domain/identity",
			"internal/services/identity",
		},
		Status: "done",
	},
	{
		Capability: CapabilityIdentityLifecycleCore,
		Layer:      LayerInfra,
		Packages: []string{
			"internal/storage/memory/identity_memory_repository.go",
			"internal/storage/postgresql/identity_repository.go",
		},
		Status: "done",
	},
	{
		Capability: CapabilityIdentitySyncLDAP,
		Layer:      LayerAppService,
		Packages: []string{
			"internal/services/sync/ldap_sync_service.go",
			"pkg/plugins/connectors/ldap/",
		},
		Status: "planned",
	},
	{
		Capability: CapabilityAuditCore,
		Layer:      LayerAppService,
		Packages: []string{
			"internal/audit/",
			"internal/services/audit/",
		},
		Status: "partial",
	},
	{
		Capability: CapabilityMetricsHTTP,
		Layer:      LayerInfra,
		Packages: []string{
			"internal/metrics/http_middleware.go",
			"pkg/observability/metrics.go",
		},
		Status: "done",
	},
	{
		Capability: CapabilityAuthzPolicyEngine,
		Layer:      LayerDomain,
		Packages: []string{
			"internal/services/authorization/evaluator.go",
		},
		Status: "done",
	},
	{
		Capability: CapabilityPlatformDevCenter,
		Layer:      LayerPresentation,
		Packages: []string{
			"internal/services/platform/",
			"internal/server/http/handlers/devcenter.go",
		},
		Status: "done",
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
