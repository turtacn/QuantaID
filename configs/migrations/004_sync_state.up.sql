-- configs/migrations/004_sync_state.up.sql
CREATE TABLE sync_state (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id       VARCHAR(64) NOT NULL,
    sync_type       VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP,
    last_change_num BIGINT,
    progress        JSONB,
    error_message   TEXT,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sync_state_source ON sync_state(source_id, started_at DESC);
