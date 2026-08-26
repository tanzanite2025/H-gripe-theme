INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT
    'brand_title',
    site_name.value,
    'string',
    site_name.locale,
    'site',
    true,
    'Legacy storefront brand title',
    NOW(),
    NOW()
FROM settings AS site_name
WHERE site_name.key = 'site_name'
  AND site_name."group" = 'site'
  AND NOT EXISTS (
      SELECT 1
      FROM settings AS legacy
      WHERE legacy.key = 'brand_title'
        AND legacy.locale = site_name.locale
  )
ON CONFLICT (key, locale) DO NOTHING;
