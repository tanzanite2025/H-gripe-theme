LOCK TABLE products, product_variants, product_shipping_template_bindings IN SHARE ROW EXCLUSIVE MODE;

-- Refuse to discard an effective selection for a live product without a SKU.
DO $$ BEGIN
  IF EXISTS (
    SELECT 1 FROM product_shipping_template_bindings b
    JOIN products p ON p.id = b.product_id AND p.deleted_at IS NULL
    WHERE b.variant_id IS NULL AND b.enabled AND b.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id AND v.deleted_at IS NULL)
  ) THEN
    RAISE EXCEPTION 'Create explicit SKUs for products with shipping bindings before retiring product-level shipping';
  END IF;
END $$;

-- Materialize the old effective selection once. Existing SKU selections win.
INSERT INTO product_shipping_template_bindings
  (product_id, variant_id, shipping_template_id, priority, enabled)
SELECT b.product_id, v.id, b.shipping_template_id, b.priority, TRUE
FROM product_shipping_template_bindings b
JOIN product_variants v ON v.product_id = b.product_id AND v.deleted_at IS NULL
WHERE b.variant_id IS NULL AND b.enabled AND b.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM product_shipping_template_bindings sku
    WHERE sku.product_id = v.product_id AND sku.variant_id = v.id
      AND sku.enabled AND sku.deleted_at IS NULL
  )
ON CONFLICT (product_id, variant_id, shipping_template_id)
  WHERE variant_id IS NOT NULL AND deleted_at IS NULL
DO UPDATE SET enabled = TRUE, priority = EXCLUDED.priority, updated_at = NOW();

DELETE FROM product_shipping_template_bindings WHERE variant_id IS NULL;
DROP INDEX IF EXISTS idx_product_shipping_template_bindings_product_scope;
ALTER TABLE product_shipping_template_bindings ALTER COLUMN variant_id SET NOT NULL;
ALTER TABLE product_shipping_template_bindings
  ADD CONSTRAINT shipping_template_binding_requires_sku CHECK (variant_id > 0);
