DROP INDEX IF EXISTS idx_product_variants_sku_active;
DROP INDEX IF EXISTS idx_product_slug_locale_active;
DROP INDEX IF EXISTS idx_products_sku_active;

ALTER TABLE products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);
ALTER TABLE product_variants
    ADD CONSTRAINT product_variants_sku_key UNIQUE (sku);
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_slug_locale ON products(slug, locale);
