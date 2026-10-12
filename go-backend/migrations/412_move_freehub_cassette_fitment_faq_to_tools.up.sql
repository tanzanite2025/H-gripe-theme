-- Move any existing freehub virtual-tab FAQ metadata to the complete tool page.
-- The virtual tab never had a separate manifest entry, but this keeps an
-- editorial page created from an older route catalog from losing its content.
UPDATE faq_pages
SET route_path = '/tools/freehub-cassette-fitment-engine',
    route_key = 'guides-wheelset-buyers-choose-freehub',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE route_path = '/guides/wheelset-buyers/choose-freehub'
  AND deleted_at IS NULL;
