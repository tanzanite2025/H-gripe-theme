DROP INDEX IF EXISTS idx_order_evidence_snapshots_spoke_tension_qc;
DROP INDEX IF EXISTS idx_order_evidence_snapshots_high_value;
DROP TABLE IF EXISTS order_evidence_snapshots;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS weight_grams;
