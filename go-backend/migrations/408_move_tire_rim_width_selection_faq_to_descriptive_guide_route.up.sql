-- Rename the route and FAQ identity once before launch. Keep the numeric page
-- and question IDs, administrator content, publication state, order and images.
-- The storefront exposes only the descriptive route; no legacy alias is created.
UPDATE faq_pages
SET page_id = 'guides-tireguides-tire-rim-width-compatibility',
    route_path = '/guides/tireguides/tire-rim-width-compatibility',
    route_key = 'guides-tireguides-tire-rim-width-compatibility',
    manifest_version = '2026-10-11',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id IN (
    'guides-tireguides-choose',
    'guides-tireguides-tire-rim-width-compatibility'
);

UPDATE faqs
SET page_id = 'guides-tireguides-tire-rim-width-compatibility'
WHERE page_id = 'guides-tireguides-choose';

-- Rename the existing catalog rows in place so the sitemap also has one URL.
UPDATE storefront_route_catalog_entries
SET route_key = 'manifest:guides-tireguides-tire-rim-width-compatibility:' || locale,
    source_key = 'guides-tireguides-tire-rim-width-compatibility',
    path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/tire-rim-width-compatibility',
    canonical_path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/tire-rim-width-compatibility',
    summary = 'Calculate tire and rim width fit, check ISO normal ranges, and compare published compatibility charts',
    manifest_version = '2026-10-11',
    updated_at = NOW()
WHERE source_type = 'static'
  AND source_key = 'guides-tireguides-choose';
