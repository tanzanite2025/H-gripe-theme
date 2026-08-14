DELETE FROM settings
WHERE key = 'site_favicon'
  AND locale = 'en'
  AND "group" = 'site';
