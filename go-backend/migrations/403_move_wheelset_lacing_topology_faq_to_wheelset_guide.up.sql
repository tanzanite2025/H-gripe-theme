-- Move the wheelset lacing FAQ route into its owning Wheelset Guide domain.
-- Keep page_id stable so existing FAQ answers and reply references stay intact.
UPDATE faq_pages
SET route_path = '/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference',
    route_key = 'guides-wheelset-buyers-spoke-lacing-topology',
    manifest_version = '2026-10-06',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'resources-wheelset-lacing-topology'
  AND deleted_at IS NULL;
