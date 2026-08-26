INSERT INTO settings (
    key,
    value,
    type,
    locale,
    "group",
    is_public,
    description,
    created_at,
    updated_at
)
SELECT
    'website_name_' || legacy.key,
    legacy.value,
    legacy.type,
    legacy.locale,
    'website_name',
    legacy.is_public,
    legacy.description,
    legacy.created_at,
    legacy.updated_at
FROM settings AS legacy
WHERE legacy."group" = 'website_name'
  AND legacy.key IN ('status', 'intro', 'eyebrow', 'title', 'body', 'note')
ON CONFLICT (key, locale) DO NOTHING;

UPDATE settings AS namespaced
SET
    value = legacy.value,
    type = legacy.type,
    is_public = legacy.is_public,
    description = legacy.description,
    updated_at = NOW()
FROM settings AS legacy
WHERE namespaced."group" = 'website_name'
  AND namespaced.key IN (
      'website_name_status',
      'website_name_intro',
      'website_name_eyebrow',
      'website_name_title',
      'website_name_body',
      'website_name_note'
  )
  AND legacy."group" = 'website_name'
  AND legacy.key = substring(namespaced.key FROM char_length('website_name_') + 1)
  AND legacy.locale = namespaced.locale
  AND NULLIF(BTRIM(namespaced.value), '') IS NULL
  AND NULLIF(BTRIM(legacy.value), '') IS NOT NULL;

DELETE FROM settings
WHERE "group" = 'website_name'
  AND key IN ('status', 'intro', 'eyebrow', 'title', 'body', 'note');
