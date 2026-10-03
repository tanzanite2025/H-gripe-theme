-- Add the repair-kit storefront category under the existing wheel-components
-- branch. Products will be assigned to this category in a later step.
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
    'Spoke Repair Kits',
    'spoke-repair-kits',
    'Replacement spoke and nipple repair-kit products for wheelsets.',
    2,
    20,
    TRUE,
    NOW(),
    NOW()
FROM wheel_components_category
ON CONFLICT (slug) DO UPDATE SET
    parent_id = EXCLUDED.parent_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    depth = EXCLUDED.depth,
    sort_order = EXCLUDED.sort_order,
    is_enabled = TRUE,
    updated_at = NOW();

WITH spoke_repair_category AS (
    SELECT id
    FROM product_categories
    WHERE slug = 'spoke-repair-kits'
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
    spoke_repair_category.id,
    seed.locale,
    seed.name,
    seed.description,
    NOW(),
    NOW()
FROM spoke_repair_category
CROSS JOIN (
    VALUES
        ('en', 'Spoke Repair Kits', 'Replacement spoke and nipple repair-kit products.'),
        ('zh_cn', '辐条修补件', '用于轮组维修的辐条和条帽修补包产品。')
) AS seed(locale, name, description)
ON CONFLICT (product_category_id, locale) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();
