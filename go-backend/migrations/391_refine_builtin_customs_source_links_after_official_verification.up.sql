-- Point the built-in templates at the exact official US HTS headings/lines used as
-- their source baseline. Preserve an administrator's later source-link edit.
UPDATE customs_classification_profiles
SET source_url = 'https://hts.usitc.gov/search?query=8714.92.10.00',
    verified_at = DATE '2026-10-05',
    review_due_at = DATE '2027-10-05',
    updated_at = NOW()
WHERE slug = 'carbon-bicycle-rim'
  AND source = 'built_in'
  AND source_url = 'https://hts.usitc.gov/search?query=8714.92';

UPDATE customs_classification_profiles
SET source_url = 'https://hts.usitc.gov/search?query=8714.92.50.00',
    verified_at = DATE '2026-10-05',
    review_due_at = DATE '2027-10-05',
    updated_at = NOW()
WHERE slug IN ('carbon-bicycle-spoke', 'bicycle-spoke')
  AND source = 'built_in'
  AND source_url = 'https://hts.usitc.gov/search?query=8714.92';

UPDATE customs_classification_profiles
SET source_url = 'https://hts.usitc.gov/search?query=8714.99.80.00',
    verified_at = DATE '2026-10-05',
    review_due_at = DATE '2027-10-05',
    updated_at = NOW()
WHERE slug IN ('bicycle-wheelset', 'carbon-bicycle-handlebar-stem', 'carbon-bicycle-handlebar')
  AND source = 'built_in'
  AND source_url = 'https://hts.usitc.gov/search?query=8714.99';

UPDATE customs_classification_profiles
SET notes = 'US HTS heading 8714.93 covers hubs other than coaster braking hubs and hub brakes; the 10 digit line depends on construction, speed, quick release, and duty conditions (for example 8714.93.05.00, 8714.93.15.00, 8714.93.24.00, 8714.93.28.00, or 8714.93.35.00). UK Trade Tariff CN heading 87149300 currently exposes declarable subitem 8714930090 (https://www.gov.uk/trade-tariff/8714930090), valid from 2012-01-01. Do not treat the 8 digit heading or one national 10 digit line as universal.',
    updated_at = NOW()
WHERE slug = 'bicycle-hub'
  AND source = 'built_in'
  AND source_url = 'https://hts.usitc.gov/search?query=8714.93'
  AND notes LIKE 'US HTS heading 8714.93 covers hubs%';

UPDATE customs_classification_profiles
SET verified_at = DATE '2026-10-05',
    review_due_at = DATE '2027-10-05',
    updated_at = NOW()
WHERE slug = 'bicycle-hub'
  AND source = 'built_in'
  AND source_url = 'https://hts.usitc.gov/search?query=8714.93';
