UPDATE faqs
SET page_id = 'products-spoke-calculator'
WHERE page_id = 'guides-spokeguides-spoke-length-calculator';

UPDATE faq_pages
SET
    page_id = 'products-spoke-calculator',
    route_path = '/resources/spoke-calculator',
    route_key = 'resources-spoke-calculator',
    domain = 'resources'
WHERE page_id = 'guides-spokeguides-spoke-length-calculator';
