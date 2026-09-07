-- Rollback restores the legacy schema shape without restoring deleted data.

ALTER TABLE faqs
    ADD COLUMN IF NOT EXISTS category VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_faqs_category ON faqs(category);

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

CREATE INDEX IF NOT EXISTS idx_faq_categories_page_id ON faq_categories (page_id);
CREATE INDEX IF NOT EXISTS idx_faq_categories_locale ON faq_categories (locale);
CREATE INDEX IF NOT EXISTS idx_faq_categories_status ON faq_categories (status);
CREATE INDEX IF NOT EXISTS idx_faq_categories_sort_order ON faq_categories (sort_order);
CREATE INDEX IF NOT EXISTS idx_faq_categories_deleted_at ON faq_categories (deleted_at);

CREATE INDEX IF NOT EXISTS idx_faqs_page_locale_category
    ON faqs (page_id, locale, category);
