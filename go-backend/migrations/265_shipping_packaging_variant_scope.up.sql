-- Allow a packaging rule to target one SKU while retaining a product default.
ALTER TABLE shipping_packaging_rule_applies
    ADD COLUMN IF NOT EXISTS variant_id BIGINT;

DROP INDEX IF EXISTS idx_shipping_packaging_rule_apply_rule_product;
DROP INDEX IF EXISTS idx_shipping_packaging_rule_apply_product;

CREATE INDEX IF NOT EXISTS idx_shipping_packaging_rule_applies_variant_id
    ON shipping_packaging_rule_applies (variant_id);

-- PostgreSQL treats NULLs as distinct in a normal unique index, so keep the
-- product default and variant targets in separate partial unique indexes.
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_rule_product_variant
    ON shipping_packaging_rule_applies (rule_id, product_id, variant_id)
    WHERE variant_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_rule_product_default
    ON shipping_packaging_rule_applies (rule_id, product_id)
    WHERE variant_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_product_variant
    ON shipping_packaging_rule_applies (product_id, variant_id)
    WHERE variant_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_packaging_rule_apply_product_default
    ON shipping_packaging_rule_applies (product_id)
    WHERE variant_id IS NULL;
