ALTER TABLE products ADD COLUMN product_type VARCHAR(40) NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN carbon_rim_specification JSONB;
ALTER TABLE product_variants ADD COLUMN carbon_rim_specification JSONB;
ALTER TABLE products ADD CONSTRAINT products_carbon_rim_type_and_specification_check CHECK (
  (product_type = '' AND carbon_rim_specification IS NULL) OR
  (product_type = 'carbon_rim' AND product_specification_template_id IS NULL
    AND carbon_rim_specification IS NOT NULL
    AND jsonb_typeof(carbon_rim_specification) = 'object'
    AND (carbon_rim_specification->>'schema_version') IS NOT DISTINCT FROM '1')
);
ALTER TABLE product_variants ADD CONSTRAINT product_variants_carbon_rim_specification_object_check
  CHECK (carbon_rim_specification IS NULL OR jsonb_typeof(carbon_rim_specification) = 'object');
CREATE INDEX idx_products_dedicated_carbon_rim_type ON products(product_type) WHERE product_type = 'carbon_rim' AND deleted_at IS NULL;
