ALTER TABLE shipping_yanwen_waybills
    ADD COLUMN IF NOT EXISTS reference_number VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS official_status INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_printed BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS last_official_synced_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_yanwen_waybills_official_status
    ON shipping_yanwen_waybills (official_status);
