CREATE TABLE IF NOT EXISTS ops_domain_bindings (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'alias',
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    provider VARCHAR(32) NOT NULL DEFAULT 'cloudflare',
    zone VARCHAR(255) NOT NULL DEFAULT '',
    target VARCHAR(255) NOT NULL DEFAULT '',
    proxy_mode VARCHAR(32) NOT NULL DEFAULT 'unknown',
    tls_mode VARCHAR(32) NOT NULL DEFAULT 'unknown',
    redirect_target VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_domain_binding_domain
    ON ops_domain_bindings (domain);
CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_role
    ON ops_domain_bindings (role);
CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_environment
    ON ops_domain_bindings (environment);
CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_status
    ON ops_domain_bindings (status);

INSERT INTO ops_domain_bindings (
    domain,
    role,
    environment,
    provider,
    zone,
    target,
    proxy_mode,
    tls_mode,
    status,
    enabled,
    notes
)
VALUES
    (
        'learn.gripe',
        'canonical',
        'production',
        'cloudflare',
        'learn.gripe',
        'theme-web:8080',
        'proxied',
        'full_strict',
        'active',
        TRUE,
        '当前生产主域名；来源于现有 Cloudflare/Caddy 发布边界。'
    ),
    (
        'www.learn.gripe',
        'alias',
        'production',
        'cloudflare',
        'learn.gripe',
        'theme-web:8080',
        'proxied',
        'full_strict',
        'active',
        TRUE,
        '当前生产 www 别名。'
    ),
    (
        'admin.learn.gripe',
        'admin',
        'production',
        'cloudflare',
        'learn.gripe',
        'theme-web:8080',
        'proxied',
        'full_strict',
        'active',
        TRUE,
        '当前后台入口域名；目前与前台共用 web 网关。'
    )
ON CONFLICT (domain) DO NOTHING;
