ALTER TABLE products
  ADD COLUMN IF NOT EXISTS display_prices JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE product_variants
  ADD COLUMN IF NOT EXISTS display_prices JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE products
SET display_prices = '[]'::jsonb
WHERE display_prices IS NULL;

UPDATE product_variants
SET display_prices = '[]'::jsonb
WHERE display_prices IS NULL;
