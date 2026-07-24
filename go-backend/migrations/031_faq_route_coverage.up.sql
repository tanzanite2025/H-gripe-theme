-- Keep the backend FAQ page/category structure aligned with Nuxt routes that
-- use the automatic PageFaqSlot container.
--
-- The storefront should not emit 404 noise just because a page has no FAQ
-- items yet. These records make the admin panel the visible source of truth
-- for current page slots, while actual answers remain empty until edited.

CREATE UNIQUE INDEX IF NOT EXISTS idx_faq_pages_route_locale_live
  ON faq_pages (route_path, locale)
  WHERE deleted_at IS NULL AND route_path <> '';

WITH seed_pages(page_id, route_path, domain, title, subtitle, sort_order) AS (
    VALUES
        ('shop', '/shop', 'products', 'Shop FAQs', 'Common questions about browsing products, filtering categories, and finding the right parts', 110),
        ('shop-product-detail', '/shop/:slug', 'products', 'Product Detail FAQs', 'Common questions shown on individual product detail pages', 120),
        ('products-product-detail', '/products/:slug', 'products', 'Legacy Product Detail FAQs', 'Common questions for legacy product detail routes that resolve to product pages', 125),
        ('blog', '/blog', 'blog', 'Blog FAQs', 'Common questions about Tanzanite articles and buying guides', 500),
        ('blog-news', '/blog/news', 'blog', 'News FAQs', 'Common questions about Tanzanite news updates', 510),
        ('blog-wheelsbuild', '/blog/wheelsbuild', 'blog', 'Wheel Build FAQs', 'Common questions about wheel build articles and references', 520),
        ('company-about', '/company/about', 'company', 'About Us FAQs', 'Common questions about Tanzanite, our factory, and our product philosophy', 405),
        ('policies', '/policies', 'policies', 'Policy FAQs', 'Common questions about Tanzanite policies', 600),
        ('policies-cookie', '/policies/cookie', 'policies', 'Cookie Policy FAQs', 'Common questions about cookies and browser data', 610),
        ('policies-privacy', '/policies/privacy', 'policies', 'Privacy Policy FAQs', 'Common questions about privacy and personal data', 620),
        ('policies-refund-return', '/policies/refund-return', 'policies', 'Refund & Return Policy FAQs', 'Common questions about refunds, returns, and after-sales handling', 630),
        ('policies-terms', '/policies/terms', 'policies', 'Terms of Service FAQs', 'Common questions about website terms and purchase conditions', 640),
        ('picture-warehouse', '/picture-warehouse', 'products', 'Picture Warehouse FAQs', 'Common questions about browsing and using gallery reference images', 700),
        ('support-faqs', '/support/faqs', 'support', 'All FAQs Page', 'FAQ aggregation page metadata; this page does not auto-insert another FAQ block', 360),
        ('faq', '/faq', 'support', 'FAQ Landing Page', 'Legacy FAQ landing page metadata; this page does not auto-insert another FAQ block', 365)
),
locales(locale) AS (
    VALUES ('en'), ('zh')
)
INSERT INTO faq_pages (page_id, route_path, domain, locale, title, subtitle, sort_order, status)
SELECT seed_pages.page_id, seed_pages.route_path, seed_pages.domain, locales.locale,
       seed_pages.title, seed_pages.subtitle, seed_pages.sort_order, 'active'
FROM seed_pages
CROSS JOIN locales
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    domain = EXCLUDED.domain,
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW(),
    deleted_at = NULL;

WITH seed_categories(page_id, category_key, name, icon, sort_order) AS (
    VALUES
        ('shop', 'general', 'General Shopping', '', 10),
        ('shop-product-detail', 'general', 'Product Detail', '', 10),
        ('products-product-detail', 'general', 'Product Detail', '', 10),
        ('blog', 'general', 'General Articles', '', 10),
        ('blog-news', 'general', 'News Updates', '', 10),
        ('blog-wheelsbuild', 'general', 'Wheel Build Articles', '', 10),
        ('company-about', 'general', 'About Tanzanite', '', 10),
        ('policies', 'general', 'General Policies', '', 10),
        ('policies-cookie', 'general', 'Cookie Policy', '', 10),
        ('policies-privacy', 'general', 'Privacy Policy', '', 10),
        ('policies-refund-return', 'general', 'Refunds & Returns', '', 10),
        ('policies-terms', 'general', 'Terms of Service', '', 10),
        ('picture-warehouse', 'general', 'Picture Warehouse', '', 10),
        ('support-faqs', 'general', 'FAQ Index', '', 10),
        ('faq', 'general', 'FAQ Landing', '', 10)
),
locales(locale) AS (
    VALUES ('en'), ('zh')
)
INSERT INTO faq_categories (page_id, category_key, name, icon, locale, sort_order, status)
SELECT seed_categories.page_id, seed_categories.category_key, seed_categories.name,
       seed_categories.icon, locales.locale, seed_categories.sort_order, 'active'
FROM seed_categories
CROSS JOIN locales
ON CONFLICT (page_id, category_key, locale) DO UPDATE
SET name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW(),
    deleted_at = NULL;
