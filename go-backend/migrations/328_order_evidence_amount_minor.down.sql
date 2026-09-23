ALTER TABLE order_evidence_snapshots
    ADD COLUMN IF NOT EXISTS order_total_amount NUMERIC(14,2),
    ADD COLUMN IF NOT EXISTS order_total_usd NUMERIC(14,2);

UPDATE order_evidence_snapshots
SET order_total_amount = order_total_amount_minor /
        CASE WHEN UPPER(currency) IN ('JPY', 'KRW', 'CLP') THEN 1 ELSE 100 END,
    order_total_usd = order_total_usd_minor / 100.0;

ALTER TABLE order_evidence_snapshots
    DROP CONSTRAINT IF EXISTS order_evidence_snapshot_total_amount_minor_check,
    DROP CONSTRAINT IF EXISTS order_evidence_snapshot_total_usd_minor_check,
    DROP COLUMN IF EXISTS order_total_amount_minor,
    DROP COLUMN IF EXISTS order_total_usd_minor;

ALTER TABLE order_evidence_packages
    ADD COLUMN IF NOT EXISTS order_total_usd_snapshot NUMERIC(14,2);

UPDATE order_evidence_packages
SET order_total_usd_snapshot = order_total_usd_snapshot_minor / 100.0;

ALTER TABLE order_evidence_packages
    DROP CONSTRAINT IF EXISTS order_evidence_package_total_usd_minor_check,
    DROP COLUMN IF EXISTS order_total_usd_snapshot_minor,
    ALTER COLUMN order_total_usd_snapshot SET NOT NULL;
