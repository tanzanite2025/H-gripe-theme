-- Historical migration 202. Keep stable after it has been applied.

DELETE FROM settings
WHERE "group" = 'website_name'
  AND key IN (
      'website_name_status',
      'website_name_intro',
      'website_name_eyebrow',
      'website_name_title',
      'website_name_body',
      'website_name_note'
  )
  AND locale IN ('en', 'zh_cn')
  AND value = '';
