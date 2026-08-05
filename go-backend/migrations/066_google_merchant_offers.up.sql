CREATE TABLE IF NOT EXISTS google_merchant_offers (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  variant_id BIGINT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
  offer_id VARCHAR(160) NOT NULL UNIQUE,
  title VARCHAR(150) NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  brand VARCHAR(140) NOT NULL DEFAULT '',
  condition VARCHAR(24) NOT NULL DEFAULT 'new',
  google_product_category TEXT NOT NULL DEFAULT '',
  gtin VARCHAR(20) NOT NULL DEFAULT '',
  mpn VARCHAR(70) NOT NULL DEFAULT '',
  identifier_exists BOOLEAN,
  target_country VARCHAR(2) NOT NULL DEFAULT '',
  content_language VARCHAR(8) NOT NULL DEFAULT '',
  currency_code VARCHAR(3) NOT NULL DEFAULT '',
  feed_label VARCHAR(64) NOT NULL DEFAULT '',
  price_override NUMERIC(12, 2),
  sale_price_override NUMERIC(12, 2),
  publication_status VARCHAR(24) NOT NULL DEFAULT 'draft',
  sync_status VARCHAR(24) NOT NULL DEFAULT 'not_synced',
  last_validated_at TIMESTAMPTZ,
  last_sync_at TIMESTAMPTZ,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT uq_google_merchant_offer_variant UNIQUE (variant_id)
);

CREATE INDEX IF NOT EXISTS idx_google_merchant_offers_product ON google_merchant_offers (product_id);
CREATE INDEX IF NOT EXISTS idx_google_merchant_offers_status ON google_merchant_offers (publication_status, sync_status);
