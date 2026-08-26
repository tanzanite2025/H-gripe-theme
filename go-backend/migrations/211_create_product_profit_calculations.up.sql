CREATE TABLE IF NOT EXISTS product_profit_calculations (
    id BIGSERIAL PRIMARY KEY,
    product_code VARCHAR(160) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    list_price NUMERIC(14, 2) NOT NULL,
    sale_price NUMERIC(14, 2),
    effective_selling_price NUMERIC(14, 2) NOT NULL,
    purchase_price NUMERIC(14, 2) NOT NULL,
    inbound_shipping_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    customs_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    packaging_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    other_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0,
    landed_cost NUMERIC(14, 2) NOT NULL,
    gross_profit NUMERIC(14, 2) NOT NULL,
    gross_margin_bps INTEGER NOT NULL,
    calculation_status VARCHAR(40) NOT NULL DEFAULT 'ready',
    formula_version VARCHAR(32) NOT NULL,
    warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_profit_calculations_product_code
    ON product_profit_calculations(product_code);
CREATE INDEX IF NOT EXISTS idx_product_profit_calculations_updated_at
    ON product_profit_calculations(updated_at DESC);
