-- Restore the former virtual-tab FAQ metadata without changing FAQ answers.
UPDATE faq_pages
SET route_path = '/guides/wheelset-buyers/choose-freehub',
    route_key = 'guides-wheelset-buyers-choose-freehub',
    manifest_version = '2026-10-12',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE route_path = '/tools/freehub-cassette-fitment-engine'
  AND route_key = 'guides-wheelset-buyers-choose-freehub'
  AND deleted_at IS NULL;
