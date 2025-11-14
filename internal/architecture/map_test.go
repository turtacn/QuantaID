package architecture

import (
	"testing"
)

func Test_DefaultMappings_CoversExpectedCapabilities(t *testing.T) {
	expectedCapabilities := []Capability{
		CapabilityAuthEngineCore,
		CapabilityAuthMFACore,
		CapabilityIdentityLifecycleCore,
		CapabilityIdentitySyncLDAP,
		CapabilityAuthzPolicyEngine,
		CapabilityAuditCore,
		CapabilityMetricsHTTP,
		CapabilityPlatformDevCenter,
	}

	for _, cap := range expectedCapabilities {
		mappings := FindMappingsByCapability(cap)
		if len(mappings) == 0 {
			t.Errorf("expected capability %q to have at least one mapping, but found none", cap)
		}
	}
}

func Test_FindMappingsByPackage_ReturnsCapabilities(t *testing.T) {
	pkg := "internal/services/identity"
	mappings := FindMappingsByPackage(pkg)
	if len(mappings) == 0 {
		t.Errorf("expected package %q to have at least one mapping, but found none", pkg)
	}

	found := false
	for _, m := range mappings {
		if m.Capability == CapabilityIdentityLifecycleCore {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected package %q to have capability %q, but it was not found", pkg, CapabilityIdentityLifecycleCore)
	}
}
