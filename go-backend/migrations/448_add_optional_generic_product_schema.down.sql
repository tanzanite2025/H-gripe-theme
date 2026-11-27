DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM products WHERE generic_product_schema IS NOT NULL) THEN
    RAISE EXCEPTION 'Cannot drop generic product schema while products contain structured facts';
  END IF;
END $$;

ALTER TABLE products
  DROP CONSTRAINT IF EXISTS products_generic_product_schema_check;

ALTER TABLE products
  DROP COLUMN IF EXISTS generic_product_schema;
