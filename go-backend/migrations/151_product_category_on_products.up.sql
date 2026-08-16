ALTER TABLE products
    ADD COLUMN IF NOT EXISTS product_category_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_product_category'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_product_category
            FOREIGN KEY (product_category_id)
            REFERENCES product_categories(id)
            ON DELETE SET NULL;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_products_product_category_id
    ON products(product_category_id);
