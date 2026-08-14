ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_brand;
DROP INDEX IF EXISTS idx_products_brand_id;
ALTER TABLE products DROP COLUMN IF EXISTS brand_id;
DROP INDEX IF EXISTS idx_product_brands_enabled_order;
DROP INDEX IF EXISTS idx_product_brands_slug;
DROP TABLE IF EXISTS product_brands;
