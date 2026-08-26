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
    supplier_address VARCHAR(500),
    supplier_product_code VARCHAR(160),
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    minimum_order_quantity INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_procurement_records_product_code
    ON product_procurement_records(product_code);
CREATE INDEX IF NOT EXISTS idx_product_procurement_records_product_name
    ON product_procurement_records(product_name);
CREATE INDEX IF NOT EXISTS idx_product_procurement_records_supplier
    ON product_procurement_records(supplier_name);
CREATE INDEX IF NOT EXISTS idx_product_procurement_records_enabled
    ON product_procurement_records(is_enabled);
