-- These are stable baseline customs templates used by the bicycle product catalog.
-- They are inserted only when the slug does not already exist, so a merchant's
-- existing template or later edits are never overwritten by a deployment.
-- HS/CN values are the project's reusable bicycle-parts baseline. Product-level
-- customs data can still override them when a destination requires a different
-- classification.
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
    status,
    created_at,
    updated_at
) VALUES
    (
        '碳纤维车圈',
        'carbon-bicycle-rim',
        'rim',
        'carbon_fiber',
        '871499',
        '87149990',
        'CN',
        'Bicycle carbon rim',
        'built_in',
        '8714.99.80',
        'https://hts.usitc.gov/search?query=8714.99',
        'Built-in baseline template for carbon bicycle rims. Confirm the destination tariff before shipment.',
        'active',
        NOW(),
        NOW()
    ),
    (
        '碳纤维辐条',
        'carbon-bicycle-spoke',
        'spoke',
        'carbon_fiber',
        '871499',
        '87149990',
        'CN',
        'Bicycle carbon spoke',
        'built_in',
        '8714.99.80',
        'https://hts.usitc.gov/search?query=8714.99',
        'Built-in baseline template for carbon bicycle spokes. Confirm the destination tariff before shipment.',
        'active',
        NOW(),
        NOW()
    ),
    (
        '自行车辐条',
        'bicycle-spoke',
        'spoke',
        '',
        '871499',
        '87149990',
        'CN',
        'Bicycle spoke',
        'built_in',
        '8714.99.80',
        'https://hts.usitc.gov/search?query=8714.99',
        'Built-in baseline template for bicycle spokes. Confirm the destination tariff before shipment.',
        'active',
        NOW(),
        NOW()
    ),
    (
        '自行车花鼓',
        'bicycle-hub',
        'hub',
        '',
        '871499',
        '87149990',
        'CN',
        'Bicycle hub',
        'built_in',
        '8714.99.80',
        'https://hts.usitc.gov/search?query=8714.99',
        'Built-in baseline template for bicycle hubs. Confirm the destination tariff before shipment.',
        'active',
        NOW(),
        NOW()
    ),
    (
        '自行车轮组',
        'bicycle-wheelset',
        'wheelset',
        '',
        '871499',
        '87149990',
        'CN',
        'Bicycle wheelset',
        'built_in',
        '8714.99.80',
        'https://hts.usitc.gov/search?query=8714.99',
        'Built-in baseline template for bicycle wheelsets. Confirm the destination tariff before shipment.',
        'active',
        NOW(),
        NOW()
    )
ON CONFLICT (slug) DO NOTHING;
