-- Move the spoke length calculator into the spoke guides topic while
-- preserving its existing FAQ questions and answers.

UPDATE faqs
SET page_id = 'guides-spokeguides-spoke-length-calculator'
WHERE page_id = 'products-spoke-calculator';

UPDATE faq_pages
SET
    page_id = 'guides-spokeguides-spoke-length-calculator',
    route_path = '/guides/spokeguides/spoke-length-calculator',
    route_key = 'guides-spokeguides-spoke-length-calculator',
    domain = 'guides'
WHERE page_id = 'products-spoke-calculator'
   OR route_key = 'resources-spoke-calculator';
