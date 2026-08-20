CREATE TABLE IF NOT EXISTS ops_network_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    vps_binding_id BIGINT NULL,
    project_binding_id BIGINT NULL,
    domain_binding_id BIGINT NULL,
    connector_id BIGINT NULL,
    owner_kind VARCHAR(32) NOT NULL DEFAULT 'manual',
    owner_id BIGINT NOT NULL DEFAULT 0,
    managed_by VARCHAR(64) NOT NULL DEFAULT 'manual',
    source_kind VARCHAR(64) NOT NULL DEFAULT '',
    scope VARCHAR(64) NOT NULL,
    direction VARCHAR(16) NOT NULL DEFAULT 'ingress',
    protocol VARCHAR(16) NOT NULL DEFAULT 'tcp',
    ports VARCHAR(120) NOT NULL DEFAULT '',
    source_cidr VARCHAR(120) NOT NULL DEFAULT '',
    target VARCHAR(255) NOT NULL DEFAULT '',
    desired_state VARCHAR(32) NOT NULL DEFAULT 'unknown',
    observed_state VARCHAR(32) NOT NULL DEFAULT 'unknown',
    effective_state VARCHAR(32) NOT NULL DEFAULT 'unknown',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_observed_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_ops_network_rule_vps
        FOREIGN KEY (vps_binding_id)
        REFERENCES ops_vps_bindings(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_ops_network_rule_project
        FOREIGN KEY (project_binding_id)
        REFERENCES ops_project_bindings(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_ops_network_rule_domain
        FOREIGN KEY (domain_binding_id)
        REFERENCES ops_domain_bindings(id)
        ON DELETE SET NULL,
    CONSTRAINT fk_ops_network_rule_connector
        FOREIGN KEY (connector_id)
        REFERENCES ops_connectors(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_ops_network_rule_environment
    ON ops_network_rules(environment);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_owner
    ON ops_network_rules(owner_kind, owner_id);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_manager
    ON ops_network_rules(managed_by);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_scope
    ON ops_network_rules(scope);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_vps
    ON ops_network_rules(vps_binding_id);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_project
    ON ops_network_rules(project_binding_id);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_domain
    ON ops_network_rules(domain_binding_id);
CREATE INDEX IF NOT EXISTS idx_ops_network_rule_connector
    ON ops_network_rules(connector_id);

-- These rows express the documented production boundary. They intentionally
-- remain pending/unknown until an integration observes the live configuration.
INSERT INTO ops_network_rules (
    name,
    environment,
    vps_binding_id,
    connector_id,
    owner_kind,
    owner_id,
    managed_by,
    source_kind,
    scope,
    direction,
    protocol,
    ports,
    source_cidr,
    target,
    desired_state,
    observed_state,
    effective_state,
    status,
    enabled,
    notes
)
SELECT
    'Hostinger production ingress',
    'production',
    vps.id,
    connector.id,
    'vps',
    vps.id,
    'hostinger',
    'firewall_rule',
    'os_firewall',
    'ingress',
    'tcp',
    '80,443',
    'Cloudflare egress ranges',
    COALESCE(NULLIF(vps.ipv4, ''), NULLIF(vps.hostname, ''), vps.name),
    'open',
    'unknown',
    'unknown',
    'pending',
    TRUE,
    '声明式架构基线：Cloudflare 经 Hostinger 防火墙进入共享网关；尚未完成实时防火墙同步。'
FROM ops_vps_bindings AS vps
LEFT JOIN ops_connectors AS connector
    ON connector.name = 'Hostinger Production'
    AND connector.environment = 'production'
    AND connector.deleted_at IS NULL
WHERE vps.name = 'Hostinger Production VPS'
  AND vps.environment = 'production'
  AND vps.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM ops_network_rules AS rule
      WHERE rule.name = 'Hostinger production ingress'
        AND rule.environment = 'production'
        AND rule.deleted_at IS NULL
  );

INSERT INTO ops_network_rules (
    name,
    environment,
    vps_binding_id,
    project_binding_id,
    connector_id,
    owner_kind,
    owner_id,
    managed_by,
    source_kind,
    scope,
    direction,
    protocol,
    ports,
    target,
    desired_state,
    observed_state,
    effective_state,
    status,
    enabled,
    notes
)
SELECT
    'shared-edge to theme-web',
    'production',
    vps.id,
    project.id,
    COALESCE(project.connector_id, vps.connector_id),
    'project',
    project.id,
    'manual',
    'manual',
    'gateway',
    'ingress',
    'tcp',
    '8080',
    'theme-web:8080',
    'open',
    'unknown',
    'unknown',
    'pending',
    TRUE,
    '声明式架构基线：共享 Caddy 网关将已登记站点路由到 theme-web:8080；尚未完成实时网关检查。'
FROM ops_project_bindings AS project
JOIN ops_vps_bindings AS vps
    ON vps.id = project.vps_binding_id
    AND vps.deleted_at IS NULL
WHERE project.name = 'commerce-platform'
  AND project.environment = 'production'
  AND project.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM ops_network_rules AS rule
      WHERE rule.name = 'shared-edge to theme-web'
        AND rule.environment = 'production'
        AND rule.deleted_at IS NULL
  );
