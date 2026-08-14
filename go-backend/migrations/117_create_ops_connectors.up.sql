CREATE TABLE IF NOT EXISTS ops_connectors (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    endpoint VARCHAR(500) NOT NULL DEFAULT '',
    auth_type VARCHAR(32) NOT NULL DEFAULT 'api_token',
    credential_ref VARCHAR(160) NOT NULL DEFAULT '',
    credentials_encrypted TEXT NOT NULL DEFAULT '',
    credential_fields TEXT NOT NULL DEFAULT '',
    scopes TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_test_status VARCHAR(32) NOT NULL DEFAULT '',
    last_tested_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_connector_name
    ON ops_connectors (name);
CREATE INDEX IF NOT EXISTS idx_ops_connector_provider
    ON ops_connectors (provider);
CREATE INDEX IF NOT EXISTS idx_ops_connector_environment
    ON ops_connectors (environment);
CREATE INDEX IF NOT EXISTS idx_ops_connector_status
    ON ops_connectors (status);

INSERT INTO ops_connectors (
    name,
    provider,
    environment,
    endpoint,
    auth_type,
    scopes,
    status,
    enabled,
    notes
)
VALUES
    (
        'Cloudflare Production',
        'cloudflare',
        'production',
        'https://api.cloudflare.com/client/v4/user/tokens/verify',
        'api_token',
        'zones:read,dns:read',
        'pending',
        TRUE,
        '只登记 Cloudflare 只读连接；首次连接测试前请填写受限 API Token。'
    ),
    (
        'Hostinger Production',
        'hostinger',
        'production',
        '',
        'api_token',
        'vps:read,project:read',
        'pending',
        TRUE,
        'Hostinger 连接先登记凭据和权限范围，确认官方只读接口后再填写。'
    ),
    (
        'GitHub GHCR Production',
        'ghcr',
        'production',
        'https://api.github.com/user',
        'api_token',
        'repo:read,packages:read',
        'pending',
        TRUE,
        '用于发布镜像和 GHCR 只读检查；连接测试不执行发布。'
    )
ON CONFLICT (name) DO NOTHING;
