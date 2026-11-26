-- Retire the old wheelset template only when no product uses it.
LOCK TABLE product_specification_templates, products, product_spec_definitions,
  product_spec_option_items, product_spec_values, product_variant_option_values,
  product_option_group_variant_rules, product_option_value_variant_rules,
  product_option_value_relations, product_custom_option_policies,
  quick_buy_step_product_specification_templates, quick_buy_step_filters
  IN SHARE ROW EXCLUSIVE MODE;

DO $$
DECLARE
    wheelset_template_count BIGINT;
BEGIN
    SELECT COUNT(*)
    INTO wheelset_template_count
    FROM product_specification_templates
    WHERE lower(btrim(slug)) = 'wheelset';

    IF wheelset_template_count > 1 THEN
        RAISE EXCEPTION 'Multiple legacy wheelset templates exist; resolve duplicate template identities before retirement';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM products AS product
        JOIN product_specification_templates AS template
          ON template.id = product.product_specification_template_id
        WHERE lower(btrim(template.slug)) = 'wheelset'
    ) THEN
        RAISE EXCEPTION 'A product still uses the legacy wheelset template; compatibility is unsupported and all products must be removed before retirement';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM product_spec_definitions AS definition
        JOIN product_specification_templates AS template
          ON template.id = definition.product_specification_template_id
        WHERE lower(btrim(template.slug)) = 'wheelset'
          AND (
            EXISTS (SELECT 1 FROM product_spec_values AS value WHERE value.spec_definition_id = definition.id)
            OR EXISTS (SELECT 1 FROM product_variant_option_values AS value WHERE value.spec_definition_id = definition.id)
            OR EXISTS (SELECT 1 FROM product_option_group_variant_rules AS rule WHERE rule.spec_definition_id = definition.id)
            OR EXISTS (SELECT 1 FROM quick_buy_step_filters AS filter WHERE filter.spec_definition_id = definition.id)
          )
    ) OR EXISTS (
        SELECT 1
        FROM product_variant_option_values AS variant_option_value
        JOIN product_spec_option_items AS option_item
          ON option_item.id = variant_option_value.template_option_item_id
        JOIN product_spec_definitions AS definition
          ON definition.id = option_item.spec_definition_id
        JOIN product_specification_templates AS template
          ON template.id = definition.product_specification_template_id
        WHERE lower(btrim(template.slug)) = 'wheelset'
    ) THEN
        RAISE EXCEPTION 'The legacy wheelset template still has product specification or variant configuration references';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM quick_buy_step_product_specification_templates AS quick_buy_template
        JOIN product_specification_templates AS template
          ON template.id = quick_buy_template.product_specification_template_id
        WHERE lower(btrim(template.slug)) = 'wheelset'
    ) THEN
        RAISE EXCEPTION 'Quick Buy still binds the legacy wheelset template; switch it to the wheelset category before retirement';
    END IF;
END $$;

DELETE FROM product_spec_option_items
WHERE spec_definition_id IN (
    SELECT definition.id
    FROM product_spec_definitions AS definition
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE lower(btrim(template.slug)) = 'wheelset'
);

DELETE FROM product_spec_definitions
WHERE product_specification_template_id IN (
    SELECT id
    FROM product_specification_templates
    WHERE lower(btrim(slug)) = 'wheelset'
);

DELETE FROM product_specification_templates
WHERE lower(btrim(slug)) = 'wheelset';

ALTER TABLE product_specification_templates
  ADD CONSTRAINT retired_wheelset_template_slug_cannot_be_recreated
  CHECK (lower(btrim(slug)) <> 'wheelset');

COMMENT ON CONSTRAINT retired_wheelset_template_slug_cannot_be_recreated
  ON product_specification_templates IS
  'The wheelset product type uses products.wheelset_specification; the legacy generic template is retired.';
