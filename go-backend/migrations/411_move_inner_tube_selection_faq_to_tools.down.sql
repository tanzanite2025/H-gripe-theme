-- Restore the previous route metadata while leaving FAQ questions and answers
-- intact. The storefront route manifest controls whether the old URL is live.

UPDATE faq_pages
SET
    route_path = '/guides/tireguides/choose-inner-tube',
    route_key = 'guides-tireguides-choose-inner-tube',
    route_status = 'current',
    domain = 'guides',
    updated_at = NOW()
WHERE page_id = 'guides-tireguides-choose-inner-tube'
  AND deleted_at IS NULL;
