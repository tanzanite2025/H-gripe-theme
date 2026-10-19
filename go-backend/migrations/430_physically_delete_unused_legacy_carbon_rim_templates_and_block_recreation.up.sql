-- Refuse to cascade-delete any product or purchasing configuration.
LOCK TABLE product_specification_templates, products, product_spec_definitions,
  product_spec_values, product_variant_option_values, product_option_group_variant_rules,
  quick_buy_step_product_specification_templates, quick_buy_step_filters, product_spec_option_items
  IN SHARE ROW EXCLUSIVE MODE;
DO $$ BEGIN
  IF EXISTS (
    SELECT 1 FROM products p JOIN product_specification_templates t ON t.id=p.product_specification_template_id
    WHERE lower(trim(t.slug)) IN ('rim','carbon_rim','carbon-rim')
  ) OR EXISTS (
    SELECT 1 FROM product_spec_definitions d JOIN product_specification_templates t ON t.id=d.product_specification_template_id
    WHERE lower(trim(t.slug)) IN ('rim','carbon_rim','carbon-rim') AND (
      EXISTS (SELECT 1 FROM product_spec_values v WHERE v.spec_definition_id=d.id) OR
      EXISTS (SELECT 1 FROM product_variant_option_values v WHERE v.spec_definition_id=d.id) OR
      EXISTS (SELECT 1 FROM product_option_group_variant_rules r WHERE r.spec_definition_id=d.id) OR
      EXISTS (SELECT 1 FROM quick_buy_step_filters f WHERE f.spec_definition_id=d.id)
    )
  ) OR EXISTS (
    SELECT 1 FROM quick_buy_step_product_specification_templates q JOIN product_specification_templates t ON t.id=q.product_specification_template_id
    WHERE lower(trim(t.slug)) IN ('rim','carbon_rim','carbon-rim')
  ) THEN
    RAISE EXCEPTION 'Legacy rim templates still have product or purchase configuration references; migrate these explicitly before physical deletion';
  END IF;
END $$;
DELETE FROM product_spec_option_items WHERE spec_definition_id IN (
  SELECT d.id FROM product_spec_definitions d JOIN product_specification_templates t ON t.id=d.product_specification_template_id
  WHERE lower(trim(t.slug)) IN ('rim','carbon_rim','carbon-rim')
);
DELETE FROM product_spec_definitions WHERE product_specification_template_id IN (
  SELECT id FROM product_specification_templates WHERE lower(trim(slug)) IN ('rim','carbon_rim','carbon-rim')
);
DELETE FROM product_specification_templates WHERE lower(trim(slug)) IN ('rim','carbon_rim','carbon-rim');
ALTER TABLE product_specification_templates ADD CONSTRAINT retired_carbon_rim_template_slugs_cannot_be_recreated
  CHECK (lower(trim(slug)) NOT IN ('rim','carbon_rim','carbon-rim'));
