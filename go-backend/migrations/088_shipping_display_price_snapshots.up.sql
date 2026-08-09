ALTER TABLE shipping_templates
  ADD COLUMN IF NOT EXISTS display_price_snapshots JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE shipping_rules
  ADD COLUMN IF NOT EXISTS display_price_snapshots JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE shipping_templates
SET display_price_snapshots = '{}'::jsonb
WHERE display_price_snapshots IS NULL;

UPDATE shipping_rules
SET display_price_snapshots = '{}'::jsonb
WHERE display_price_snapshots IS NULL;
