-- Recreate the procurement table when a database has an advanced migration
-- ledger but the table was removed or never created.
CREATE TABLE IF NOT EXISTS product_procurement_records (
    id BIGSERIAL PRIMARY KEY,
    product_code VARCHAR(160) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    purchase_price NUMERIC(14, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    supplier_name VARCHAR(255) NOT NULL,
    supplier_contact_name VARCHAR(255),
    supplier_phone VARCHAR(80),
    supplier_email VARCHAR(190),
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    minimum_order_quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inbound_shipping_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    packaging_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    other_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0
);

ALTER TABLE product_procurement_records
    ADD COLUMN IF NOT EXISTS inbound_shipping_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS packaging_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS other_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_procurement_records_product_code
    ON product_procurement_records(product_code);
CREATE INDEX IF NOT EXISTS idx_product_procurement_records_product_name
    ON product_procurement_records(product_name);
CREATE INDEX IF NOT EXISTS idx_product_procurement_records_supplier
    ON product_procurement_records(supplier_name);
