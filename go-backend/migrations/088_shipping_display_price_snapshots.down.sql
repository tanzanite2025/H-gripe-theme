ALTER TABLE shipping_rules
  DROP COLUMN IF EXISTS display_price_snapshots;

ALTER TABLE shipping_templates
  DROP COLUMN IF EXISTS display_price_snapshots;
