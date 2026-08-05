INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT
  'currency_default_order_currency',
  value,
  type,
  locale,
  "group",
  is_public,
  'Default currency locked onto new orders',
  created_at,
  CURRENT_TIMESTAMP
FROM settings
WHERE key = 'currency_default_checkout_currency'
  AND "group" = 'currency'
ON CONFLICT (key, locale) DO UPDATE
SET value = EXCLUDED.value,
    type = EXCLUDED.type,
    "group" = EXCLUDED."group",
    is_public = EXCLUDED.is_public,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT
  'currency_accepted_currencies',
  value,
  type,
  locale,
  "group",
  is_public,
  'Business currencies accepted for order payment collection',
  created_at,
  CURRENT_TIMESTAMP
FROM settings
WHERE key = 'currency_checkout_currencies'
  AND "group" = 'currency'
ON CONFLICT (key, locale) DO UPDATE
SET value = EXCLUDED.value,
    type = EXCLUDED.type,
    "group" = EXCLUDED."group",
    is_public = EXCLUDED.is_public,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

DELETE FROM settings
WHERE key IN ('currency_default_checkout_currency', 'currency_checkout_currencies')
  AND "group" = 'currency';
