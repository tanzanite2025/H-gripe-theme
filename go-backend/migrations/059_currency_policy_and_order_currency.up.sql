INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
  ('currency_accounting_currency', 'USD', 'string', 'en', 'currency', true, 'Internal accounting/base currency', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  ('currency_default_checkout_currency', 'USD', 'string', 'en', 'currency', true, 'Default currency locked onto new orders', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  ('currency_checkout_currencies', 'USD', 'string', 'en', 'currency', true, 'Currencies allowed for customer checkout', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key, locale) DO NOTHING;

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS currency VARCHAR(3);

UPDATE orders
SET currency = (
  SELECT value
  FROM settings
  WHERE key = 'currency_default_checkout_currency' AND locale = 'en'
  LIMIT 1
)
WHERE currency IS NULL OR BTRIM(currency) = '';

ALTER TABLE orders
  ALTER COLUMN currency SET NOT NULL;

ALTER TABLE orders
  ALTER COLUMN currency DROP DEFAULT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'orders_currency_iso_code'
  ) THEN
    ALTER TABLE orders
      ADD CONSTRAINT orders_currency_iso_code CHECK (currency ~ '^[A-Z]{3}$');
  END IF;
END $$;
