-- Catalog/SKU prices are the sole source prices. Discounts are applied by
-- Checkout coupon rules and must not be stored as a second catalog price.
ALTER TABLE products DROP COLUMN IF EXISTS sale_price_minor;
ALTER TABLE product_variants DROP COLUMN IF EXISTS sale_price_minor;
ALTER TABLE google_merchant_offers DROP COLUMN IF EXISTS sale_price_override_minor;
ALTER TABLE product_profit_calculations DROP COLUMN IF EXISTS sale_price_minor;

-- Converted snapshots may contain the former sale price. Rebuild them from
-- the catalog price rather than trusting an obsolete source fingerprint.
DELETE FROM product_display_price_snapshots;
ALTER TABLE product_display_price_snapshots DROP COLUMN IF EXISTS source_sale_price_minor;
