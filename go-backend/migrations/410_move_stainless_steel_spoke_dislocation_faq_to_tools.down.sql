-- Restore the previous route metadata while leaving FAQ questions and answers
-- intact. The storefront route manifest controls whether the old route is live.

UPDATE faq_pages
SET
    route_path = '/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics',
    route_key = 'guides-spokeguides-stainless-steel-microstructural-dislocation-mechanics',
    route_status = 'current',
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-spokeguides-stainless-steel-microstructural-dislocation-mechanics'
  AND deleted_at IS NULL;
