# Multi-Tenant Architecture

This document describes the multi-tenant architecture implemented in QuantaID.

## Overview

QuantaID supports multi-tenancy to allow multiple organizations to use the same instance while maintaining data isolation and resource fairness.

The architecture relies on two main pillars:
1. **Data Isolation**: Enforced via PostgreSQL Row-Level Security (RLS).
2. **Resource Management**: Enforced via a `QuotaManager`.

## Data Isolation (RLS)

We use PostgreSQL's native Row-Level Security to ensure that tenants can only access their own data.

### Implementation

- **Table Setup**: Tables like `users`, `groups`, and `applications` have a `tenant_id` column.
- **RLS Policy**: A policy is applied to these tables:
  ```sql
  CREATE POLICY tenant_isolation ON table_name
  USING (tenant_id = current_setting('app.current_tenant'))
  ```
- **Context Propagation**: The application middleware extracts the Tenant ID from the request (e.g., header, token) and sets the PostgreSQL session variable `app.current_tenant` for the duration of the request/transaction.

### Code Components

- `internal/multitenant/tenant_isolator.go`: Handles RLS enablement and session variable setting.
- `internal/storage/postgresql/tenant_middleware.go`: GORM middleware that adds `WHERE tenant_id = ?` clause for additional safety and query optimization.
- `internal/multitenant/context.go`: Helpers for passing Tenant ID through Go `context.Context`.

## Quota Management

To prevent "noisy neighbor" issues, we enforce quotas on key resources.

### Quotas

- **Max Users**: Maximum number of users a tenant can have.
- **Max Applications**: Maximum number of applications.
- **Max API Calls**: Daily limit on API calls.

### Configuration

Quotas are configured in `server.yaml` under `multitenant.quotas`.

Example:
```yaml
multitenant:
  enabled: true
  quotas:
    "tenant-1":
      max_users: 100
      max_api_calls_per_day: 10000
```

### Code Components

- `internal/multitenant/quota_manager.go`: Logic for checking quotas against the database (for counts) and Redis (for rate limits).

## Usage

### Development

To enable RLS in development:
1. Configure `multitenant.enabled: true`.
2. Ensure database user is not a superuser (superusers bypass RLS), or use `SET ROLE` to simulate a normal application user.

### Testing

Integration tests in `tests/integration/multitenant_isolation_test.go` verify that RLS correctly hides data between tenants.
