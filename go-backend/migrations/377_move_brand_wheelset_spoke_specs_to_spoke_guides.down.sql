UPDATE faqs
SET page_id = 'resources-brand-wheelset-spoke-specs'
WHERE page_id = 'guides-spokeguides-brand-wheelset-spoke-specs';

UPDATE faq_pages
SET
    page_id = 'resources-brand-wheelset-spoke-specs',
    route_path = '/resources/brand-wheelset-spoke-specs',
    route_key = 'resources-brand-wheelset-spoke-specs',
    domain = 'resources'
WHERE page_id = 'guides-spokeguides-brand-wheelset-spoke-specs';
