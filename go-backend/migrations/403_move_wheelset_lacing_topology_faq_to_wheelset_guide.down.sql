-- Restore the previous FAQ route metadata without changing FAQ answers.
UPDATE faq_pages
SET route_path = '/resources/wheelset-spoke-lacing-topology-and-geometry-reference',
    route_key = 'resources-wheelset-lacing-topology',
    manifest_version = '2026-10-01',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'resources',
    updated_at = NOW()
WHERE page_id = 'resources-wheelset-lacing-topology'
  AND deleted_at IS NULL;
