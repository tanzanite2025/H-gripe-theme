-- Restore database identity only when explicitly rolling this migration back.
-- This file does not register a storefront redirect or a second page.
UPDATE faq_pages
SET page_id = 'guides-tireguides-choose',
    route_path = '/guides/tireguides/choose',
    route_key = 'guides-tireguides-choose',
    manifest_version = '2026-10-11',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility';

UPDATE storefront_route_catalog_entries
SET route_key = 'manifest:guides-tireguides-choose:' || locale,
    source_key = 'guides-tireguides-choose',
    path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/choose',
    canonical_path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/choose',
    summary = 'Convert inch tire widths and resolve DT Swiss rim internal width references',
    manifest_version = '2026-10-11',
    updated_at = NOW()
WHERE source_type = 'static'
  AND source_key = 'guides-tireguides-tire-rim-width-compatibility';

UPDATE faqs
SET page_id = 'guides-tireguides-choose'
WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility';
