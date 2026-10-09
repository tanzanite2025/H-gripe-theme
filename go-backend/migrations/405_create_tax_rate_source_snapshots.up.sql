-- Keep periodically collected external VAT reference data versioned and
-- separate from the tax-rate records consumed by checkout.
CREATE TABLE IF NOT EXISTS tax_rate_source_configs (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    provider_code VARCHAR(40) NOT NULL DEFAULT 'vatcomply' CHECK (provider_code = 'vatcomply'),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    refresh_interval_hours INTEGER NOT NULL DEFAULT 720
        CHECK (refresh_interval_hours BETWEEN 24 AND 8760),
    last_checked_at TIMESTAMPTZ,
    last_successful_sync_at TIMESTAMPTZ,
    last_sync_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tax_rate_source_configs (id, provider_code, enabled, refresh_interval_hours)
VALUES (1, 'vatcomply', FALSE, 720)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS tax_rate_source_snapshots (
    id BIGSERIAL PRIMARY KEY,
    provider_code VARCHAR(40) NOT NULL,
    source_endpoint VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL UNIQUE,
    content_sha256 CHAR(64) NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL,
    country_count INTEGER NOT NULL CHECK (country_count > 0),
    rate_count INTEGER NOT NULL CHECK (rate_count > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tax_rate_source_snapshots_provider_created
    ON tax_rate_source_snapshots (provider_code, captured_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS tax_rate_source_snapshot_entries (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL REFERENCES tax_rate_source_snapshots(id) ON DELETE CASCADE,
    country_code CHAR(2) NOT NULL,
    source_country_code CHAR(2) NOT NULL,
    country_name VARCHAR(120) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT '',
    rate_type VARCHAR(32) NOT NULL,
    rate_category VARCHAR(96) NOT NULL DEFAULT '',
    rate_decimal NUMERIC(12, 8) NOT NULL CHECK (rate_decimal BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tax_rate_source_snapshot_entries_fact
        UNIQUE (snapshot_id, country_code, rate_type, rate_category, rate_decimal)
);

CREATE INDEX IF NOT EXISTS idx_tax_rate_source_snapshot_entries_snapshot_country
    ON tax_rate_source_snapshot_entries (snapshot_id, country_code, rate_type);
