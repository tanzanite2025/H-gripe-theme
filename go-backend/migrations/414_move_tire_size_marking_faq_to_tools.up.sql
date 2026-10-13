-- Move the complete tire-size marking reference and mapping tool to its
-- standalone tools route while preserving the existing FAQ identity.
UPDATE faq_pages
SET route_path = '/tools/tire-size-marking-converter',
    route_key = 'guides-tireguides-tire-size-markings',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-tire-size-markings'
  AND route_path IN (
      '/guides/tireguides/tire-size-markings',
      '/tools/tire-size-marking-converter'
  )
  AND deleted_at IS NULL;

UPDATE storefront_route_catalog_entries
SET route_key = 'manifest:guides-tireguides-tire-size-markings:' || locale,
    source_key = 'guides-tireguides-tire-size-markings',
    path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/tools/tire-size-marking-converter',
    canonical_path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/tools/tire-size-marking-converter',
    summary = 'Convert ETRTO, inch, and French tire markings with the complete size reference',
    manifest_version = '2026-10-13',
    updated_at = NOW()
WHERE source_type = 'static'
  AND source_key = 'guides-tireguides-tire-size-markings';
