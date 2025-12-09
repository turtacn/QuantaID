-- Up Migration
CREATE TABLE user_profiles (
    id                  VARCHAR(64) PRIMARY KEY,
    user_id             VARCHAR(64) NOT NULL UNIQUE,
    tenant_id           VARCHAR(64) NOT NULL,
    behavior            JSONB DEFAULT '{}',
    risk                JSONB DEFAULT '{}',
    risk_score          INT DEFAULT 0,
    risk_level          VARCHAR(16) DEFAULT 'low',
    auto_tags           JSONB DEFAULT '[]',
    manual_tags         JSONB DEFAULT '[]',
    quality_score       INT DEFAULT 0,
    quality_details     JSONB DEFAULT '{}',
    last_activity_at    TIMESTAMP,
    last_risk_update_at TIMESTAMP,
    created_at          TIMESTAMP DEFAULT NOW(),
    updated_at          TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_profiles_tenant ON user_profiles(tenant_id);
CREATE INDEX idx_profiles_risk ON user_profiles(risk_score);
CREATE INDEX idx_profiles_quality ON user_profiles(quality_score);

-- GIN indexes for tag queries
CREATE INDEX idx_profiles_auto_tags ON user_profiles USING GIN(auto_tags);
CREATE INDEX idx_profiles_manual_tags ON user_profiles USING GIN(manual_tags);

-- RLS
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON user_profiles
    USING (tenant_id = current_setting('app.current_tenant', true));

-- Down Migration
DROP TABLE IF EXISTS user_profiles;
