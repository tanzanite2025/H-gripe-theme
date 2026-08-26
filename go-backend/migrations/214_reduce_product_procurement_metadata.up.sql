ALTER TABLE product_procurement_records
    DROP COLUMN IF EXISTS supplier_address,
    DROP COLUMN IF EXISTS supplier_product_code,
    DROP COLUMN IF EXISTS notes,
    DROP COLUMN IF EXISTS is_enabled;
