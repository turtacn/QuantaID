-- 003_audit_logs.up.sql

CREATE TABLE IF NOT EXISTS audit_events (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    actor JSONB,
    target JSONB,
    result VARCHAR(255),
    metadata JSONB,
    ip_address VARCHAR(255),
    user_agent TEXT,
    trace_id VARCHAR(255),
    category VARCHAR(255),
    user_id VARCHAR(255),
    ip VARCHAR(255),
    details JSONB,
    resource VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_events(timestamp);
