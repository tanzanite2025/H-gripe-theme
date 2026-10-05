CREATE TABLE IF NOT EXISTS shipping_yanwen_country_catalog_entries (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL,
    country_id VARCHAR(80) NOT NULL,
    country_code VARCHAR(16) NOT NULL,
    name_ch VARCHAR(200) NOT NULL DEFAULT '',
    name_en VARCHAR(200) NOT NULL DEFAULT '',
    last_synced_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_yanwen_country_catalog_environment_country UNIQUE (environment, country_id)
);

CREATE INDEX IF NOT EXISTS idx_yanwen_country_catalog_environment
    ON shipping_yanwen_country_catalog_entries (environment);
