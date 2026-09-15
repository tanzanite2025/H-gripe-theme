DROP INDEX IF EXISTS idx_order_items_product_category_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS product_category_id;
ALTER TABLE order_items DROP COLUMN IF EXISTS product_category_slug;
