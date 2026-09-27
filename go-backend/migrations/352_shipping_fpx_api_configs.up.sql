CREATE TABLE IF NOT EXISTS shipping_fpx_api_configs (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL DEFAULT 'production' UNIQUE,
    endpoint VARCHAR(500) NOT NULL,
    app_key_encrypted TEXT NOT NULL DEFAULT '',
    app_secret_encrypted TEXT NOT NULL DEFAULT '',
    access_token_encrypted TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_sync_status VARCHAR(32) NOT NULL DEFAULT '',
    last_synced_at TIMESTAMPTZ,
    last_error VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
