-- Move shipping template ownership onto products and product variants.
-- Product/variant fields are the only source of truth after this migration.

ALTER TABLE products
  ADD COLUMN IF NOT EXISTS shipping_template_id BIGINT;

ALTER TABLE product_variants
  ADD COLUMN IF NOT EXISTS shipping_template_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_products_shipping_template_id
  ON products (shipping_template_id);

CREATE INDEX IF NOT EXISTS idx_product_variants_shipping_template_id
  ON product_variants (shipping_template_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_products_shipping_template'
  ) THEN
    ALTER TABLE products
      ADD CONSTRAINT fk_products_shipping_template
      FOREIGN KEY (shipping_template_id)
      REFERENCES shipping_templates(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_product_variants_shipping_template'
  ) THEN
    ALTER TABLE product_variants
      ADD CONSTRAINT fk_product_variants_shipping_template
      FOREIGN KEY (shipping_template_id)
      REFERENCES shipping_templates(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL;
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.shipping_template_bindings') IS NOT NULL THEN
    WITH ranked_product_bindings AS (
      SELECT
        shipping_template_bindings.product_id,
        shipping_template_bindings.template_id,
        ROW_NUMBER() OVER (
          PARTITION BY shipping_template_bindings.product_id
          ORDER BY shipping_template_bindings.priority DESC, shipping_template_bindings.id DESC
        ) AS row_number
      FROM shipping_template_bindings
      INNER JOIN shipping_templates
        ON shipping_templates.id = shipping_template_bindings.template_id
       AND shipping_templates.deleted_at IS NULL
      WHERE shipping_template_bindings.deleted_at IS NULL
        AND shipping_template_bindings.enabled = TRUE
        AND shipping_template_bindings.scope = 'product'
        AND shipping_template_bindings.product_id IS NOT NULL
    )
    UPDATE products
    SET shipping_template_id = ranked_product_bindings.template_id
    FROM ranked_product_bindings
    WHERE products.id = ranked_product_bindings.product_id
      AND ranked_product_bindings.row_number = 1;

    WITH ranked_variant_bindings AS (
      SELECT
        shipping_template_bindings.variant_id,
        shipping_template_bindings.template_id,
        ROW_NUMBER() OVER (
          PARTITION BY shipping_template_bindings.variant_id
          ORDER BY shipping_template_bindings.priority DESC, shipping_template_bindings.id DESC
        ) AS row_number
      FROM shipping_template_bindings
      INNER JOIN shipping_templates
        ON shipping_templates.id = shipping_template_bindings.template_id
       AND shipping_templates.deleted_at IS NULL
      WHERE shipping_template_bindings.deleted_at IS NULL
        AND shipping_template_bindings.enabled = TRUE
        AND shipping_template_bindings.scope = 'variant'
        AND shipping_template_bindings.variant_id IS NOT NULL
    )
    UPDATE product_variants
    SET shipping_template_id = ranked_variant_bindings.template_id
    FROM ranked_variant_bindings
    WHERE product_variants.id = ranked_variant_bindings.variant_id
      AND ranked_variant_bindings.row_number = 1;

    DROP TABLE shipping_template_bindings;
  END IF;
END
$$;
