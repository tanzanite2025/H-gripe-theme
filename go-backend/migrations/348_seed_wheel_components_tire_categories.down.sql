WITH tire_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'tire'
),
wheel_components_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheel-components'
)
DELETE FROM product_category_translations
WHERE product_category_id IN (
    SELECT id FROM tire_category
    UNION ALL
    SELECT id FROM wheel_components_category
)
AND locale IN ('en', 'zh_cn');

DELETE FROM product_categories
WHERE slug IN ('tire', 'wheel-components');
