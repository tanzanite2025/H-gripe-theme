-- FAQ categories were removed before launch.
-- Keep the page_id relation as the only FAQ grouping contract.

DROP INDEX IF EXISTS idx_faqs_page_locale_category;
DROP INDEX IF EXISTS idx_faqs_category;

ALTER TABLE faqs
    DROP COLUMN IF EXISTS category;

DROP TABLE IF EXISTS faq_categories;
