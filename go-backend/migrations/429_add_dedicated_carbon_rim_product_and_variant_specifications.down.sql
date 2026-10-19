DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM products WHERE product_type = 'carbon_rim' OR carbon_rim_specification IS NOT NULL)
    OR EXISTS (SELECT 1 FROM product_variants WHERE carbon_rim_specification IS NOT NULL) THEN
    RAISE EXCEPTION 'Cannot remove dedicated carbon rim specifications while product data exists';
  END IF;
END $$;
DROP INDEX idx_products_dedicated_carbon_rim_type;
ALTER TABLE product_variants DROP CONSTRAINT product_variants_carbon_rim_specification_object_check;
ALTER TABLE products DROP CONSTRAINT products_carbon_rim_type_and_specification_check;
ALTER TABLE product_variants DROP COLUMN carbon_rim_specification;
ALTER TABLE products DROP COLUMN carbon_rim_specification, DROP COLUMN product_type;
