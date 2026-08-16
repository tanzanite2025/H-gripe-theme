CREATE TABLE IF NOT EXISTS ops_connector_oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    connector_id BIGINT NULL,
    state_hash VARCHAR(128) NOT NULL,
    code_verifier_encrypted TEXT NOT NULL DEFAULT '',
    client_id VARCHAR(255) NOT NULL DEFAULT '',
    redirect_uri VARCHAR(500) NOT NULL DEFAULT '',
    return_path VARCHAR(500) NOT NULL DEFAULT '',
    created_by_user_id BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ops_connector_oauth_session_connector
        FOREIGN KEY (connector_id) REFERENCES ops_connectors (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_connector_oauth_session_state_hash
    ON ops_connector_oauth_sessions (state_hash);
CREATE INDEX IF NOT EXISTS idx_ops_connector_oauth_session_expiry
    ON ops_connector_oauth_sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_ops_connector_oauth_session_connector
    ON ops_connector_oauth_sessions (connector_id);
