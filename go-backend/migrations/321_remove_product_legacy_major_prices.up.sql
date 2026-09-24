-- Product source prices are now represented exclusively by integer minor units.
-- The legacy major-unit columns are removed before launch so no ORM query or
-- catalog worker can accidentally reintroduce float-based pricing.
ALTER TABLE products
    DROP COLUMN IF EXISTS price,
    DROP COLUMN IF EXISTS sale_price;

ALTER TABLE product_variants
    DROP CONSTRAINT IF EXISTS chk_product_variants_price_positive,
    DROP CONSTRAINT IF EXISTS chk_product_variants_sale_price_non_negative,
    DROP COLUMN IF EXISTS price,
    DROP COLUMN IF EXISTS sale_price;
