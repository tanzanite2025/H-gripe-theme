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
    substring(namespaced.key FROM char_length('website_name_') + 1),
    namespaced.value,
    namespaced.type,
    namespaced.locale,
    'website_name',
    namespaced.is_public,
    namespaced.description,
    namespaced.created_at,
    namespaced.updated_at
FROM settings AS namespaced
WHERE namespaced."group" = 'website_name'
  AND namespaced.key IN (
      'website_name_status',
      'website_name_intro',
      'website_name_eyebrow',
      'website_name_title',
      'website_name_body',
      'website_name_note'
  )
ON CONFLICT (key, locale) DO NOTHING;

DELETE FROM settings AS namespaced
WHERE namespaced."group" = 'website_name'
  AND namespaced.key IN (
      'website_name_status',
      'website_name_intro',
      'website_name_eyebrow',
      'website_name_title',
      'website_name_body',
      'website_name_note'
  )
  AND EXISTS (
      SELECT 1
      FROM settings AS legacy
      WHERE legacy.key = substring(namespaced.key FROM char_length('website_name_') + 1)
        AND legacy.locale = namespaced.locale
        AND legacy."group" = 'website_name'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM settings AS conflicting
      WHERE conflicting.key = substring(namespaced.key FROM char_length('website_name_') + 1)
        AND conflicting.locale = namespaced.locale
        AND conflicting."group" <> 'website_name'
  );
