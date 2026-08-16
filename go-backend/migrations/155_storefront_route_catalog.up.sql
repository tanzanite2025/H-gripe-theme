CREATE TABLE IF NOT EXISTS storefront_route_catalog_entries (
    id BIGSERIAL PRIMARY KEY,
    route_key VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    locale VARCHAR(20) NOT NULL DEFAULT 'en',
    source_type VARCHAR(32) NOT NULL,
    source_id BIGINT,
    source_key VARCHAR(255),
    title TEXT,
    summary TEXT,
    canonical_path TEXT,
    is_alias BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT TRUE,
    is_checkable BOOLEAN NOT NULL DEFAULT TRUE,
    is_indexable BOOLEAN NOT NULL DEFAULT TRUE,
    entry_status VARCHAR(32) NOT NULL DEFAULT 'active',
    duplicate_group_key VARCHAR(255),
    manifest_version VARCHAR(64),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_check_status VARCHAR(32),
    last_http_status INTEGER NOT NULL DEFAULT 0,
    last_final_url TEXT,
    last_canonical_url TEXT,
    last_response_ms INTEGER NOT NULL DEFAULT 0,
    last_redirect_count INTEGER NOT NULL DEFAULT 0,
    last_content_hash VARCHAR(64),
    last_check_error TEXT,
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_route_key
    ON storefront_route_catalog_entries (route_key);
CREATE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_path
    ON storefront_route_catalog_entries (path);
CREATE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_locale
    ON storefront_route_catalog_entries (locale);
CREATE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_source_type
    ON storefront_route_catalog_entries (source_type);
CREATE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_entry_status
    ON storefront_route_catalog_entries (entry_status);
CREATE INDEX IF NOT EXISTS idx_storefront_route_catalog_entries_last_check_status
    ON storefront_route_catalog_entries (last_check_status);

CREATE TABLE IF NOT EXISTS storefront_route_check_results (
    id BIGSERIAL PRIMARY KEY,
    route_entry_id BIGINT NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    http_status INTEGER NOT NULL DEFAULT 0,
    final_url TEXT,
    canonical_url TEXT,
    response_ms INTEGER NOT NULL DEFAULT 0,
    redirect_count INTEGER NOT NULL DEFAULT 0,
    content_hash VARCHAR(64),
    status VARCHAR(32) NOT NULL,
    error_message TEXT,
    CONSTRAINT fk_storefront_route_check_results_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_storefront_route_check_results_entry_id
    ON storefront_route_check_results (route_entry_id);
CREATE INDEX IF NOT EXISTS idx_storefront_route_check_results_checked_at
    ON storefront_route_check_results (checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_storefront_route_check_results_status
    ON storefront_route_check_results (status);
CREATE INDEX IF NOT EXISTS idx_storefront_route_check_results_content_hash
    ON storefront_route_check_results (content_hash);
