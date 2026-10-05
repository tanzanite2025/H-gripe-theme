DROP INDEX IF EXISTS idx_shipping_yanwen_waybills_official_status;

ALTER TABLE shipping_yanwen_waybills
    DROP COLUMN IF EXISTS last_official_synced_at,
    DROP COLUMN IF EXISTS is_printed,
    DROP COLUMN IF EXISTS official_status,
    DROP COLUMN IF EXISTS reference_number;
