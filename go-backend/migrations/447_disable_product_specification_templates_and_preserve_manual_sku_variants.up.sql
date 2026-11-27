LOCK TABLE product_specification_templates, products, product_spec_definitions,
  product_spec_values, product_spec_option_items, product_variant_option_values,
  product_custom_option_policies, product_option_group_variant_rules,
  product_option_value_variant_rules, product_option_value_relations,
  quick_buy_step_product_specification_templates, quick_buy_step_filters,
  product_media, product_variants
  IN SHARE ROW EXCLUSIVE MODE;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM products WHERE product_specification_template_id IS NOT NULL)
     OR EXISTS (SELECT 1 FROM product_spec_values)
     OR EXISTS (SELECT 1 FROM product_spec_option_items)
     OR EXISTS (SELECT 1 FROM product_variant_option_values)
     OR EXISTS (SELECT 1 FROM product_custom_option_policies)
     OR EXISTS (SELECT 1 FROM product_option_group_variant_rules)
     OR EXISTS (SELECT 1 FROM product_option_value_variant_rules)
     OR EXISTS (SELECT 1 FROM product_option_value_relations)
     OR EXISTS (SELECT 1 FROM product_media WHERE variant_option_value_id IS NOT NULL)
     OR EXISTS (SELECT 1 FROM product_variants WHERE option_values <> '{}'::jsonb)
     OR EXISTS (SELECT 1 FROM quick_buy_step_product_specification_templates)
     OR EXISTS (SELECT 1 FROM quick_buy_step_filters WHERE spec_definition_id IS NOT NULL) THEN
    RAISE EXCEPTION 'Product specification templates still have product, SKU, media, shipping, or Quick Buy references; resolve them before disabling the template system';
  END IF;
END $$;

DELETE FROM product_spec_option_items;
DELETE FROM product_spec_definitions;
DELETE FROM product_specification_templates;

ALTER TABLE product_specification_templates
  ADD CONSTRAINT product_specification_templates_disabled_check CHECK (FALSE);

ALTER TABLE product_spec_definitions
  ADD CONSTRAINT product_spec_definitions_disabled_check CHECK (FALSE);

ALTER TABLE product_spec_values
  ADD CONSTRAINT product_spec_values_disabled_check CHECK (FALSE);

ALTER TABLE product_spec_option_items
  ADD CONSTRAINT product_spec_option_items_disabled_check CHECK (FALSE);

ALTER TABLE product_variant_option_values
  ADD CONSTRAINT product_variant_option_values_disabled_check CHECK (FALSE);

ALTER TABLE product_custom_option_policies
  ADD CONSTRAINT product_custom_option_policies_disabled_check CHECK (FALSE);

ALTER TABLE product_option_group_variant_rules
  ADD CONSTRAINT product_option_group_variant_rules_disabled_check CHECK (FALSE);

ALTER TABLE product_option_value_variant_rules
  ADD CONSTRAINT product_option_value_variant_rules_disabled_check CHECK (FALSE);

ALTER TABLE product_option_value_relations
  ADD CONSTRAINT product_option_value_relations_disabled_check CHECK (FALSE);

ALTER TABLE quick_buy_step_product_specification_templates
  ADD CONSTRAINT quick_buy_step_product_specification_templates_disabled_check CHECK (FALSE);

ALTER TABLE quick_buy_step_filters
  ADD CONSTRAINT quick_buy_step_filters_no_specification_check CHECK (spec_definition_id IS NULL);

ALTER TABLE product_media
  ADD CONSTRAINT product_media_no_template_option_check CHECK (variant_option_value_id IS NULL);

ALTER TABLE product_variants
  ADD CONSTRAINT product_variants_empty_template_options_check CHECK (option_values = '{}'::jsonb);

ALTER TABLE products
  ADD CONSTRAINT products_product_specification_template_retired_check
  CHECK (product_specification_template_id IS NULL);

DROP INDEX IF EXISTS idx_product_variant_options;
CREATE UNIQUE INDEX idx_product_variant_options_nonempty
  ON product_variants(product_id, option_values)
  WHERE deleted_at IS NULL AND option_values <> '{}'::jsonb;
