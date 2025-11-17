-- +migrate Up
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY,
    timestamp     TIMESTAMPTZ NOT NULL,
    event_type    VARCHAR(50) NOT NULL,
    actor_id      VARCHAR(255),
    actor_type    VARCHAR(20), -- 'user' | 'service' | 'system'
    actor_name    TEXT,
    target_id     VARCHAR(255),
    target_type   VARCHAR(50),
    target_name   TEXT,
    action        VARCHAR(100) NOT NULL,
    result        VARCHAR(20) NOT NULL, -- 'success' | 'failure'
    metadata      JSONB,
    ip_address    INET,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Partitioning by range on the timestamp is a good strategy for large volumes of time-series data.
-- This example creates partitions for each month.
-- Note: You would need a mechanism to create new partitions automatically.
-- For simplicity in this example, we'll create a few manually.
CREATE TABLE audit_logs_y2025m11 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

CREATE TABLE audit_logs_y2025m12 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Indexes for common query patterns
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id, actor_type, timestamp DESC);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type, timestamp DESC);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_id, target_type, timestamp DESC);


-- +migrate Down
DROP TABLE IF EXISTS audit_logs;
