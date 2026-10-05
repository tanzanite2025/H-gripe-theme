CREATE TABLE IF NOT EXISTS shipping_yanwen_warehouse_catalog_entries (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL,
    warehouse_code VARCHAR(80) NOT NULL,
    name VARCHAR(200) NOT NULL,
    area VARCHAR(200) NOT NULL DEFAULT '',
    last_synced_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_yanwen_warehouse_catalog_environment_code UNIQUE (environment, warehouse_code)
);

CREATE INDEX IF NOT EXISTS idx_yanwen_warehouse_catalog_environment
    ON shipping_yanwen_warehouse_catalog_entries (environment);
