-- Remove accidental H-GRIPE public/default brand values from existing
-- databases. Production branding should come from editable settings, while
-- Docker resource names and GHCR image names remain operational identifiers.

UPDATE settings
SET value = 'Store',
    updated_at = NOW()
WHERE "group" = 'site'
  AND key IN ('site_name', 'brand_title')
  AND value IN ('H-GRIPE', 'H GRIPE', 'Tanzanite', 'TANZANITE');

UPDATE settings
SET value = 'Premium Cycling Components',
    updated_at = NOW()
WHERE "group" = 'seo'
  AND key IN ('home_meta_title', 'article_meta_title', 'product_meta_title')
  AND value IN ('H-GRIPE - Premium E-commerce', 'Tanzanite - Premium E-commerce');

UPDATE settings
SET value = 'Shop premium cycling components',
    updated_at = NOW()
WHERE "group" = 'seo'
  AND key IN ('home_meta_description', 'article_meta_description', 'product_meta_description')
  AND value IN ('Shop premium products at H-GRIPE', 'Shop premium products at Tanzanite');

UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE "group" = 'site'
  AND key = 'contact_email'
  AND value IN ('contact@tanzanite.site', 'contact@tanzanite.com');

UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE "group" = 'site'
  AND key = 'contact_phone'
  AND value = '+1-234-567-8900';

UPDATE settings
SET value = 'noreply@example.com',
    updated_at = NOW()
WHERE "group" = 'email'
  AND key = 'from_email'
  AND value IN ('noreply@tanzanite.site', 'noreply@tanzanite.com');

UPDATE settings
SET value = 'Store Support',
    updated_at = NOW()
WHERE "group" = 'email'
  AND key = 'from_name'
  AND value IN ('H-GRIPE', 'H GRIPE', 'Tanzanite', 'TANZANITE');

UPDATE faq_pages
SET subtitle = 'Learn more about our story and mission',
    updated_at = NOW()
WHERE subtitle = 'Learn more about H-GRIPE and our mission';

UPDATE faq_pages
SET subtitle = 'Common questions about articles and buying guides',
    updated_at = NOW()
WHERE subtitle = 'Common questions about H-GRIPE articles and buying guides';

UPDATE faq_pages
SET subtitle = 'Common questions about news updates',
    updated_at = NOW()
WHERE subtitle = 'Common questions about H-GRIPE news updates';

UPDATE faq_pages
SET subtitle = 'Common questions about our factory and product philosophy',
    updated_at = NOW()
WHERE subtitle = 'Common questions about H-GRIPE, our factory, and our product philosophy';

UPDATE faq_pages
SET subtitle = 'Common questions about store policies',
    updated_at = NOW()
WHERE subtitle = 'Common questions about H-GRIPE policies';

UPDATE faq_categories
SET name = 'About our company',
    updated_at = NOW()
WHERE name = 'About H-GRIPE';

UPDATE faqs
SET question = 'Are all of your wheels UCI approved?',
    updated_at = NOW()
WHERE question = 'Are all H-GRIPE wheels UCI approved?';

UPDATE faqs
SET question = 'Where is your team based?',
    answer = 'Our Global Headquarters are in Hong Kong, and our Manufacturing & R&D Base is in Xiamen, China.',
    updated_at = NOW()
WHERE question = 'Where is H-GRIPE based?';

UPDATE faqs
SET question = 'What is your core mission?',
    updated_at = NOW()
WHERE question = 'What is H-GRIPE’s core mission?';

UPDATE faqs
SET question = 'How do you approach sustainability?',
    updated_at = NOW()
WHERE question = 'How does H-GRIPE approach sustainability?';

UPDATE faqs
SET answer = REPLACE(answer, 'The product is a genuine H-GRIPE product', 'The product is a genuine product from this store'),
    updated_at = NOW()
WHERE answer LIKE '%The product is a genuine H-GRIPE product%';

UPDATE faqs
SET question = 'What is the warranty period for your wheels?',
    answer = REPLACE(answer, 'all H-GRIPE series Wheels/Rims', 'all eligible Wheels/Rims'),
    updated_at = NOW()
WHERE question = 'What is the warranty period for H-GRIPE wheels?';

UPDATE faqs
SET answer = REPLACE(answer, 'H-GRIPE covers shipping costs.', 'we cover shipping costs.'),
    updated_at = NOW()
WHERE answer LIKE '%H-GRIPE covers shipping costs.%';

UPDATE spoke_build_presets
SET name = 'AR 45 Disc + DT Swiss 350',
    updated_at = NOW()
WHERE name = 'H-GRIPE AR 45 Disc + DT Swiss 350';

UPDATE spoke_build_presets
SET name = 'AR 50 Disc + DT Swiss 240 EXP',
    updated_at = NOW()
WHERE name = 'H-GRIPE AR 50 Disc + DT Swiss 240 EXP';
