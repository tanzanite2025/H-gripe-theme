ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS fk_product_variants_master_variant;
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS ck_product_variants_master_stock_zero;
DROP INDEX IF EXISTS idx_product_variants_master_variant_id;
ALTER TABLE product_variants DROP COLUMN IF EXISTS master_variant_id;
