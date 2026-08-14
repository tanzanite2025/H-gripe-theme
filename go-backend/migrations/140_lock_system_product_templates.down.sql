DROP INDEX IF EXISTS idx_product_types_system_managed;

ALTER TABLE product_types
    DROP COLUMN IF EXISTS is_system_managed;
