INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT 'brand_title', site_name.value, 'string', site_name.locale, 'site', true, 'Public brand title', NOW(), NOW()
FROM settings AS site_name
WHERE site_name.key = 'site_name'
  AND site_name."group" = 'site'
  AND NULLIF(BTRIM(site_name.value), '') IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM settings AS brand_title
    WHERE brand_title.key = 'brand_title'
      AND brand_title.locale = site_name.locale
  )
ON CONFLICT (key, locale) DO NOTHING;

UPDATE settings AS brand_title
SET value = site_name.value,
    description = 'Public brand title',
    updated_at = NOW()
FROM settings AS site_name
WHERE brand_title.key = 'brand_title'
  AND site_name.key = 'site_name'
  AND site_name."group" = 'site'
  AND brand_title.locale = site_name.locale
  AND NULLIF(BTRIM(brand_title.value), '') IS NULL
  AND NULLIF(BTRIM(site_name.value), '') IS NOT NULL;

UPDATE settings
SET description = 'Legacy alias for brand_title',
    updated_at = NOW()
WHERE key = 'site_name'
  AND "group" = 'site';

UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE key = 'site_logo'
  AND "group" = 'site'
  AND value = '/images/logo.png';
