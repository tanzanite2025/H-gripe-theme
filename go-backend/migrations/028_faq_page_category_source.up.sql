-- Create a database-owned FAQ page/category structure.
-- FAQ content stays in faqs; page/category metadata lives here so the admin
-- panel can manage the same page sections rendered by the Nuxt storefront.

ALTER TABLE faqs ADD COLUMN IF NOT EXISTS page_id VARCHAR(120) DEFAULT '';
ALTER TABLE faqs ALTER COLUMN page_id SET DEFAULT '';
ALTER TABLE faqs ALTER COLUMN category TYPE VARCHAR(120);

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

CREATE TABLE IF NOT EXISTS faq_categories (
    id BIGSERIAL PRIMARY KEY,
    page_id VARCHAR(120) NOT NULL,
    category_key VARCHAR(120) NOT NULL,
    name VARCHAR(180) NOT NULL DEFAULT '',
    icon VARCHAR(40) NOT NULL DEFAULT '',
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT faq_categories_page_key_locale_key UNIQUE (page_id, category_key, locale)
);

CREATE INDEX IF NOT EXISTS idx_faq_pages_domain ON faq_pages (domain);
CREATE INDEX IF NOT EXISTS idx_faq_pages_locale ON faq_pages (locale);
CREATE INDEX IF NOT EXISTS idx_faq_pages_status ON faq_pages (status);
CREATE INDEX IF NOT EXISTS idx_faq_pages_sort_order ON faq_pages (sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_pages_deleted_at ON faq_pages (deleted_at);

CREATE INDEX IF NOT EXISTS idx_faq_categories_page_id ON faq_categories (page_id);
CREATE INDEX IF NOT EXISTS idx_faq_categories_locale ON faq_categories (locale);
CREATE INDEX IF NOT EXISTS idx_faq_categories_status ON faq_categories (status);
CREATE INDEX IF NOT EXISTS idx_faq_categories_sort_order ON faq_categories (sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_categories_deleted_at ON faq_categories (deleted_at);

CREATE INDEX IF NOT EXISTS idx_faqs_page_id ON faqs (page_id);
CREATE INDEX IF NOT EXISTS idx_faqs_page_locale_category ON faqs (page_id, locale, category);

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
        ('company-ourstory', '/company/ourstory', 'company', 'Our Story & Brand FAQ', 'Learn more about Tanzanite and our mission', 450)
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
        ('products-spoke-calculator', 'usage', 'Calculator Usage', '🧮', 10),
        ('products-spoke-calculator', 'troubleshooting', 'Troubleshooting', '🔧', 20),
        ('guides-tireguides', 'sizing', 'Sizing & compatibility', '📏', 10),
        ('guides-tireguides', 'installation', 'Installation & Setup', '🔧', 20),
        ('guides-tireguides', 'maintenance', 'Maintenance', '🛠️', 30),
        ('guides-wheelset-buyers', 'choosing', 'Choosing Wheelsets', '🎯', 10),
        ('guides-wheelset-buyers', 'customization', 'Customization', '🎨', 20),
        ('guides-wheelset-buyers', 'specs', 'Specifications', '📐', 30),
        ('support-payment', 'payment-methods', 'Payment Methods', '💳', 10),
        ('support-payment', 'security', 'Security & Privacy', '🔒', 20),
        ('support-payment', 'billing', 'Billing & Charges', '🧾', 30),
        ('support-payment', 'troubleshooting', 'Troubleshooting', '🔧', 40),
        ('support-shipping', 'delivery', 'Delivery & Timing', '📦', 10),
        ('support-shipping', 'tracking', 'Tracking & Updates', '🔍', 20),
        ('support-shipping', 'international', 'International Shipping', '🌍', 30),
        ('support-shipping', 'issues', 'Shipping Issues', '⚠️', 40),
        ('support-warranty', 'policy', 'Warranty Policy', '🛡️', 10),
        ('support-warranty', 'crash', 'Accidental Damage', '💥', 20),
        ('support-warranty', 'claims', 'Claims & Process', '📋', 30),
        ('support-warranty-check', 'how-to-check', 'How to Check Warranty', '🔍', 10),
        ('support-warranty-check', 'troubleshooting', 'Troubleshooting', '🔧', 20),
        ('support-product-feedback', 'feedback', 'Submitting Feedback', '💬', 10),
        ('support-product-feedback', 'reviews', 'Product Reviews', '⭐', 20),
        ('support-test-report', 'testing', 'Product Testing', '🔬', 10),
        ('support-test-report', 'reports', 'Accessing Reports', '📄', 20),
        ('support-test-report', 'quality', 'Quality Assurance', '✅', 30),
        ('support-test-report', 'assembly', 'Wheel Assembly', '🔧', 40),
        ('company-membership', 'membership', 'Membership Tiers', '🏆', 10),
        ('company-membership', 'points', 'Points System', '💎', 20),
        ('company-membership', 'benefits', 'Member Benefits', '🎁', 30),
        ('company-oem-odm', 'general', 'General Inquiries', '', 10),
        ('company-certificates', 'general', 'General Inquiries', '', 10),
        ('company-contact', 'general', 'General Support', '', 10),
        ('company-global-partners', 'partnership', 'Partnership Program', '', 10),
        ('company-ourstory', 'brand', 'Brand & Mission', '', 10),
        ('company-ourstory', 'products', 'Product Philosophy', '', 20)
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
