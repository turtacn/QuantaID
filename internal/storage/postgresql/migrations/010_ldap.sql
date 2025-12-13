-- Migration for LDAP support
-- Currently just a placeholder for future LDAP specific configurations
-- or mapping tables if complex mappings are needed.

CREATE TABLE IF NOT EXISTS ldap_mappings (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    resource_type VARCHAR(32) NOT NULL, -- user, group
    ldap_attribute VARCHAR(64) NOT NULL,
    app_attribute VARCHAR(64) NOT NULL,
    direction VARCHAR(16) DEFAULT 'both', -- import, export, both
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ldap_mappings_tenant ON ldap_mappings(tenant_id);
