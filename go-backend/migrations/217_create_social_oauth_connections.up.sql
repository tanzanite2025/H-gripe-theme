CREATE TABLE IF NOT EXISTS social_oauth_connections (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'disconnected',
    provider_account_id VARCHAR(255) NOT NULL DEFAULT '',
    provider_account_name VARCHAR(255) NOT NULL DEFAULT '',
    provider_account_url VARCHAR(500) NOT NULL DEFAULT '',
    provider_account_email VARCHAR(320) NOT NULL DEFAULT '',
    access_token_encrypted TEXT NOT NULL DEFAULT '',
    refresh_token_encrypted TEXT NOT NULL DEFAULT '',
    token_expires_at TIMESTAMPTZ NULL,
    granted_scopes TEXT NOT NULL DEFAULT '',
    provider_resources JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_connected_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_social_oauth_connection_provider UNIQUE (provider)
);

CREATE INDEX IF NOT EXISTS idx_social_oauth_connection_status
    ON social_oauth_connections (status);

CREATE TABLE IF NOT EXISTS social_oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    state_hash VARCHAR(128) NOT NULL,
    code_verifier_encrypted TEXT NOT NULL DEFAULT '',
    redirect_uri VARCHAR(500) NOT NULL DEFAULT '',
    return_path VARCHAR(500) NOT NULL DEFAULT '',
    initiated_by_user_id BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_social_oauth_session_state_hash UNIQUE (state_hash)
);

CREATE INDEX IF NOT EXISTS idx_social_oauth_session_expiry
    ON social_oauth_sessions (expires_at);

CREATE INDEX IF NOT EXISTS idx_social_oauth_session_provider
    ON social_oauth_sessions (provider);

CREATE INDEX IF NOT EXISTS idx_social_oauth_session_user
    ON social_oauth_sessions (initiated_by_user_id);
