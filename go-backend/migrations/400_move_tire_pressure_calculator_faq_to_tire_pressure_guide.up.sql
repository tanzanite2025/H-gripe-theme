-- Move the tire-pressure calculation FAQ to the guide page that now owns the
-- calculator. Existing questions, answers, translations, order, and edits are
-- preserved because only the route-owned page ID changes.

UPDATE faqs
SET page_id = 'guides-tireguides-tire-pressure'
WHERE page_id = 'guides-tireguides-tire-pressure-calculator';

UPDATE faq_pages
SET
    page_id = 'guides-tireguides-tire-pressure',
    route_path = '/guides/tireguides/tire-pressure',
    route_key = 'guides-tireguides-tire-pressure',
    manifest_version = '2026-10-06',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides'
WHERE page_id = 'guides-tireguides-tire-pressure-calculator';
