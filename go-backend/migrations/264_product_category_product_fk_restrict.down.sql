ALTER TABLE products
    DROP CONSTRAINT IF EXISTS fk_products_product_category;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'products'
          AND column_name = 'product_category_id'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = current_schema()
          AND table_name = 'product_categories'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_product_category
            FOREIGN KEY (product_category_id)
            REFERENCES product_categories(id)
            ON DELETE SET NULL;
    END IF;
END
$$;
