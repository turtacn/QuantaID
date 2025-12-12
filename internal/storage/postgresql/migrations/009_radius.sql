-- RADIUS Client Table
CREATE TABLE radius_clients (
    id              VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    ip_address      VARCHAR(45) NOT NULL,
    secret          VARCHAR(256) NOT NULL,
    tenant_id       VARCHAR(64) NOT NULL,
    enabled         BOOLEAN DEFAULT true,
    vendor_type     VARCHAR(32) DEFAULT 'generic',
    attributes      JSONB DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_radius_clients_ip ON radius_clients(ip_address);
CREATE INDEX idx_radius_clients_tenant ON radius_clients(tenant_id);

-- RADIUS Accounting Table
CREATE TABLE radius_accounting (
    id              VARCHAR(64) PRIMARY KEY,
    session_id      VARCHAR(128) NOT NULL,
    user_id         VARCHAR(64),
    username        VARCHAR(256),
    nas_identifier  VARCHAR(128),
    nas_ip_address  VARCHAR(45),
    nas_port        INT,
    status_type     INT NOT NULL,
    session_time    INT DEFAULT 0,
    input_octets    BIGINT DEFAULT 0,
    output_octets   BIGINT DEFAULT 0,
    input_packets   BIGINT DEFAULT 0,
    output_packets  BIGINT DEFAULT 0,
    terminate_cause INT,
    framed_ip       VARCHAR(45),
    called_station  VARCHAR(128),
    calling_station VARCHAR(128),
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_radius_acct_session ON radius_accounting(session_id);
CREATE INDEX idx_radius_acct_user ON radius_accounting(user_id);
CREATE INDEX idx_radius_acct_time ON radius_accounting(created_at);
