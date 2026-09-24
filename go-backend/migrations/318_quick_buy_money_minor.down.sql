ALTER TABLE quick_buy_sessions
    DROP CONSTRAINT IF EXISTS ck_quick_buy_sessions_subtotal_snapshot_minor_nonnegative;

ALTER TABLE quick_buy_session_items
    DROP CONSTRAINT IF EXISTS ck_quick_buy_session_items_unit_price_snapshot_minor_nonnegative;

ALTER TABLE quick_buy_sessions
    ADD COLUMN IF NOT EXISTS subtotal_snapshot NUMERIC(12,2) NOT NULL DEFAULT 0;

UPDATE quick_buy_sessions
SET subtotal_snapshot = subtotal_snapshot_minor / CASE
    WHEN currency IN ('JPY', 'KRW', 'CLP') THEN 1
    ELSE 100
END;

ALTER TABLE quick_buy_sessions
    DROP COLUMN IF EXISTS subtotal_snapshot_minor;

ALTER TABLE quick_buy_session_items
    ADD COLUMN IF NOT EXISTS unit_price_snapshot NUMERIC(12,2) NOT NULL DEFAULT 0;

UPDATE quick_buy_session_items
SET unit_price_snapshot = unit_price_snapshot_minor / CASE
    WHEN currency_snapshot IN ('JPY', 'KRW', 'CLP') THEN 1
    ELSE 100
END;

ALTER TABLE quick_buy_session_items
    DROP COLUMN IF EXISTS unit_price_snapshot_minor;
