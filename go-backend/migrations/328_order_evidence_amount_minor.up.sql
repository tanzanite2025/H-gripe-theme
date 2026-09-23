-- Order evidence is an immutable financial snapshot. Persist exact minor
-- units; do not retain a second floating-point representation.
ALTER TABLE order_evidence_snapshots
    ADD COLUMN IF NOT EXISTS order_total_amount_minor BIGINT,
    ADD COLUMN IF NOT EXISTS order_total_usd_minor BIGINT;

UPDATE order_evidence_snapshots
SET order_total_amount_minor = ROUND(
        order_total_amount * CASE
            WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN 1
            ELSE 100
        END
    )::BIGINT,
    order_total_usd_minor = ROUND(order_total_usd * 100)::BIGINT
WHERE order_total_amount_minor IS NULL OR order_total_usd_minor IS NULL;

ALTER TABLE order_evidence_snapshots
    ALTER COLUMN order_total_amount_minor SET NOT NULL,
    ALTER COLUMN order_total_amount_minor SET DEFAULT 0,
    ALTER COLUMN order_total_usd_minor SET NOT NULL,
    ALTER COLUMN order_total_usd_minor SET DEFAULT 0,
    ADD CONSTRAINT order_evidence_snapshot_total_amount_minor_check
        CHECK (order_total_amount_minor >= 0),
    ADD CONSTRAINT order_evidence_snapshot_total_usd_minor_check
        CHECK (order_total_usd_minor >= 0),
    DROP COLUMN IF EXISTS order_total_amount,
    DROP COLUMN IF EXISTS order_total_usd;

ALTER TABLE order_evidence_packages
    ADD COLUMN IF NOT EXISTS order_total_usd_snapshot_minor BIGINT;

UPDATE order_evidence_packages
SET order_total_usd_snapshot_minor = ROUND(order_total_usd_snapshot * 100)::BIGINT
WHERE order_total_usd_snapshot_minor IS NULL;

ALTER TABLE order_evidence_packages
    ALTER COLUMN order_total_usd_snapshot_minor SET NOT NULL,
    ALTER COLUMN order_total_usd_snapshot_minor SET DEFAULT 0,
    ADD CONSTRAINT order_evidence_package_total_usd_minor_check
        CHECK (order_total_usd_snapshot_minor >= 0),
    DROP COLUMN IF EXISTS order_total_usd_snapshot;
