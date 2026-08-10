UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'site_name'
  AND value IN ('Tanzanite', 'TANZANITE');

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key IN ('home_meta_title', 'article_meta_title', 'product_meta_title')
  AND value = 'Tanzanite - Premium E-commerce';

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key IN ('home_meta_description', 'article_meta_description', 'product_meta_description')
  AND value = 'Shop premium products at Tanzanite';

UPDATE settings
SET value = '', updated_at = NOW()
WHERE locale = 'en'
  AND key = 'from_name'
  AND value IN ('Tanzanite', 'TANZANITE');
