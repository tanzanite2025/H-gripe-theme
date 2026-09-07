ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS weight_grams INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS order_evidence_snapshots (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE RESTRICT,
    schema_version INTEGER NOT NULL,
    confirmed_at TIMESTAMPTZ NOT NULL,
    currency VARCHAR(3) NOT NULL,
    order_total_amount NUMERIC(14,2) NOT NULL,
    order_total_usd NUMERIC(14,2) NOT NULL,
    is_high_value BOOLEAN NOT NULL DEFAULT FALSE,
    has_spoke_tension_qc BOOLEAN NOT NULL DEFAULT FALSE,
    snapshot_data JSONB NOT NULL,
    snapshot_sha256 CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_evidence_snapshot_schema_version_check
        CHECK (schema_version = 1),
    CONSTRAINT order_evidence_snapshot_currency_check
        CHECK (char_length(currency) = 3),
    CONSTRAINT order_evidence_snapshot_total_amount_check
        CHECK (order_total_amount >= 0),
    CONSTRAINT order_evidence_snapshot_total_usd_check
        CHECK (order_total_usd >= 0)
);

CREATE INDEX IF NOT EXISTS idx_order_evidence_snapshots_high_value
    ON order_evidence_snapshots(is_high_value);
CREATE INDEX IF NOT EXISTS idx_order_evidence_snapshots_spoke_tension_qc
    ON order_evidence_snapshots(has_spoke_tension_qc);
