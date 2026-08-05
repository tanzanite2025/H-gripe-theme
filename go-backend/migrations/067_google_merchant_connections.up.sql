CREATE TABLE IF NOT EXISTS google_merchant_connections (
  id BIGSERIAL PRIMARY KEY,
  provider VARCHAR(40) NOT NULL UNIQUE,
  google_subject VARCHAR(160) NOT NULL DEFAULT '',
  google_account_email VARCHAR(320) NOT NULL DEFAULT '',
  merchant_account_id VARCHAR(40) NOT NULL DEFAULT '',
  data_source_id VARCHAR(40) NOT NULL DEFAULT '',
  refresh_token_encrypted TEXT NOT NULL DEFAULT '',
  granted_scopes TEXT NOT NULL DEFAULT '',
  token_expires_at TIMESTAMPTZ,
  status VARCHAR(24) NOT NULL DEFAULT 'disconnected',
  oauth_state_hash VARCHAR(128) NOT NULL DEFAULT '',
  oauth_state_expires_at TIMESTAMPTZ,
  oauth_initiated_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  last_connected_at TIMESTAMPTZ,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_google_merchant_connections_status
  ON google_merchant_connections (status);
