package multitenant

import (
	"context"
	"database/sql"
	"net/http"
)

// TenantIsolator provides middleware for enforcing tenant isolation.
type TenantIsolator struct {
	config      TenantConfig
	quotaChecker QuotaChecker
}

// TenantConfig holds the configuration for multi-tenancy.
type TenantConfig struct {
	EnableRowLevelSecurity bool
	DefaultQuotas          TenantQuotas
}

// TenantQuotas defines the resource quotas for a tenant.
type TenantQuotas struct {
	MaxClients       int `yaml:"maxClients"`
	MaxTokensPerHour int `yaml:"maxTokensPerHour"`
	MaxUsers         int `yaml:"maxUsers"`
}

// QuotaChecker defines the interface for checking tenant quotas.
type QuotaChecker interface {
	CheckQuota(ctx context.Context, tenantID string) error
}

// NewTenantIsolator creates a new TenantIsolator.
func NewTenantIsolator(config TenantConfig, quotaChecker QuotaChecker) *TenantIsolator {
	return &TenantIsolator{config: config, quotaChecker: quotaChecker}
}

// Middleware is an HTTP middleware that extracts the tenant ID from the request
// and injects it into the context.
func (t *TenantIsolator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := t.extractTenantID(r)
		if tenantID == "" {
			http.Error(w, "Invalid tenant", http.StatusForbidden)
			return
		}

		if err := t.quotaChecker.CheckQuota(r.Context(), tenantID); err != nil {
			http.Error(w, "Quota exceeded", http.StatusTooManyRequests)
			return
		}

		ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (t *TenantIsolator) extractTenantID(r *http.Request) string {
	// TODO: Implement logic to extract tenant ID from hostname, header, or token.
	return r.Header.Get("X-Tenant-ID")
}

// EnableRowLevelSecurity enables row-level security in the database for all tenant-scoped tables.
func (t *TenantIsolator) EnableRowLevelSecurity(ctx context.Context, db *sql.DB) error {
	// TODO: This is a placeholder. The actual implementation would read migration
	// files or have a more robust way of applying these policies.
	queries := []string{
		"ALTER TABLE oauth_clients ENABLE ROW LEVEL SECURITY;",
		`CREATE POLICY tenant_isolation ON oauth_clients USING (tenant_id = current_setting('app.current_tenant_id')::uuid);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// SetTenantContext sets the current tenant ID for the database session.
func (t *TenantIsolator) SetTenantContext(ctx context.Context, db *sql.DB, tenantID string) error {
	_, err := db.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
	return err
}
