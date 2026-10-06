-- Restore the calculator FAQ route if migration 400 is rolled back.

UPDATE faqs
SET page_id = 'guides-tireguides-tire-pressure-calculator'
WHERE page_id = 'guides-tireguides-tire-pressure';

UPDATE faq_pages
SET
    page_id = 'guides-tireguides-tire-pressure-calculator',
    route_path = '/guides/tireguides/tire-pressure-calculator',
    route_key = 'guides-tireguides-tire-pressure-calculator',
    manifest_version = '2026-10-06',
    route_status = 'current',
    last_seen_at = NOW(),
    domain = 'guides'
WHERE page_id = 'guides-tireguides-tire-pressure';
