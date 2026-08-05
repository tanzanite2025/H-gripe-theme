ALTER TABLE google_merchant_connections
  ADD COLUMN IF NOT EXISTS storefront_base_url VARCHAR(255) NOT NULL DEFAULT '';
