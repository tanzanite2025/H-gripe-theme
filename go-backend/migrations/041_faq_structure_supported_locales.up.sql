-- Align FAQ page/category structure with the storefront's supported locale set.
--
-- FAQ item content is not copied to every locale. Missing localized answers
-- should remain empty until they are created in Admin.

UPDATE faq_pages AS source
SET locale = 'zh_cn',
    updated_at = NOW()
WHERE source.locale = 'zh'
  AND source.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM faq_pages AS existing
    WHERE existing.locale = 'zh_cn'
      AND existing.deleted_at IS NULL
      AND (
        existing.page_id = source.page_id
        OR (source.route_path <> '' AND existing.route_path = source.route_path)
      )
  );

UPDATE faq_categories AS source
SET locale = 'zh_cn',
    updated_at = NOW()
WHERE source.locale = 'zh'
  AND source.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM faq_categories AS existing
    WHERE existing.locale = 'zh_cn'
      AND existing.deleted_at IS NULL
      AND existing.page_id = source.page_id
      AND existing.category_key = source.category_key
  );

UPDATE faqs
SET locale = 'zh_cn',
    updated_at = NOW()
WHERE locale = 'zh'
  AND deleted_at IS NULL;

WITH supported_locales(locale) AS (
  VALUES
    ('en'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'), ('pt'),
    ('ru'), ('ar'), ('fi'), ('da'), ('th'), ('sv'), ('id'), ('ms'),
    ('be'), ('tr'), ('bn'), ('fa'), ('nl'), ('hi'), ('ur'), ('mr'),
    ('pcm'), ('fil'), ('te'), ('ha'), ('ps'), ('sw'), ('tl'), ('ta'),
    ('jv'), ('zh_cn')
),
source_pages AS (
  SELECT DISTINCT ON (page_id)
    page_id,
    route_path,
    domain,
    title,
    subtitle,
    sort_order,
    status
  FROM faq_pages
  WHERE locale = 'en'
    AND deleted_at IS NULL
  ORDER BY page_id, sort_order ASC, id ASC
)
INSERT INTO faq_pages (
  page_id,
  route_path,
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
  source_pages.page_id,
  source_pages.route_path,
  source_pages.domain,
  supported_locales.locale,
  source_pages.title,
  source_pages.subtitle,
  source_pages.sort_order,
  source_pages.status,
  NOW(),
  NOW()
FROM source_pages
CROSS JOIN supported_locales
ON CONFLICT DO NOTHING;

WITH supported_locales(locale) AS (
  VALUES
    ('en'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'), ('pt'),
    ('ru'), ('ar'), ('fi'), ('da'), ('th'), ('sv'), ('id'), ('ms'),
    ('be'), ('tr'), ('bn'), ('fa'), ('nl'), ('hi'), ('ur'), ('mr'),
    ('pcm'), ('fil'), ('te'), ('ha'), ('ps'), ('sw'), ('tl'), ('ta'),
    ('jv'), ('zh_cn')
),
source_categories AS (
  SELECT DISTINCT ON (page_id, category_key)
    page_id,
    category_key,
    name,
    icon,
    sort_order,
    status
  FROM faq_categories
  WHERE locale = 'en'
    AND deleted_at IS NULL
  ORDER BY page_id, category_key, sort_order ASC, id ASC
)
INSERT INTO faq_categories (
  page_id,
  category_key,
  name,
  icon,
  locale,
  sort_order,
  status,
  created_at,
  updated_at
)
SELECT
  source_categories.page_id,
  source_categories.category_key,
  source_categories.name,
  source_categories.icon,
  supported_locales.locale,
  source_categories.sort_order,
  source_categories.status,
  NOW(),
  NOW()
FROM source_categories
CROSS JOIN supported_locales
ON CONFLICT DO NOTHING;
