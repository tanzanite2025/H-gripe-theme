-- Move the complete inner-tube selection page to its standalone tools route.
-- Keep the page ID and route key stable so existing questions, answers, and
-- administrator metadata remain attached to the same content identity.

UPDATE faq_pages
SET
    route_path = '/tools/valve-length-fitment-calculator',
    route_key = 'guides-tireguides-choose-inner-tube',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-choose-inner-tube'
  AND deleted_at IS NULL;
