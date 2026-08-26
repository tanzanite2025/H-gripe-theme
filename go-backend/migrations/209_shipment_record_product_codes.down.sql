ALTER TABLE shipment_records
    DROP COLUMN IF EXISTS details_bound,
    DROP COLUMN IF EXISTS product_codes;
