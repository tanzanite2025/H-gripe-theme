-- Removed template rows and definitions are intentionally not reconstructed.
ALTER TABLE product_specification_templates
  DROP CONSTRAINT IF EXISTS retired_wheelset_template_slug_cannot_be_recreated;
