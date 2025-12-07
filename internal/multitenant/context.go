package multitenant

import (
	"context"
	"fmt"
)

type tenantKey struct{}

// WithTenantID returns a new context with the given tenant ID.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// GetTenantID retrieves the tenant ID from the context.
// It returns the tenant ID and a boolean indicating if it was found.
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantKey{}).(string)
	return tenantID, ok
}

// MustGetTenantID retrieves the tenant ID from the context or panics if not found.
// Use this only when you are certain a tenant ID must be present.
func MustGetTenantID(ctx context.Context) string {
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		panic(fmt.Sprintf("tenant ID not found in context"))
	}
	return tenantID
}
