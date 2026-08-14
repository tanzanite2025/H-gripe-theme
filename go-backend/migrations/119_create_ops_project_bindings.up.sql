CREATE TABLE IF NOT EXISTS ops_project_bindings (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    vps_binding_id BIGINT NOT NULL,
    connector_id BIGINT NULL,
    provider_resource_id VARCHAR(120) NOT NULL DEFAULT '',
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    compose_source VARCHAR(255) NOT NULL DEFAULT '',
    compose_project_name VARCHAR(120) NOT NULL DEFAULT '',
    gateway_network VARCHAR(120) NOT NULL DEFAULT '',
    gateway_alias VARCHAR(120) NOT NULL DEFAULT '',
    services TEXT NOT NULL DEFAULT '',
    networks TEXT NOT NULL DEFAULT '',
    volumes TEXT NOT NULL DEFAULT '',
    current_image_tag VARCHAR(160) NOT NULL DEFAULT '',
    current_commit_sha VARCHAR(80) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_deployment_at TIMESTAMPTZ NULL,
    last_checked_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    backup_policy TEXT NOT NULL DEFAULT '',
    restore_notes TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_ops_project_binding_vps
        FOREIGN KEY (vps_binding_id) REFERENCES ops_vps_bindings (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_project_binding_name
    ON ops_project_bindings (name);
CREATE INDEX IF NOT EXISTS idx_ops_project_binding_vps
    ON ops_project_bindings (vps_binding_id);
CREATE INDEX IF NOT EXISTS idx_ops_project_binding_environment
    ON ops_project_bindings (environment);
CREATE INDEX IF NOT EXISTS idx_ops_project_binding_status
    ON ops_project_bindings (status);

INSERT INTO ops_project_bindings (
    name,
    vps_binding_id,
    provider_resource_id,
    environment,
    compose_source,
    compose_project_name,
    gateway_network,
    gateway_alias,
    services,
    networks,
    volumes,
    current_image_tag,
    status,
    health_status,
    enabled,
    backup_policy,
    restore_notes,
    notes
)
SELECT
    'commerce-platform',
    vps.id,
    '',
    'production',
    'compose.prod.yml',
    'commerce-platform',
    'shared-edge',
    'theme-web',
    'db, redis, api, storefront, admin, web',
    'db, cache, app, shared-edge',
    'commerce-platform-postgres-data, commerce-platform-redis-data, commerce-platform-uploads',
    'master',
    'active',
    'unknown',
    TRUE,
    '每日 PostgreSQL 逻辑备份；每日 uploads 备份；备份异地保存；每月恢复演练。',
    'Hostinger snapshot 只作为灾备辅助，不能替代数据库和 uploads 备份。恢复演练结果待补录。',
    '当前生产项目初始基线来自 docs/ops/hostinger-vps-docker-runbook.md；未登记项目级 Hostinger ID、实时服务健康和最后部署时间。'
FROM ops_vps_bindings vps
WHERE vps.name = 'Hostinger Production VPS'
ON CONFLICT (name) DO NOTHING;
