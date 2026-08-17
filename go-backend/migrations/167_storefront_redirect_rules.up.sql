CREATE TABLE IF NOT EXISTS storefront_redirect_rules (
    id BIGSERIAL PRIMARY KEY,
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 301,
    state VARCHAR(32) NOT NULL DEFAULT 'draft',
    reason TEXT NOT NULL DEFAULT '',
    created_by_id BIGINT NOT NULL DEFAULT 0,
    published_by_id BIGINT,
    published_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_storefront_redirect_rules_status_code
        CHECK (status_code IN (301, 308)),
    CONSTRAINT chk_storefront_redirect_rules_state
        CHECK (state IN ('draft', 'published', 'disabled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storefront_redirect_rules_source_path
    ON storefront_redirect_rules (source_path);
CREATE INDEX IF NOT EXISTS idx_storefront_redirect_rules_state
    ON storefront_redirect_rules (state);
CREATE INDEX IF NOT EXISTS idx_storefront_redirect_rules_published_at
    ON storefront_redirect_rules (published_at DESC);
