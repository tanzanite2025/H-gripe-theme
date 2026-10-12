-- Move the stainless-steel spoke mechanics FAQ page to its complete tools route.
-- Keep the page ID and route key stable so existing questions, answers, and
-- automatic-reply references remain attached to the same content identity.

UPDATE faq_pages
SET
    route_path = '/tools/spoke-stress-relief-mechanics',
    route_key = 'guides-spokeguides-stainless-steel-microstructural-dislocation-mechanics',
    manifest_version = '2026-10-13',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-spokeguides-stainless-steel-microstructural-dislocation-mechanics'
  AND deleted_at IS NULL;
