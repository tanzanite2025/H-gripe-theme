-- P0: repair the historical FAQ route snapshot and add the metadata needed
-- for non-destructive reconciliation from Nuxt's storefront route manifest.
-- FAQ answers and page IDs are deliberately preserved.

ALTER TABLE faq_pages ADD COLUMN IF NOT EXISTS route_key VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE faq_pages ADD COLUMN IF NOT EXISTS manifest_version VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE faq_pages ADD COLUMN IF NOT EXISTS route_status VARCHAR(20) NOT NULL DEFAULT 'stale';
ALTER TABLE faq_pages ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_faq_pages_route_key ON faq_pages (route_key);
CREATE INDEX IF NOT EXISTS idx_faq_pages_manifest_version ON faq_pages (manifest_version);
CREATE INDEX IF NOT EXISTS idx_faq_pages_route_status ON faq_pages (route_status);
CREATE INDEX IF NOT EXISTS idx_faq_pages_last_seen_at ON faq_pages (last_seen_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_faq_pages_route_key_locale_live
  ON faq_pages (route_key, locale)
  WHERE deleted_at IS NULL AND route_key <> '';

-- Repair routes that were moved in the Nuxt storefront while migration 031
-- remained applied. Keep page_id and all FAQ rows stable for deep links and
-- automatic-reply references.
UPDATE faq_pages
SET route_path = '/resources/spoke-calculator',
    route_key = 'resources-spoke-calculator',
    route_status = 'current',
    manifest_version = '2026-09-30',
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE page_id = 'products-spoke-calculator';

UPDATE faq_pages
SET route_path = '/resources/membershipandpoints',
    route_key = 'resources-membershipandpoints',
    route_status = 'current',
    manifest_version = '2026-09-30',
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE page_id = 'company-membership';

UPDATE faq_pages
SET route_path = '/resources/blog',
    route_key = 'resources-blog',
    route_status = 'current',
    manifest_version = '2026-09-30',
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE page_id = 'blog';

UPDATE faq_pages
SET route_path = '/resources/picture-warehouse',
    route_key = 'resources-picture-warehouse',
    route_status = 'current',
    manifest_version = '2026-09-30',
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE page_id = 'picture-warehouse';

UPDATE faq_pages
SET route_path = '/policies/refund-cancellation',
    route_key = 'policies-refund-cancellation',
    route_status = 'current',
    manifest_version = '2026-09-30',
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE page_id = 'policies-refund-cancellation'
   OR route_path = '/policies/refund-return';

-- These pages no longer have a canonical declaration in the current Nuxt
-- manifest. Keep their content for audit/deep-link safety, but expose the
-- stale state so Admin can decide whether to migrate or hide them.
UPDATE faq_pages
SET route_status = 'stale',
    updated_at = NOW()
WHERE page_id IN ('blog-news', 'blog-wheelsbuild', 'support-product-feedback', 'shop-product-detail')
   OR route_path IN ('/blog/news', '/blog/wheelsbuild', '/membershipandpoints', '/spoke-calculator');

WITH manifest_routes(route_key, route_path, label, description, is_alias, sort_order) AS (
    VALUES
        ('home', '/', 'Home', 'Storefront home page', FALSE, 10),
        ('shop', '/shop', 'Shop', 'Product catalog', FALSE, 20),
        ('resources-blog', '/resources/blog', 'Blog', 'Blog landing page', FALSE, 30),
        ('company-about', '/company/about', 'About Us', 'Company and product philosophy', FALSE, 40),
        ('company-certificates', '/company/certificates', 'Certificates', 'Company certificates', FALSE, 50),
        ('company-contact', '/company/contact', 'Contact', 'Contact information', FALSE, 60),
        ('company-global-partners', '/company/global-partners', 'Global Partners', 'Global partner network', FALSE, 70),
        ('company-oem-odm', '/company/oem-odm', 'OEM / ODM', 'OEM and ODM services', FALSE, 80),
        ('company-ourstory', '/company/ourstory', 'Our Story', 'Company history', FALSE, 90),
        ('website', '/website', 'Website', 'Information about this website', FALSE, 100),
        ('website-me-this-website', '/website/me-this-website', 'Me & This Website', 'The person and work behind this website', FALSE, 110),
        ('website-why-this-name', '/website/why-this-name', 'Why This Name', 'Why this website uses its name', FALSE, 120),
        ('guides-tireguides', '/guides/tireguides', 'Tire Guides', 'Tire guides', FALSE, 130),
        ('guides-tireguides-tire-frame-clearance', '/guides/tireguides/tire-frame-clearance', 'Tire and frame clearance guide', 'Compare tire dimensions with bicycle frame, fork, and stay clearance', FALSE, 140),
        ('guides-tireguides-schwalbe-tire-circumference', '/guides/tireguides/schwalbe-tire-circumference', 'Schwalbe tire circumference guide', 'Reference circumference values for Schwalbe bicycle tire sizes', FALSE, 150),
        ('guides-tireguides-schwalbe-tire-selector', '/guides/tireguides/schwalbe-tire-selector', 'Schwalbe tire selector', 'Search the sourced Schwalbe tire catalog by model and size', FALSE, 160),
        ('guides-wheelset-buyers', '/guides/wheelset-buyers', 'Wheelset Buyers Guide', 'Wheelset buying guide', FALSE, 170),
        ('resources-membershipandpoints', '/resources/membershipandpoints', 'Membership and Points', 'Membership and points', FALSE, 180),
        ('resources-picture-warehouse', '/resources/picture-warehouse', 'Picture Warehouse', 'Product image reference gallery', FALSE, 190),
        ('policies-cookie', '/policies/cookie', 'Cookie Policy', 'Cookie policy', FALSE, 200),
        ('policies-privacy', '/policies/privacy', 'Privacy Policy', 'Privacy policy', FALSE, 210),
        ('policies-refund-cancellation', '/policies/refund-cancellation', 'Refund & Cancellation Policy', 'Refund & Cancellation Policy', FALSE, 220),
        ('policies-terms', '/policies/terms', 'Terms of Service', 'Terms of service', FALSE, 230),
        ('resources-spoke-calculator', '/resources/spoke-calculator', 'Spoke Calculator', 'Spoke length calculator', FALSE, 240),
        ('resources-brand-wheelset-spoke-specs', '/resources/brand-wheelset-spoke-specs', 'Complete Wheelset Spoke Specs', 'Market-available branded wheelset spoke lengths, types, and lacing references', FALSE, 250),
        ('support-faqs', '/support/faqs', 'FAQs', 'All frequently asked questions', FALSE, 260),
        ('support-payment', '/support/payment', 'Payment Support', 'Payment support', FALSE, 270),
        ('support-shipping', '/support/shipping', 'Shipping Support', 'Shipping support', FALSE, 280),
        ('support-test-report', '/support/test-report', 'Test Report', 'Product test reports', FALSE, 290),
        ('support-warranty-check', '/support/warranty-check', 'Warranty Check', 'Warranty verification', FALSE, 300),
        ('support-warranty', '/support/warranty', 'Warranty', 'Warranty information', FALSE, 310),
        ('legacy-faq', '/faq', 'FAQ (Legacy Alias)', 'Legacy FAQ route redirected to the FAQ index', TRUE, 320),
        ('products-product-detail', '/products/:slug', 'Product Detail', 'Common questions shown on individual product detail pages', FALSE, 330)
), supported_locales(locale) AS (
    VALUES
        ('en'), ('zh_cn'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'),
        ('pt'), ('ru'), ('ar'), ('nl'), ('tr'), ('id'), ('th'), ('sv'),
        ('da'), ('fi'), ('hi'), ('ms')
)
INSERT INTO faq_pages (
    page_id,
    route_path,
    route_key,
    manifest_version,
    route_status,
    last_seen_at,
    domain,
    locale,
    title,
    subtitle,
    sort_order,
    status,
    created_at,
    updated_at
)
SELECT
    COALESCE(
        (
            SELECT existing.page_id
            FROM faq_pages AS existing
            WHERE existing.route_path = manifest_routes.route_path
              AND existing.locale = supported_locales.locale
              AND existing.deleted_at IS NULL
            ORDER BY existing.id ASC
            LIMIT 1
        ),
        CASE manifest_routes.route_key
            WHEN 'resources-spoke-calculator' THEN 'products-spoke-calculator'
            WHEN 'resources-membershipandpoints' THEN 'company-membership'
            WHEN 'resources-blog' THEN 'blog'
            WHEN 'resources-picture-warehouse' THEN 'picture-warehouse'
            WHEN 'guides-tireguides-schwalbe-tire-selector' THEN 'guides-schwalbe-tire-selector'
            WHEN 'legacy-faq' THEN 'faq'
            ELSE manifest_routes.route_key
        END
    ),
    manifest_routes.route_path,
    manifest_routes.route_key,
    '2026-09-30',
    CASE WHEN manifest_routes.is_alias THEN 'alias' ELSE 'current' END,
    NOW(),
    CASE
        WHEN split_part(trim(both '/' from manifest_routes.route_path), '/', 1) = 'resources' THEN 'resources'
        WHEN split_part(trim(both '/' from manifest_routes.route_path), '/', 1) = '' THEN 'general'
        ELSE split_part(trim(both '/' from manifest_routes.route_path), '/', 1)
    END,
    supported_locales.locale,
    manifest_routes.label || CASE WHEN manifest_routes.is_alias THEN ' FAQs (Alias)' ELSE ' FAQs' END,
    manifest_routes.description,
    manifest_routes.sort_order,
    'active',
    NOW(),
    NOW()
FROM manifest_routes
CROSS JOIN supported_locales
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    route_key = EXCLUDED.route_key,
    manifest_version = EXCLUDED.manifest_version,
    route_status = EXCLUDED.route_status,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at = NOW(),
    deleted_at = NULL;
