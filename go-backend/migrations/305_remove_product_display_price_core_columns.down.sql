ALTER TABLE products
    ADD COLUMN IF NOT EXISTS display_prices JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS display_prices JSONB NOT NULL DEFAULT '[]'::jsonb;
