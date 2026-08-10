-- SEO is a domain-owned setting only for the fixed storefront home route.
-- Article and product SEO lives on the corresponding content resources.

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT 'home_meta_title', value, type, locale, 'seo', is_public, 'Home meta title', created_at, NOW()
FROM settings
WHERE "group" = 'seo' AND key = 'meta_title'
ON CONFLICT (key, locale) DO NOTHING;

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
SELECT 'home_meta_description', value, type, locale, 'seo', is_public, 'Home meta description', created_at, NOW()
FROM settings
WHERE "group" = 'seo' AND key = 'meta_description'
ON CONFLICT (key, locale) DO NOTHING;

INSERT INTO settings (key, value, type, locale, "group", is_public, description, created_at, updated_at)
VALUES
    ('home_meta_title', '', 'string', 'en', 'seo', true, 'Home meta title', NOW(), NOW()),
    ('home_meta_description', '', 'string', 'en', 'seo', true, 'Home meta description', NOW(), NOW())
ON CONFLICT (key, locale) DO NOTHING;

DELETE FROM settings
WHERE "group" = 'seo'
  AND key IN (
    'meta_title',
    'meta_description',
    'meta_keywords',
    'home_meta_keywords',
    'article_meta_title',
    'article_meta_description',
    'article_meta_keywords',
    'product_meta_title',
    'product_meta_description',
    'product_meta_keywords'
  );
