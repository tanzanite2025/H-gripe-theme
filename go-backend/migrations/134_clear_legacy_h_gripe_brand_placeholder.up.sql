-- Brand values are optional storefront settings. Remove the retired placeholder
-- so an unconfigured footer renders no brand name instead of H-GRIPE.
UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE "group" = 'site'
  AND key IN ('site_name', 'brand_title')
  AND BTRIM(value) IN ('H-GRIPE', 'H GRIPE');
