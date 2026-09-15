ALTER TABLE products
    DROP CONSTRAINT IF EXISTS chk_products_price_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_products_sale_price_minor_non_negative;
DROP INDEX IF EXISTS idx_products_price_currency;
ALTER TABLE products
    DROP COLUMN IF EXISTS sale_price_minor,
    DROP COLUMN IF EXISTS price_minor;

ALTER TABLE product_variants
    DROP CONSTRAINT IF EXISTS chk_product_variants_price_minor_non_negative,
    DROP CONSTRAINT IF EXISTS chk_product_variants_sale_price_minor_non_negative;
DROP INDEX IF EXISTS idx_product_variants_price_currency;
ALTER TABLE product_variants
    DROP COLUMN IF EXISTS sale_price_minor,
    DROP COLUMN IF EXISTS price_minor;
