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
    'Wheel Components',
    'wheel-components',
    'System category for wheel components used by storefront category filtering.',
    1,
    20,
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_enabled = TRUE,
    updated_at = NOW();

WITH wheel_components_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheel-components'
)
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
SELECT
    wheel_components_category.id,
    'Tires',
    'tire',
    'System category for tire products used by the tire guide product search sheet.',
    2,
    10,
    TRUE,
    NOW(),
    NOW()
FROM wheel_components_category
ON CONFLICT (slug) DO UPDATE SET
    parent_id = EXCLUDED.parent_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    depth = 2,
    is_enabled = TRUE,
    updated_at = NOW();

WITH wheel_components_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'wheel-components'
),
tire_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'tire'
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
    seed.product_category_id,
    seed.locale,
    seed.name,
    seed.description,
    NOW(),
    NOW()
FROM (
    SELECT wheel_components_category.id AS product_category_id,
        'en' AS locale,
        'Wheel Components' AS name,
        'Wheel component products.' AS description
    FROM wheel_components_category
    UNION ALL
    SELECT wheel_components_category.id,
        'zh_cn',
        '轮组配件',
        '轮组配件产品。'
    FROM wheel_components_category
    UNION ALL
    SELECT tire_category.id,
        'en',
        'Tires',
        'Tire products.'
    FROM tire_category
    UNION ALL
    SELECT tire_category.id,
        'zh_cn',
        '外胎',
        '外胎产品。'
    FROM tire_category
) AS seed
ON CONFLICT (product_category_id, locale) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();
