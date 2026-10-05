CREATE TABLE IF NOT EXISTS shipping_yanwen_tracking_snapshots (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL,
    tracking_number VARCHAR(120) NOT NULL,
    waybill_number VARCHAR(120) NOT NULL,
    exchange_number VARCHAR(120) NOT NULL DEFAULT '',
    last_mile_carrier VARCHAR(160) NOT NULL DEFAULT '',
    last_mile_carrier_website VARCHAR(500) NOT NULL DEFAULT '',
    last_mile_carrier_contact_number VARCHAR(160) NOT NULL DEFAULT '',
    tracking_status VARCHAR(80) NOT NULL,
    tracking_status_level1 VARCHAR(40) NOT NULL DEFAULT '',
    tracking_status_level2 VARCHAR(40) NOT NULL DEFAULT '',
    tracking_status_level3 VARCHAR(80) NOT NULL DEFAULT '',
    last_mile_tracking_expected BOOLEAN NOT NULL DEFAULT FALSE,
    origin_country VARCHAR(16) NOT NULL DEFAULT '',
    destination_country VARCHAR(16) NOT NULL DEFAULT '',
    latest_checkpoint_status VARCHAR(80) NOT NULL DEFAULT '',
    latest_checkpoint_time_stamp VARCHAR(80) NOT NULL DEFAULT '',
    checkpoints_data JSONB NOT NULL DEFAULT '[]'::jsonb,
    response_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_yanwen_tracking_snapshot_environment CHECK (environment = 'production'),
    CONSTRAINT uq_yanwen_tracking_snapshot_environment_number UNIQUE (environment, tracking_number)
);

CREATE INDEX IF NOT EXISTS idx_yanwen_tracking_snapshots_status
    ON shipping_yanwen_tracking_snapshots (tracking_status);

CREATE INDEX IF NOT EXISTS idx_yanwen_tracking_snapshots_last_synced_at
    ON shipping_yanwen_tracking_snapshots (last_synced_at);
