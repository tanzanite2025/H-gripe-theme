ALTER TABLE product_procurement_records
    ADD COLUMN IF NOT EXISTS customs_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0;

ALTER TABLE product_profit_calculations
    ADD COLUMN IF NOT EXISTS customs_unit_cost NUMERIC(14, 2) NOT NULL DEFAULT 0;
