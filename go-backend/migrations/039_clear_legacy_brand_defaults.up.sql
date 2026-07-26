UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'site_name'
  AND value IN ('Tanzanite', 'TANZANITE');

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'meta_title'
  AND value = 'Tanzanite - Premium E-commerce';

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'meta_description'
  AND value = 'Shop premium products at Tanzanite';

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'from_name'
  AND value IN ('Tanzanite', 'TANZANITE');
