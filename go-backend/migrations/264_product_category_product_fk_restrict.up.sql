-- Keep category deletion consistent for databases that already ran migration 151.
ALTER TABLE products
    DROP CONSTRAINT IF EXISTS fk_products_product_category;

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
            ON DELETE RESTRICT;
    END IF;
END
$$;
