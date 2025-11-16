-- configs/migrations/005_identity_conflicts.up.sql
CREATE TABLE identity_conflicts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_a_id UUID NOT NULL,
    identity_b_id UUID NOT NULL,
    conflict_type VARCHAR(50),
    resolution    VARCHAR(20),
    resolved_by   UUID,
    resolved_at   TIMESTAMP,
    details       JSONB,
    created_at    TIMESTAMP DEFAULT NOW()
);
