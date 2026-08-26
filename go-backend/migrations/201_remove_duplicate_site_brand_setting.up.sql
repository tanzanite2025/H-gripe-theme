INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT
    'site_name',
    legacy.value,
    'string',
    legacy.locale,
    'site',
    true,
    'Site name',
    NOW(),
    NOW()
FROM settings AS legacy
WHERE legacy.key = 'brand_title'
  AND legacy."group" = 'site'
  AND NULLIF(BTRIM(legacy.value), '') IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM settings AS current_site
      WHERE current_site.key = 'site_name'
        AND current_site.locale = legacy.locale
  )
ON CONFLICT (key, locale) DO NOTHING;

UPDATE settings AS current_site
SET value = legacy.value,
    type = 'string',
    "group" = 'site',
    is_public = true,
    description = 'Site name',
    updated_at = NOW()
FROM settings AS legacy
WHERE current_site.key = 'site_name'
  AND current_site."group" = 'site'
  AND current_site.locale = legacy.locale
  AND legacy.key = 'brand_title'
  AND legacy."group" = 'site'
  AND NULLIF(BTRIM(current_site.value), '') IS NULL
  AND NULLIF(BTRIM(legacy.value), '') IS NOT NULL;

DELETE FROM settings
WHERE key = 'brand_title'
  AND "group" = 'site';
