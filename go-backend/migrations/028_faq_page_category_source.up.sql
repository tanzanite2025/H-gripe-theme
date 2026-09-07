-- Create a database-owned FAQ page structure.
-- FAQ content stays in faqs; page metadata lives here so the admin panel can
-- manage the same page sections rendered by the Nuxt storefront.

ALTER TABLE faqs ADD COLUMN IF NOT EXISTS page_id VARCHAR(120) DEFAULT '';
ALTER TABLE faqs ALTER COLUMN page_id SET DEFAULT '';

CREATE TABLE IF NOT EXISTS faq_pages (
    id BIGSERIAL PRIMARY KEY,
    page_id VARCHAR(120) NOT NULL,
    route_path VARCHAR(255) NOT NULL DEFAULT '',
    domain VARCHAR(80) NOT NULL DEFAULT '',
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    title VARCHAR(255) NOT NULL DEFAULT '',
    subtitle TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT faq_pages_page_locale_key UNIQUE (page_id, locale)
);

CREATE INDEX IF NOT EXISTS idx_faq_pages_domain ON faq_pages (domain);
CREATE INDEX IF NOT EXISTS idx_faq_pages_locale ON faq_pages (locale);
CREATE INDEX IF NOT EXISTS idx_faq_pages_status ON faq_pages (status);
CREATE INDEX IF NOT EXISTS idx_faq_pages_sort_order ON faq_pages (sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_pages_deleted_at ON faq_pages (deleted_at);

CREATE INDEX IF NOT EXISTS idx_faqs_page_id ON faqs (page_id);
CREATE INDEX IF NOT EXISTS idx_faqs_page_locale ON faqs (page_id, locale);

WITH seed_pages(page_id, route_path, domain, title, subtitle, sort_order) AS (
    VALUES
        ('products-spoke-calculator', '/spoke-calculator', 'products', 'Spoke Calculator FAQs', 'Common questions about using our calculator and obtaining accurate measurements', 100),
        ('guides-tireguides', '/guides/tireguides', 'guides', 'Tire Guides FAQs', 'Key questions about tire sizing, pressure, and inner tubes', 200),
        ('guides-wheelset-buyers', '/guides/wheelset-buyers', 'guides', 'Wheelset Buyers Guide FAQs', 'Common questions about choosing and customizing wheelsets', 210),
        ('support-payment', '/support/payment', 'support', 'Payment FAQs', 'Common questions about payment methods, fees, and troubleshooting', 300),
        ('support-shipping', '/support/shipping', 'support', 'Shipping FAQs', 'Common questions about shipping, delivery times, and tracking', 310),
        ('support-warranty', '/support/warranty', 'support', 'Warranty FAQs', 'Common questions about warranty coverage and claims', 320),
        ('support-warranty-check', '/support/warranty-check', 'support', 'Warranty Check FAQs', 'Common questions about checking your product warranty', 330),
        ('support-product-feedback', '/support/product-feedback', 'support', 'Product Feedback FAQs', 'Common questions about sharing feedback and suggestions', 340),
        ('support-test-report', '/support/test-report', 'support', 'Test Report FAQs', 'Common questions about product testing and quality assurance', 350),
        ('company-membership', '/membershipandpoints', 'company', 'Membership & Points FAQs', 'Common questions about membership tiers, benefits, and points', 400),
        ('company-oem-odm', '/company/oem-odm', 'company', 'OEM/ODM Services FAQ', 'Common questions about our manufacturing services', 410),
        ('company-certificates', '/company/certificates', 'company', 'Certificates & Testing FAQ', 'Common questions about our quality standards and certifications', 420),
        ('company-contact', '/company/contact', 'company', 'Contact & Support FAQ', 'Common questions about reaching our team', 430),
        ('company-global-partners', '/company/global-partners', 'company', 'Global Partnerships FAQ', 'Common questions about becoming a global partner', 440),
        ('company-ourstory', '/company/ourstory', 'company', 'Our Story & Brand FAQ', 'Learn more about our story and mission', 450)
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
