ALTER TABLE product_procurement_records
    DROP COLUMN IF EXISTS inbound_shipping_unit_cost,
    DROP COLUMN IF EXISTS customs_unit_cost,
    DROP COLUMN IF EXISTS packaging_unit_cost,
    DROP COLUMN IF EXISTS other_unit_cost;
