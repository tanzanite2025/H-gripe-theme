ALTER TABLE shipping_rules
  DROP CONSTRAINT IF EXISTS chk_shipping_rules_currency_iso;

ALTER TABLE shipping_templates
  DROP CONSTRAINT IF EXISTS chk_shipping_templates_currency_iso;

DROP INDEX IF EXISTS idx_shipping_rules_currency;
DROP INDEX IF EXISTS idx_shipping_templates_currency;

ALTER TABLE shipping_rules
  DROP COLUMN IF EXISTS currency;

ALTER TABLE shipping_templates
  DROP COLUMN IF EXISTS currency;
