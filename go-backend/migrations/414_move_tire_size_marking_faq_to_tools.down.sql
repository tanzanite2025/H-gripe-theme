-- Restore the previous guide route only when this migration is explicitly
-- rolled back. No redirect or second public page is registered.
UPDATE faq_pages
SET route_path = '/guides/tireguides/tire-size-markings',
    route_key = 'guides-tireguides-tire-size-markings',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-tire-size-markings'
  AND route_path = '/tools/tire-size-marking-converter'
  AND deleted_at IS NULL;

UPDATE storefront_route_catalog_entries
SET route_key = 'manifest:guides-tireguides-tire-size-markings:' || locale,
    source_key = 'guides-tireguides-tire-size-markings',
    path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/tire-size-markings',
    canonical_path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/guides/tireguides/tire-size-markings',
    summary = 'Learn to read ETRTO, inch, and French tire markings before checking rim fit',
    manifest_version = '2026-10-13',
    updated_at = NOW()
WHERE source_type = 'static'
  AND source_key = 'guides-tireguides-tire-size-markings';
