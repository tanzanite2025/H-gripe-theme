-- Create the reusable product specification template for spoke repair kits.
-- Wheelset compatibility is a separate product relation and is intentionally
-- not stored as a template field.

INSERT INTO product_specification_templates (
    name,
    slug,
    description,
    sort_order,
    is_enabled,
    is_system_managed
)
VALUES (
    'Spoke Repair Kit',
    'spoke_repair_kit',
    'Product template for replacement spokes and nipples. Wheelset compatibility is maintained separately.',
    70,
    TRUE,
    FALSE
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO product_spec_definitions (
    product_specification_template_id,
    "group",
    name,
    slug,
    field_type,
    unit,
    is_required,
    is_filterable,
    is_visible,
    role,
    selection_mode,
    min_selections,
    max_selections,
    presentation,
    sort_order,
    validation
)
SELECT template.id,
       seed."group",
       seed.name,
       seed.slug,
       seed.field_type,
       seed.unit,
       seed.is_required,
       seed.is_filterable,
       seed.is_visible,
       'attribute',
       'single',
       0,
       NULL,
       'text',
       seed.sort_order,
       ''
FROM product_specification_templates AS template
JOIN (
    VALUES
        ('Repair-kit specification', 'Spoke model', 'spoke_model', 'text', '', FALSE, FALSE, TRUE, 10),
        ('Repair-kit specification', 'Head type', 'head_type', 'select', '', FALSE, TRUE, TRUE, 20),
        ('Repair-kit specification', 'Spoke length(s)', 'spoke_length', 'text', 'mm', FALSE, FALSE, TRUE, 30),
        ('Repair-kit specification', 'Spoke diameter', 'spoke_diameter', 'text', 'mm', FALSE, TRUE, TRUE, 40),
        ('Repair-kit specification', 'Nipple model', 'nipple_model', 'text', '', FALSE, FALSE, TRUE, 50),
        ('Repair-kit specification', 'Nipple length(s)', 'nipple_length', 'text', 'mm', FALSE, FALSE, TRUE, 60),
        ('Repair-kit specification', 'Pack quantity', 'pack_quantity', 'number', 'pcs', FALSE, FALSE, TRUE, 70)
) AS seed("group", name, slug, field_type, unit, is_required, is_filterable, is_visible, sort_order)
    ON TRUE
WHERE template.slug = 'spoke_repair_kit'
  AND NOT EXISTS (
      SELECT 1
      FROM product_spec_definitions AS existing
      WHERE existing.product_specification_template_id = template.id
        AND existing.slug = seed.slug
  );

INSERT INTO product_spec_option_items (
    spec_definition_id,
    value_key,
    default_label,
    is_enabled_by_default,
    is_default,
    sort_order,
    revision
)
SELECT definition.id,
       seed.value_key,
       seed.default_label,
       TRUE,
       FALSE,
       seed.sort_order,
       1
FROM product_spec_definitions AS definition
JOIN product_specification_templates AS template
  ON template.id = definition.product_specification_template_id
JOIN (
    VALUES
        ('straight_pull', 'Straight-pull', 10),
        ('j_bend', 'J-bend', 20),
        ('unknown', 'Unknown', 30)
) AS seed(value_key, default_label, sort_order)
    ON TRUE
WHERE template.slug = 'spoke_repair_kit'
  AND definition.slug = 'head_type'
ON CONFLICT (spec_definition_id, value_key) DO NOTHING;
