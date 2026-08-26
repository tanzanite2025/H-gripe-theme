ALTER TABLE product_procurement_records
    ADD COLUMN IF NOT EXISTS supplier_address VARCHAR(500),
    ADD COLUMN IF NOT EXISTS supplier_product_code VARCHAR(160),
    ADD COLUMN IF NOT EXISTS notes TEXT,
    ADD COLUMN IF NOT EXISTS is_enabled BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_product_procurement_records_enabled
    ON product_procurement_records(is_enabled);
