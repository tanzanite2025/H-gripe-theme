-- The repair-kit workflow is product-owned. It does not use the generic
-- product specification template/field editor. Keep the old 366 seed out of
-- the active catalog when it is still unused.
-- 366 created generic definitions before the workflow was simplified. They
-- are removed in dependency order before the unused template is removed.
-- Only retire the generic template when no product still references it. A
-- historical product must keep its old definitions and values readable while
-- new repair-kit products use the dedicated relation below.
DELETE FROM product_spec_values
WHERE spec_definition_id IN (
    SELECT definition.id
    FROM product_spec_definitions AS definition
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE template.slug = 'spoke_repair_kit'
      AND NOT EXISTS (
          SELECT 1
          FROM products
          WHERE products.product_specification_template_id = template.id
      )
);
DELETE FROM product_spec_option_items
WHERE spec_definition_id IN (
    SELECT definition.id
    FROM product_spec_definitions AS definition
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE template.slug = 'spoke_repair_kit'
      AND NOT EXISTS (
          SELECT 1
          FROM products
          WHERE products.product_specification_template_id = template.id
      )
);
DELETE FROM product_spec_definitions
WHERE product_specification_template_id IN (
    SELECT template.id
    FROM product_specification_templates AS template
    WHERE template.slug = 'spoke_repair_kit'
      AND NOT EXISTS (
          SELECT 1
          FROM products
          WHERE products.product_specification_template_id = template.id
      )
);
DELETE FROM product_specification_templates
WHERE slug = 'spoke_repair_kit'
  AND NOT EXISTS (
      SELECT 1 FROM products
      WHERE products.product_specification_template_id = product_specification_templates.id
  );

CREATE TABLE IF NOT EXISTS product_spoke_repair_kit_models (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    brand_slug VARCHAR(120) NOT NULL,
    brand_name VARCHAR(160) NOT NULL,
    wheelset_model_slug VARCHAR(160) NOT NULL,
    wheelset_model_name VARCHAR(255) NOT NULL,
    lifecycle_status VARCHAR(32) NOT NULL DEFAULT 'current',
    source_checked_at VARCHAR(40),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_product_spoke_repair_kit_model
        UNIQUE (product_id, brand_slug, wheelset_model_slug)
);

CREATE INDEX IF NOT EXISTS idx_product_spoke_repair_kit_models_product
    ON product_spoke_repair_kit_models (product_id, sort_order, id);

CREATE INDEX IF NOT EXISTS idx_product_spoke_repair_kit_models_model
    ON product_spoke_repair_kit_models (brand_slug, wheelset_model_slug);
