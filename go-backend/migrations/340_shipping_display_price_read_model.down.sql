ALTER TABLE shipping_templates
    ADD COLUMN IF NOT EXISTS display_price_snapshots JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE shipping_rules
    ADD COLUMN IF NOT EXISTS display_price_snapshots JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE shipping_templates t
SET display_price_snapshots = COALESCE(s.display_prices, '{}'::jsonb)
FROM shipping_display_price_snapshots s
WHERE s.template_id = t.id
  AND s.rule_id IS NULL;

UPDATE shipping_rules r
SET display_price_snapshots = COALESCE(s.display_prices, '{}'::jsonb)
FROM shipping_display_price_snapshots s
WHERE s.rule_id = r.id;

DROP TABLE IF EXISTS shipping_display_price_snapshots;
