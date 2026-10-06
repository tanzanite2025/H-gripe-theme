CREATE TABLE IF NOT EXISTS shipping_yanwen_api_configs (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL DEFAULT 'fat' UNIQUE,
    endpoint VARCHAR(500) NOT NULL,
    user_id_encrypted TEXT NOT NULL DEFAULT '',
    api_token_encrypted TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
