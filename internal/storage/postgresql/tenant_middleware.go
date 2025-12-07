package postgresql

import (
	"gorm.io/gorm"
	"github.com/turtacn/QuantaID/internal/multitenant"
)

// TenantScopeMiddleware returns a GORM middleware that injects tenant scope.
func TenantScopeMiddleware() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement.Context == nil {
			return
		}

		tenantID, ok := multitenant.GetTenantID(db.Statement.Context)
		if !ok {
			return
		}

		// Check if the model has a TenantID field or if we are querying a table that should be isolated.
		// For simplicity, let's check if the Statement.Schema (model) has a field named "TenantID".
		// Note: Schema might be nil if we are running raw SQL or table scan without model.

		// Strategy:
		// 1. If explicit SQL, we assume RLS handles it (if configured).
		// 2. If GORM query, we add WHERE tenant_id = ?.
		// Note: RLS handles enforcement at DB level. Middleware adds filter for correctness/performance (index usage)
		// and to handle cases where RLS might not be enough or for in-memory enforcement if needed.
		// BUT, RLS is the "Target" of this phase.
		// If RLS is enabled, we need to set the session variable.

		// The prompt says: "P1-T5: Create tenant_middleware.go, 在GORM查询前自动注入tenant_id过滤条件"
		// It mentions injecting filter condition.
		// ALSO, "SetTenantContext" needs to be called.

		// If we rely on RLS, we MUST set `app.current_tenant`.
		// If we rely on WHERE clause, we append it.
		// Let's do both for robustness? Or follow prompt strictly?
		// "P1-T5: ... automatically inject tenant_id filter condition" -> This usually means Where("tenant_id = ?")

		// However, "P1-T1: ... EnableRowLevelSecurity ... SetTenantContext"
		// So we likely need to set the postgres variable too.

		// Setting session variable in a callback is tricky because GORM might borrow a connection only for execution.
		// We need to ensure `SET LOCAL` is executed on the *same* connection.
		// This is hard with simple GORM callbacks unless we wrap the execution in a transaction,
		// or use GORM's connection handling if available.

		// But wait, "P1-T5" says "inject tenant_id filter condition". This implies `db.Where`.
		// If RLS is enabled, `db.Where` is redundant for security but good for query optimizer.
		// BUT RLS requires `app.current_tenant` to be set.
		// If we don't set it, RLS will block everything (default deny) or error out.

		// How to set `app.current_tenant` in GORM?
		// We can't easily do it in a `Before` callback for a single query unless we force a transaction
		// or if we trust `SET LOCAL` works with the connection pool behavior (it doesn't across queries).
		// Wait, `SET LOCAL` is transaction scoped. `SET` is session scoped.

		// If we are not in a transaction, `SET LOCAL` applies to the transaction block implicitly created by the statement?
		// No, for a single statement it might work if prepended?
		// Best practice for RLS with connection pooling is often to use a transaction or
		// `db.Transaction(func(tx *gorm.DB) ...)` block where we set it first.

		// But the prompt asks for a middleware.
		// Let's implement the filter injection as requested.
		// And maybe we assume the connection is managed or RLS variable setting happens elsewhere (e.g. at connection checkout)?
		// Or maybe the middleware *only* does the `Where` clause as per P1-T5 description?
		// "P1-T5: Create tenant_middleware.go, automatically inject tenant_id filter condition before GORM query"
		// I will implement the WHERE clause injection.

		schema := db.Statement.Schema
		if schema != nil {
			// Check if field exists
			if field := schema.LookUpField("TenantID"); field != nil {
				db.Where("tenant_id = ?", tenantID)
			} else if field := schema.LookUpField("tenant_id"); field != nil {
				// GORM might map it differently
				db.Where("tenant_id = ?", tenantID)
			}
		} else {
			// If no schema, we might check table name?
			// Or just safer to not add WHERE if we are unsure.
			// Let's try to add it if table is known to be multi-tenant.
			// users, groups, applications
			table := db.Statement.Table
			if table == "users" || table == "groups" || table == "applications" {
				db.Where("tenant_id = ?", tenantID)
			}
		}
	}
}

// RegisterTenantCallbacks registers the tenant middleware.
func RegisterTenantCallbacks(db *gorm.DB) {
	db.Callback().Query().Before("gorm:query").Register("tenant_scope", TenantScopeMiddleware())
	db.Callback().Create().Before("gorm:create").Register("tenant_scope", TenantScopeMiddleware())
	db.Callback().Update().Before("gorm:update").Register("tenant_scope", TenantScopeMiddleware())
	db.Callback().Delete().Before("gorm:delete").Register("tenant_scope", TenantScopeMiddleware())
	db.Callback().Row().Before("gorm:row").Register("tenant_scope", TenantScopeMiddleware())
}
