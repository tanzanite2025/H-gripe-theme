DROP INDEX IF EXISTS idx_products_product_category_id;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS fk_products_product_category;

ALTER TABLE products
    DROP COLUMN IF EXISTS product_category_id;
