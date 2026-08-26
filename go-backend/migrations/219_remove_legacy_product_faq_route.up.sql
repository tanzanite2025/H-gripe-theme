-- Product detail FAQ lookup is owned by /products/:slug.
-- /shop/... is reserved for category routes; the pre-launch product FAQ
-- structure must not remain in the database.

UPDATE faqs
SET page_id = 'products-product-detail',
    updated_at = NOW()
WHERE page_id = 'shop-product-detail';

DELETE FROM faq_categories
WHERE page_id = 'shop-product-detail';

UPDATE faq_pages
SET route_path = '/products/:slug',
    domain = 'products',
    title = 'Product Detail FAQs',
    subtitle = 'Common questions shown on individual product detail pages',
    status = 'active',
    deleted_at = NULL,
    updated_at = NOW()
WHERE page_id = 'products-product-detail';

DELETE FROM faq_pages
WHERE page_id = 'shop-product-detail';
