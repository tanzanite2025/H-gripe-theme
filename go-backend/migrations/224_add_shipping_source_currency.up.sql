ALTER TABLE shipping_templates
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

ALTER TABLE shipping_rules
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE shipping_templates
SET currency = COALESCE(
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_primary_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_default_checkout_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  NULLIF(UPPER(TRIM(currency)), ''),
  'USD'
);

UPDATE shipping_rules sr
SET currency = COALESCE(
  (
    SELECT NULLIF(UPPER(TRIM(st.currency)), '')
    FROM shipping_templates st
    WHERE st.id = sr.template_id
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_primary_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  (
    SELECT NULLIF(UPPER(TRIM(value)), '')
    FROM settings
    WHERE key = 'currency_default_checkout_currency'
      AND locale = 'en'
    LIMIT 1
  ),
  NULLIF(UPPER(TRIM(sr.currency)), ''),
  'USD'
);

CREATE INDEX IF NOT EXISTS idx_shipping_templates_currency ON shipping_templates(currency);
CREATE INDEX IF NOT EXISTS idx_shipping_rules_currency ON shipping_rules(currency);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_shipping_templates_currency_iso'
  ) THEN
    ALTER TABLE shipping_templates
      ADD CONSTRAINT chk_shipping_templates_currency_iso CHECK (currency ~ '^[A-Z]{3}$');
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_shipping_rules_currency_iso'
  ) THEN
    ALTER TABLE shipping_rules
      ADD CONSTRAINT chk_shipping_rules_currency_iso CHECK (currency ~ '^[A-Z]{3}$');
  END IF;
END $$;
