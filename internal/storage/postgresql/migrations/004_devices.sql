-- Up Migration
CREATE TABLE devices (
    id              VARCHAR(64) PRIMARY KEY,
    tenant_id       VARCHAR(64) NOT NULL,
    user_id         VARCHAR(64),
    fingerprint     VARCHAR(256) NOT NULL,
    fingerprint_raw JSONB,
    device_name     VARCHAR(128),
    device_type     VARCHAR(32),
    os              VARCHAR(64),
    browser         VARCHAR(64),
    trust_score     INT DEFAULT 0,
    last_ip         VARCHAR(45),
    last_location   VARCHAR(128),
    last_active_at  TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    bound_at        TIMESTAMP,
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_devices_fingerprint ON devices(fingerprint);
CREATE INDEX idx_devices_user ON devices(user_id);
CREATE INDEX idx_devices_tenant ON devices(tenant_id);

-- Enable RLS
ALTER TABLE devices ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON devices
    USING (tenant_id = current_setting('app.current_tenant', true));

-- Down Migration
DROP TABLE IF EXISTS devices;
