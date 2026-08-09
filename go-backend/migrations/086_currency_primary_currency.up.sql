INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES (
  'currency_primary_currency',
  'USD',
  'string',
  'en',
  'currency',
  true,
  'Primary pricing currency for product, SKU, shipping, and commercial amounts',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (key, locale) DO UPDATE
SET type = EXCLUDED.type,
    "group" = EXCLUDED."group",
    is_public = EXCLUDED.is_public,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;
