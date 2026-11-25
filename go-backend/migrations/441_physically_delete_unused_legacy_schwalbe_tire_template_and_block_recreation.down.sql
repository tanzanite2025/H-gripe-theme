-- The deleted template and its old index are intentionally not reconstructed.
ALTER TABLE product_specification_templates
  DROP CONSTRAINT IF EXISTS retired_schwalbe_tire_template_slug_cannot_be_recreated;
