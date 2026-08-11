-- Remove legacy storefront brand text from public content.
-- Keep media URLs and internal resource names unchanged so existing assets
-- continue to resolve during the brand transition.

UPDATE settings
SET value = regexp_replace(value, 'tanzanite', 'H-GRIPE', 'gi'),
    updated_at = NOW()
WHERE is_public = TRUE
  AND "group" IN ('site', 'website_profile')
  AND key NOT ILIKE '%url%'
  AND key NOT ILIKE '%logo%'
  AND value ~* 'tanzanite';

UPDATE faq_pages
SET title = CASE
                WHEN title ~* 'tanzanite' THEN regexp_replace(title, 'tanzanite', 'H-GRIPE', 'gi')
                ELSE title
            END,
    subtitle = CASE
                   WHEN subtitle ~* 'tanzanite' THEN regexp_replace(subtitle, 'tanzanite', 'H-GRIPE', 'gi')
                   ELSE subtitle
               END,
    updated_at = NOW()
WHERE COALESCE(title, '') ~* 'tanzanite'
   OR COALESCE(subtitle, '') ~* 'tanzanite';

UPDATE faq_categories
SET name = CASE
               WHEN name ~* 'tanzanite' THEN regexp_replace(name, 'tanzanite', 'H-GRIPE', 'gi')
               ELSE name
           END,
    updated_at = NOW()
WHERE COALESCE(name, '') ~* 'tanzanite';

UPDATE faqs
SET question = CASE
                   WHEN question ~* 'tanzanite' THEN regexp_replace(question, 'tanzanite', 'H-GRIPE', 'gi')
                   ELSE question
               END,
    answer = CASE
                 WHEN answer ~* 'tanzanite' THEN regexp_replace(answer, 'tanzanite', 'H-GRIPE', 'gi')
                 ELSE answer
             END,
    updated_at = NOW()
WHERE COALESCE(question, '') ~* 'tanzanite'
   OR COALESCE(answer, '') ~* 'tanzanite';
