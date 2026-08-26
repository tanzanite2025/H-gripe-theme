CREATE TABLE IF NOT EXISTS fitment_fork_entries (
    id BIGSERIAL PRIMARY KEY,
    brand_name VARCHAR(160) NOT NULL,
    model_name VARCHAR(160) NOT NULL,
    series_name VARCHAR(160),
    generation_name VARCHAR(160),
    year_mode VARCHAR(16) NOT NULL DEFAULT 'unknown',
    year_from INTEGER,
    year_to INTEGER,
    market_code VARCHAR(32),
    notes TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fitment_fork_entries_year_mode_check CHECK (
        (year_mode = 'single' AND year_from IS NOT NULL AND year_to IS NULL)
        OR (year_mode = 'range' AND year_from IS NOT NULL AND year_to IS NOT NULL AND year_from <= year_to)
        OR (year_mode IN ('all', 'unknown') AND year_from IS NULL AND year_to IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_fitment_fork_entries_brand
    ON fitment_fork_entries(brand_name);
CREATE INDEX IF NOT EXISTS idx_fitment_fork_entries_model
    ON fitment_fork_entries(model_name);
CREATE INDEX IF NOT EXISTS idx_fitment_fork_entries_year
    ON fitment_fork_entries(year_from, year_to);
CREATE INDEX IF NOT EXISTS idx_fitment_fork_entries_enabled
    ON fitment_fork_entries(is_enabled);
CREATE INDEX IF NOT EXISTS idx_fitment_fork_entries_deleted
    ON fitment_fork_entries(deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS uk_fitment_fork_entries_identity
    ON fitment_fork_entries (
        LOWER(BTRIM(brand_name)),
        LOWER(BTRIM(model_name)),
        LOWER(BTRIM(COALESCE(series_name, ''))),
        LOWER(BTRIM(COALESCE(generation_name, ''))),
        year_mode,
        COALESCE(year_from, 0),
        COALESCE(year_to, 0),
        LOWER(BTRIM(COALESCE(market_code, '')))
    )
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS fitment_fork_hub_specifications (
    fork_entry_id BIGINT NOT NULL
        REFERENCES fitment_fork_entries(id) ON DELETE CASCADE,
    hub_specification_id BIGINT NOT NULL
        REFERENCES fitment_hub_specifications(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (fork_entry_id, hub_specification_id)
);

CREATE INDEX IF NOT EXISTS idx_fitment_fork_hub_specifications_hub
    ON fitment_fork_hub_specifications(hub_specification_id);
