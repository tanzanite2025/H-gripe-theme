-- Move the brand wheelset spoke specification page into the spoke guides
-- topic while preserving its existing FAQ content and editorial ordering.

UPDATE faqs
SET page_id = 'guides-spokeguides-brand-wheelset-spoke-specs'
WHERE page_id = 'resources-brand-wheelset-spoke-specs';

UPDATE faq_pages
SET
    page_id = 'guides-spokeguides-brand-wheelset-spoke-specs',
    route_path = '/guides/spokeguides/brand-wheelset-spoke-specs',
    route_key = 'guides-spokeguides-brand-wheelset-spoke-specs',
    domain = 'guides'
WHERE page_id = 'resources-brand-wheelset-spoke-specs';
