-- Replace global uniqueness with live-row uniqueness so soft-deleted catalog
-- records do not permanently reserve SKU or localized slug values.
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_sku_key;
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_sku_key;
DROP INDEX IF EXISTS idx_product_slug_locale;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku_active
    ON products(sku)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_slug_locale_active
    ON products(slug, locale)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_sku_active
    ON product_variants(sku)
    WHERE deleted_at IS NULL;
