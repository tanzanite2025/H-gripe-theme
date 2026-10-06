ALTER TABLE shipping_yanwen_tracking_snapshots
    ADD COLUMN IF NOT EXISTS has_official_result BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS polling_state VARCHAR(16) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS next_sync_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_polling_attempt_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS polling_lease_token VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS polling_lease_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS polling_failure_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_polling_error VARCHAR(1000) NOT NULL DEFAULT '';

ALTER TABLE shipping_yanwen_tracking_snapshots
    ALTER COLUMN tracking_status SET DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_yanwen_tracking_snapshots_polling_due
    ON shipping_yanwen_tracking_snapshots (polling_state, next_sync_at, polling_lease_until);

-- Existing rows were created from a validated official response. Keep those
-- facts active; newly backfilled rows below are only internal poll targets.
UPDATE shipping_yanwen_tracking_snapshots
SET has_official_result = TRUE,
    polling_state = CASE
        WHEN UPPER(TRIM(tracking_status)) IN ('LM40', 'LM90') THEN 'terminal'
        ELSE 'active'
    END,
    next_sync_at = CASE
        WHEN UPPER(TRIM(tracking_status)) IN ('LM40', 'LM90') THEN NULL
        ELSE COALESCE(next_sync_at, NOW())
    END
WHERE tracking_status <> '';

-- Production waybills created before polling was introduced still need a
-- target. These rows contain no official status or checkpoint and therefore
-- cannot be displayed as a tracking result until the real endpoint answers.
INSERT INTO shipping_yanwen_tracking_snapshots (
    environment,
    tracking_number,
    waybill_number,
    tracking_status,
    checkpoints_data,
    response_data,
    has_official_result,
    polling_state,
    next_sync_at,
    last_synced_at,
    created_at,
    updated_at
)
SELECT
    waybill.environment,
    waybill.waybill_number,
    waybill.waybill_number,
    '',
    '[]'::jsonb,
    '{}'::jsonb,
    FALSE,
    'pending',
    NOW(),
    TIMESTAMPTZ '0001-01-01 00:00:00+00',
    NOW(),
    NOW()
FROM shipping_yanwen_waybills AS waybill
WHERE waybill.environment = 'production'
ON CONFLICT (environment, tracking_number) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_yanwen_tracking_snapshots_has_official_result
    ON shipping_yanwen_tracking_snapshots (has_official_result);
