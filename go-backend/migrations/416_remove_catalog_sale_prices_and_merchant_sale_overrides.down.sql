-- Recreate nullable fields for schema rollback. Removed promotional amounts
-- are not restored or synthesized from catalog prices.
ALTER TABLE products ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT;
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT;
ALTER TABLE google_merchant_offers ADD COLUMN IF NOT EXISTS sale_price_override_minor BIGINT;
ALTER TABLE product_profit_calculations ADD COLUMN IF NOT EXISTS sale_price_minor BIGINT;
ALTER TABLE product_display_price_snapshots ADD COLUMN IF NOT EXISTS source_sale_price_minor BIGINT;
