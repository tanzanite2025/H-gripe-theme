-- This migration retires the former product specification template only.
-- Schwalbe catalog and hookless compatibility tables remain in service.
LOCK TABLE product_specification_templates, products, product_spec_definitions,
  product_spec_values, product_variant_option_values, product_option_group_variant_rules,
  quick_buy_step_product_specification_templates, quick_buy_step_filters,
  product_spec_option_items
  IN SHARE ROW EXCLUSIVE MODE;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM products AS product
    JOIN product_specification_templates AS template
      ON template.id = product.product_specification_template_id
    WHERE lower(btrim(template.slug)) = 'schwalbe_tire'
  ) OR EXISTS (
    SELECT 1
    FROM product_spec_definitions AS definition
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE lower(btrim(template.slug)) = 'schwalbe_tire'
      AND (
        EXISTS (SELECT 1 FROM product_spec_values AS value WHERE value.spec_definition_id = definition.id)
        OR EXISTS (SELECT 1 FROM product_variant_option_values AS value WHERE value.spec_definition_id = definition.id)
        OR EXISTS (SELECT 1 FROM product_option_group_variant_rules AS rule WHERE rule.spec_definition_id = definition.id)
        OR EXISTS (SELECT 1 FROM quick_buy_step_filters AS filter WHERE filter.spec_definition_id = definition.id)
      )
  ) OR EXISTS (
    SELECT 1
    FROM quick_buy_step_product_specification_templates AS quick_buy_template
    JOIN product_specification_templates AS template
      ON template.id = quick_buy_template.product_specification_template_id
    WHERE lower(btrim(template.slug)) = 'schwalbe_tire'
  ) OR EXISTS (
    SELECT 1
    FROM product_variant_option_values AS variant_option_value
    JOIN product_spec_option_items AS option_item
      ON option_item.id = variant_option_value.template_option_item_id
    JOIN product_spec_definitions AS definition
      ON definition.id = option_item.spec_definition_id
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE lower(btrim(template.slug)) = 'schwalbe_tire'
  ) THEN
    RAISE EXCEPTION 'Legacy Schwalbe tire template still has product or configuration references; resolve them explicitly before physical deletion';
  END IF;
END $$;

DROP INDEX IF EXISTS uq_schwalbe_product_article_no;

DELETE FROM product_spec_option_items
WHERE spec_definition_id IN (
  SELECT definition.id
  FROM product_spec_definitions AS definition
  JOIN product_specification_templates AS template
    ON template.id = definition.product_specification_template_id
  WHERE lower(btrim(template.slug)) = 'schwalbe_tire'
);

DELETE FROM product_spec_definitions
WHERE product_specification_template_id IN (
  SELECT id FROM product_specification_templates
  WHERE lower(btrim(slug)) = 'schwalbe_tire'
);

DELETE FROM product_specification_templates
WHERE lower(btrim(slug)) = 'schwalbe_tire';

ALTER TABLE product_specification_templates
  ADD CONSTRAINT retired_schwalbe_tire_template_slug_cannot_be_recreated
  CHECK (lower(btrim(slug)) <> 'schwalbe_tire');
