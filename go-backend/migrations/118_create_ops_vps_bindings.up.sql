CREATE TABLE IF NOT EXISTS ops_vps_bindings (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    provider VARCHAR(32) NOT NULL DEFAULT 'hostinger',
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    connector_id BIGINT NULL,
    provider_resource_id VARCHAR(120) NOT NULL DEFAULT '',
    hostname VARCHAR(255) NOT NULL DEFAULT '',
    ipv4 VARCHAR(64) NOT NULL DEFAULT '',
    region VARCHAR(120) NOT NULL DEFAULT '',
    operating_system VARCHAR(160) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    observed_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_observed_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_vps_binding_name
    ON ops_vps_bindings (name);
CREATE INDEX IF NOT EXISTS idx_ops_vps_binding_provider
    ON ops_vps_bindings (provider);
CREATE INDEX IF NOT EXISTS idx_ops_vps_binding_environment
    ON ops_vps_bindings (environment);
CREATE INDEX IF NOT EXISTS idx_ops_vps_binding_status
    ON ops_vps_bindings (status);

INSERT INTO ops_vps_bindings (
    name,
    provider,
    environment,
    provider_resource_id,
    hostname,
    ipv4,
    region,
    operating_system,
    status,
    observed_status,
    enabled,
    notes
)
VALUES (
    'Hostinger Production VPS',
    'hostinger',
    'production',
    '1834903',
    'srv1834903.hstgr.cloud',
    '2.25.85.201',
    '',
    'Ubuntu 24.04 LTS',
    'active',
    'unknown',
    TRUE,
    '初始基线来自 docs/ops/hostinger-vps-docker-runbook.md；当前只登记资源，不代表已完成 Hostinger 实时同步。'
)
ON CONFLICT (name) DO NOTHING;
