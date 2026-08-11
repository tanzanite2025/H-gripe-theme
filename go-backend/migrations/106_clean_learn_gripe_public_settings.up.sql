-- Finish the storefront domain cutover in persisted public settings.
-- Keep this migration narrow: never rewrite the separate ERP hostname.

UPDATE settings
SET value = REPLACE(value, 'https://admin.tanzanite.site', 'https://admin.learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%https://admin.tanzanite.site%';

UPDATE settings
SET value = REPLACE(value, 'http://admin.tanzanite.site', 'https://admin.learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%http://admin.tanzanite.site%';

UPDATE settings
SET value = REPLACE(value, 'https://www.tanzanite.site', 'https://www.learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%https://www.tanzanite.site%';

UPDATE settings
SET value = REPLACE(value, 'http://www.tanzanite.site', 'https://www.learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%http://www.tanzanite.site%';

UPDATE settings
SET value = REPLACE(value, 'https://tanzanite.site', 'https://learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%https://tanzanite.site%';

UPDATE settings
SET value = REPLACE(value, 'http://tanzanite.site', 'https://learn.gripe'),
    updated_at = NOW()
WHERE value LIKE '%http://tanzanite.site%';

UPDATE settings
SET value = 'H-GRIPE',
    updated_at = NOW()
WHERE "group" = 'site'
  AND key IN ('site_name', 'brand_title')
  AND value IN ('Store', 'Tanzanite', 'TANZANITE');
