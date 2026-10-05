-- Keep the legacy source_url field for API compatibility while recording the
-- official destination authority links independently for each customs market.
ALTER TABLE customs_classification_profiles
    ADD COLUMN IF NOT EXISTS source_url_us TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_url_eu TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_url_uk TEXT NOT NULL DEFAULT '';

-- Existing records used source_url as the primary official source. Preserve
-- that information as the US source until an administrator supplies a more
-- specific destination link.
UPDATE customs_classification_profiles
SET source_url_us = source_url
WHERE BTRIM(source_url_us) = ''
  AND BTRIM(source_url) <> '';

-- EU TARIC is intentionally stored at its official consultation entry point;
-- the applicable CN/TARIC national measure still depends on destination,
-- origin, date, and product facts.
UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714921000'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug = 'carbon-bicycle-rim'
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');

-- Keep the verification window visible in the destination-source matrix. Do
-- not replace a later administrator verification with this baseline date.
UPDATE customs_classification_profiles
SET verified_at = COALESCE(verified_at, DATE '2026-10-05'),
    review_due_at = COALESCE(review_due_at, DATE '2027-10-05')
WHERE source = 'built_in'
  AND slug IN (
      'carbon-bicycle-rim',
      'carbon-bicycle-spoke',
      'bicycle-spoke',
      'bicycle-hub',
      'bicycle-wheelset',
      'carbon-bicycle-handlebar-stem',
      'carbon-bicycle-handlebar'
  );

UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714929000'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug IN ('carbon-bicycle-spoke', 'bicycle-spoke')
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');

UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714930090'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug = 'bicycle-hub'
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');

UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714999011'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug = 'bicycle-wheelset'
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');

UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714999040'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug = 'carbon-bicycle-handlebar-stem'
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');

UPDATE customs_classification_profiles
SET source_url_eu = CASE
        WHEN BTRIM(source_url_eu) = '' THEN 'https://ec.europa.eu/taxation_customs/dds2/taric'
        ELSE source_url_eu
    END,
    source_url_uk = CASE
        WHEN BTRIM(source_url_uk) = '' THEN 'https://www.gov.uk/trade-tariff/8714991020'
        ELSE source_url_uk
    END,
    updated_at = NOW()
WHERE slug = 'carbon-bicycle-handlebar'
  AND source = 'built_in'
  AND (BTRIM(source_url_eu) = '' OR BTRIM(source_url_uk) = '');
