CREATE TABLE IF NOT EXISTS storefront_url_search_profiles (
    id BIGSERIAL PRIMARY KEY,
    route_entry_id BIGINT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    search_weight INTEGER NOT NULL DEFAULT 100,
    keywords_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    display_title TEXT NOT NULL DEFAULT '',
    display_summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_storefront_url_search_profiles_route_entry_id UNIQUE (route_entry_id),
    CONSTRAINT fk_storefront_url_search_profiles_route_entry
        FOREIGN KEY (route_entry_id)
        REFERENCES storefront_route_catalog_entries (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_storefront_url_search_profiles_enabled
    ON storefront_url_search_profiles (enabled);
CREATE INDEX IF NOT EXISTS idx_storefront_url_search_profiles_search_weight
    ON storefront_url_search_profiles (search_weight);
