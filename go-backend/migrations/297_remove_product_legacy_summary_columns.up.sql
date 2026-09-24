-- Products are catalog containers. SKU, price, sale_price and stock belong
-- to product_variants and must not remain as competing transaction facts.

DROP INDEX IF EXISTS idx_products_sku_active;
DROP INDEX IF EXISTS idx_products_sku;

ALTER TABLE products
    DROP COLUMN IF EXISTS sku,
    DROP COLUMN IF EXISTS price,
    DROP COLUMN IF EXISTS sale_price,
    DROP COLUMN IF EXISTS stock;
