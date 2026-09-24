-- Display prices are a storefront read model. Product source rows must only
-- contain transactional pricing and never carry mutable converted JSON.
ALTER TABLE products
    DROP COLUMN IF EXISTS display_prices;

ALTER TABLE product_variants
    DROP COLUMN IF EXISTS display_prices;
