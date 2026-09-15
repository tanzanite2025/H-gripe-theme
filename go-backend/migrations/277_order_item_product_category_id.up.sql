ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS product_category_id BIGINT;
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS product_category_slug VARCHAR(120);

CREATE INDEX IF NOT EXISTS idx_order_items_product_category_id
    ON order_items(product_category_id);
