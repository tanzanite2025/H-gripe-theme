ALTER TABLE ops_domain_bindings
    ADD COLUMN IF NOT EXISTS observed_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS observed_target VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_proxy_mode VARCHAR(32) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS observed_tls_mode VARCHAR(32) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS observed_source VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_observed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS observed_error TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_observed_status
    ON ops_domain_bindings (observed_status);
