ALTER TABLE product_variants
  DROP COLUMN IF EXISTS display_prices;

ALTER TABLE products
  DROP COLUMN IF EXISTS display_prices;
