-- Move the complete tire/rim width workbench to its standalone tool route.
-- Keep the existing page identity and editorial FAQ content; no compatibility URL is created.
UPDATE faq_pages
SET route_path = '/tools/tire-rim-width-compatibility-engine',
    route_key = 'guides-tireguides-tire-rim-width-compatibility',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-tire-rim-width-compatibility'
  AND route_path IN (
      '/guides/tireguides/tire-rim-width-compatibility',
      '/tools/tire-rim-width-compatibility-engine'
  )
  AND deleted_at IS NULL;

UPDATE storefront_route_catalog_entries
SET route_key = 'manifest:guides-tireguides-tire-rim-width-compatibility:' || locale,
    source_key = 'guides-tireguides-tire-rim-width-compatibility',
    path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/tools/tire-rim-width-compatibility-engine',
    canonical_path = CASE WHEN locale = 'en' THEN '' ELSE '/' || locale END
        || '/tools/tire-rim-width-compatibility-engine',
    summary = 'Calculate tire and rim width fit, inspect simulations, and compare published compatibility charts',
    manifest_version = '2026-10-13',
    updated_at = NOW()
WHERE source_type = 'static'
  AND source_key = 'guides-tireguides-tire-rim-width-compatibility';
