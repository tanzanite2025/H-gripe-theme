INSERT INTO product_categories (
    parent_id,
    name,
    slug,
    description,
    depth,
    sort_order,
    is_enabled,
    created_at,
    updated_at
)
VALUES (
    NULL,
    'Wheelsets',
    'wheelset',
    'System category for complete wheelset products used by guided selection and storefront filtering.',
    1,
    10,
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_enabled = TRUE,
    updated_at = NOW();

WITH wheelset_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheelset'
)
INSERT INTO product_category_translations (
    product_category_id,
    locale,
    name,
    description,
    created_at,
    updated_at
)
SELECT
    wheelset_category.id,
    seed.locale,
    seed.name,
    seed.description,
    NOW(),
    NOW()
FROM wheelset_category
CROSS JOIN (
    VALUES
        ('en', 'Wheelsets', 'Complete wheelset products.'),
        ('zh_cn', '轮组', '完整轮组产品。')
) AS seed(locale, name, description)
ON CONFLICT (product_category_id, locale) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();
