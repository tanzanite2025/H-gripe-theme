-- Restore the pre-launch structure only when explicitly rolling the migration
-- back in a development database.

UPDATE faqs
SET page_id = 'shop-product-detail',
    updated_at = NOW()
WHERE page_id = 'products-product-detail';

INSERT INTO faq_pages (
    page_id, route_path, domain, locale, title, subtitle, sort_order, status
)
SELECT
    'shop-product-detail',
    '/shop/:slug',
    'products',
    locale,
    'Product Detail FAQs',
    'Common questions shown on individual product detail pages',
    120,
    'active'
FROM faq_pages
WHERE page_id = 'products-product-detail'
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    domain = EXCLUDED.domain,
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    deleted_at = NULL,
    updated_at = NOW();

INSERT INTO faq_categories (
    page_id, category_key, name, icon, locale, sort_order, status
)
SELECT
    'shop-product-detail',
    category_key,
    name,
    icon,
    locale,
    sort_order,
    'active'
FROM faq_categories
WHERE page_id = 'products-product-detail'
ON CONFLICT (page_id, category_key, locale) DO UPDATE
SET name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    deleted_at = NULL,
    updated_at = NOW();
