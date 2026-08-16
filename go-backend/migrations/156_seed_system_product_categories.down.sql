WITH wheelset_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheelset'
)
DELETE FROM product_category_translations
WHERE product_category_id IN (SELECT id FROM wheelset_category)
  AND locale IN ('en', 'zh_cn');

DELETE FROM product_categories
WHERE slug = 'wheelset';
