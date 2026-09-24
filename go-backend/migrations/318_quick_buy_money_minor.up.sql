-- QUICK sessions store transactional snapshots in currency minor units.
-- Existing rows are converted once at migration time; this is safe for the
-- pre-launch schema and avoids carrying a floating-point pricing path.
ALTER TABLE quick_buy_sessions
    ADD COLUMN IF NOT EXISTS subtotal_snapshot_minor BIGINT NOT NULL DEFAULT 0;

UPDATE quick_buy_sessions
SET subtotal_snapshot_minor = ROUND(
    subtotal_snapshot * CASE
        WHEN currency IN ('JPY', 'KRW', 'CLP') THEN 1
        ELSE 100
    END
)
WHERE subtotal_snapshot_minor = 0 AND subtotal_snapshot <> 0;

ALTER TABLE quick_buy_sessions
    DROP COLUMN IF EXISTS subtotal_snapshot;

ALTER TABLE quick_buy_session_items
    ADD COLUMN IF NOT EXISTS unit_price_snapshot_minor BIGINT NOT NULL DEFAULT 0;

UPDATE quick_buy_session_items
SET unit_price_snapshot_minor = ROUND(
    unit_price_snapshot * CASE
        WHEN currency_snapshot IN ('JPY', 'KRW', 'CLP') THEN 1
        ELSE 100
    END
)
WHERE unit_price_snapshot_minor = 0 AND unit_price_snapshot <> 0;

ALTER TABLE quick_buy_session_items
    DROP COLUMN IF EXISTS unit_price_snapshot;

ALTER TABLE quick_buy_sessions
    ADD CONSTRAINT ck_quick_buy_sessions_subtotal_snapshot_minor_nonnegative
    CHECK (subtotal_snapshot_minor >= 0);

ALTER TABLE quick_buy_session_items
    ADD CONSTRAINT ck_quick_buy_session_items_unit_price_snapshot_minor_nonnegative
    CHECK (unit_price_snapshot_minor >= 0);
