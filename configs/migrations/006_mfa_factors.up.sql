-- configs/migrations/006_mfa_factors.up.sql
CREATE TABLE mfa_factors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            VARCHAR(20) NOT NULL, -- 'totp' | 'sms' | 'webauthn'
    status          VARCHAR(20) NOT NULL, -- 'pending' | 'active' | 'disabled'
    credential_id   VARCHAR(255),         -- WebAuthn credential ID
    public_key      BYTEA,                -- WebAuthn公钥
    secret          TEXT,                 -- TOTP密钥 (加密)
    phone_number    VARCHAR(20),          -- SMS手机号
    backup_codes    JSONB,                -- 备用恢复码
    last_used_at    TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_mfa_factors_user ON mfa_factors(user_id, status);

CREATE TABLE mfa_verification_logs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factor_id    UUID NOT NULL REFERENCES mfa_factors(id),
    success      BOOLEAN NOT NULL,
    error_reason TEXT,
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMP DEFAULT NOW()
);
