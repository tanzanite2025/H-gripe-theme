-- Record when built-in customs data was checked and when it should be checked again.
-- The 6/8 digit HS/CN values are reusable baseline classifications. Destination
-- authorities may require a longer national tariff code based on material, use,
-- origin, construction, or trade preference.
ALTER TABLE customs_classification_profiles
    ADD COLUMN IF NOT EXISTS verified_at DATE,
    ADD COLUMN IF NOT EXISTS review_due_at DATE;

CREATE INDEX IF NOT EXISTS idx_customs_classification_profiles_review_due_at
    ON customs_classification_profiles(review_due_at);

-- Correct the five templates seeded by migration 385 without overwriting a row
-- that a merchant created with the same slug before the built-in seed existed.
UPDATE customs_classification_profiles
SET
    hs_code = '871492',
    cn_code = '87149210',
    customs_description = 'Bicycle carbon rim',
    source_code = '8714.92.10.00',
    source_url = 'https://hts.usitc.gov/search?query=8714.92',
    notes = 'US HTS 8714.92.10.00 is Wheel rims. UK Trade Tariff 8714921000 is Rims and is valid from 1972-01-01 (https://www.gov.uk/trade-tariff/8714921000). Stored HS/CN values are 6/8 digit reusable baselines; confirm the destination tariff, origin rules, and any special measures before declaration.',
    verified_at = DATE '2026-10-04',
    review_due_at = DATE '2027-10-04'
WHERE slug = 'carbon-bicycle-rim' AND source = 'built_in';

UPDATE customs_classification_profiles
SET
    hs_code = '871492',
    cn_code = '87149290',
    customs_description = 'Bicycle carbon spoke',
    source_code = '8714.92.50.00',
    source_url = 'https://hts.usitc.gov/search?query=8714.92',
    notes = 'US HTS 8714.92.50.00 is Spokes. UK Trade Tariff 8714929000 is Spokes and is valid from 1972-01-01 (https://www.gov.uk/trade-tariff/8714929000). Stored HS/CN values are 6/8 digit reusable baselines; confirm the destination tariff, origin rules, and any special measures before declaration.',
    verified_at = DATE '2026-10-04',
    review_due_at = DATE '2027-10-04'
WHERE slug = 'carbon-bicycle-spoke' AND source = 'built_in';

UPDATE customs_classification_profiles
SET
    hs_code = '871492',
    cn_code = '87149290',
    customs_description = 'Bicycle spoke',
    source_code = '8714.92.50.00',
    source_url = 'https://hts.usitc.gov/search?query=8714.92',
    notes = 'US HTS 8714.92.50.00 is Spokes. UK Trade Tariff 8714929000 is Spokes and is valid from 1972-01-01 (https://www.gov.uk/trade-tariff/8714929000). Stored HS/CN values are 6/8 digit reusable baselines; confirm the destination tariff, origin rules, and any special measures before declaration.',
    verified_at = DATE '2026-10-04',
    review_due_at = DATE '2027-10-04'
WHERE slug = 'bicycle-spoke' AND source = 'built_in';

UPDATE customs_classification_profiles
SET
    hs_code = '871493',
    cn_code = '87149300',
    customs_description = 'Bicycle hub',
    source_code = '8714.93',
    source_url = 'https://hts.usitc.gov/search?query=8714.93',
    notes = 'US HTS heading 8714.93 covers hubs other than coaster braking hubs and hub brakes; the 10 digit line depends on construction, speed, quick release, and duty conditions (for example 8714.93.05.00, 8714.93.15.00, 8714.93.24.00, 8714.93.28.00, or 8714.93.35.00). UK Trade Tariff 8714930000 is the Hubs heading and is valid from 1972-01-01 (https://www.gov.uk/trade-tariff/8714930000). Do not treat one US 10 digit line as universal.',
    verified_at = DATE '2026-10-04',
    review_due_at = DATE '2027-10-04'
WHERE slug = 'bicycle-hub' AND source = 'built_in';

UPDATE customs_classification_profiles
SET
    hs_code = '871499',
    cn_code = '87149990',
    customs_description = 'Bicycle wheelset',
    source_code = '8714.99.80.00',
    source_url = 'https://hts.usitc.gov/search?query=8714.99',
    notes = 'Complete wheels are listed under UK Trade Tariff 8714999011 within CN 87149990 (https://www.gov.uk/trade-tariff/8714999011). The US baseline reference is HTS 8714.99.80.00 Other; destination measures and whether the shipment is a complete wheel may change the final line. Confirm before declaration.',
    verified_at = DATE '2026-10-04',
    review_due_at = DATE '2027-10-04'
WHERE slug = 'bicycle-wheelset' AND source = 'built_in';

-- Carbon stems are separate from the aluminum-alloy stem line in the US HTS.
INSERT INTO customs_classification_profiles (
    name,
    slug,
    component_kind,
    material,
    hs_code,
    cn_code,
    country_of_origin,
    customs_description,
    source,
    source_code,
    source_url,
    notes,
    verified_at,
    review_due_at,
    status,
    created_at,
    updated_at
) VALUES (
    '碳纤维把立',
    'carbon-bicycle-handlebar-stem',
    'handlebar_stem',
    'carbon_fiber',
    '871499',
    '87149990',
    'CN',
    'Bicycle carbon handlebar stem',
    'built_in',
    '8714.99.80.00',
    'https://hts.usitc.gov/search?query=8714.99',
    'The US HTS 8714.99.60.00 line is specifically for bicycle handlebar stems wholly of aluminum alloy and must not be used for a carbon-fibre stem. Carbon stems use the 8714.99 Other baseline here. UK Trade Tariff 8714999040 describes a stem for use in bicycle manufacture from 2020-01-01 (https://www.gov.uk/trade-tariff/8714999040); a finished retail part may require another national subcode. Confirm destination and use before declaration.',
    DATE '2026-10-04',
    DATE '2027-10-04',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO customs_classification_profiles (
    name,
    slug,
    component_kind,
    material,
    hs_code,
    cn_code,
    country_of_origin,
    customs_description,
    source,
    source_code,
    source_url,
    notes,
    verified_at,
    review_due_at,
    status,
    created_at,
    updated_at
) VALUES (
    '碳纤维车把',
    'carbon-bicycle-handlebar',
    'handlebar',
    'carbon_fiber',
    '871499',
    '87149910',
    'CN',
    'Bicycle carbon handlebar',
    'built_in',
    '8714.99.80.00',
    'https://hts.usitc.gov/search?query=8714.99',
    'UK Trade Tariff places handlebars in CN 87149910 and lists carbon-fibre or aluminium bicycle handlebars in the 87149910 branch, with relevant entries beginning 2018-07-01 (https://www.gov.uk/trade-tariff/8714991020). The US baseline reference is HTS 8714.99.80.00 Other; the final national line depends on destination rules and use. Confirm before declaration.',
    DATE '2026-10-04',
    DATE '2027-10-04',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO NOTHING;
