INSERT INTO product_specification_templates (
    name,
    slug,
    description,
    sort_order,
    is_enabled,
    is_system_managed,
    created_at,
    updated_at
)
VALUES (
    'Schwalbe Tire',
    'schwalbe_tire',
    'Schwalbe official tire facts. Values are sourced from individually verified Schwalbe product pages.',
    70,
    TRUE,
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO NOTHING;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM product_specification_templates
        WHERE slug = 'schwalbe_tire'
          AND (is_system_managed IS NOT TRUE OR is_enabled IS NOT TRUE)
    ) THEN
        RAISE EXCEPTION 'schwalbe_tire template already exists with incompatible system ownership';
    END IF;
END $$;

WITH schwalbe_template AS (
    SELECT id
    FROM product_specification_templates
    WHERE slug = 'schwalbe_tire'
),
seed (
    "group",
    name,
    slug,
    field_type,
    presentation,
    unit,
    is_required,
    is_filterable,
    is_visible,
    role,
    selection_mode,
    min_selections,
    max_selections,
    sort_order
) AS (
    VALUES
        ('Official product facts', 'Article No.', 'article_no', 'text', 'text', '', TRUE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 10),
        ('Official product facts', 'EAN', 'ean', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 20),
        ('Official product facts', 'Product name', 'model_name', 'text', 'text', '', TRUE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 30),
        ('Official product facts', 'ETRTO', 'etrto', 'text', 'text', '', TRUE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 40),
        ('Official product facts', 'Inch', 'inch_designation', 'text', 'text', '', FALSE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 50),
        ('Official product facts', 'Weight', 'weight_g', 'number', 'text', 'g', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 60),
        ('Official product facts', 'Version', 'version_label', 'text', 'text', '', FALSE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 70),
        ('Official product facts', 'Compound', 'compound', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 80),
        ('Official product facts', 'Color', 'color', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 90),
        ('Official product facts', 'Bead', 'bead', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 100),
        ('Official product facts', 'E-Bike', 'e_bike_rating', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 110),
        ('Official product facts', 'Epi', 'epi', 'number', 'text', 'EPI', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 120),
        ('Official product facts', 'Load (kg)', 'load_kg', 'number', 'text', 'kg', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 130),
        ('Official product facts', 'Seal', 'seal', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 140),
        ('Official product facts', 'Tread', 'tread', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 150),
        ('Official product facts', 'min. Bar', 'min_pressure_bar', 'number', 'text', 'bar', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 160),
        ('Official product facts', 'max. Bar', 'max_pressure_bar', 'number', 'text', 'bar', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 170),
        ('Official product facts', 'min. PSI', 'min_pressure_psi', 'number', 'text', 'psi', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 180),
        ('Official product facts', 'max. PSI', 'max_pressure_psi', 'number', 'text', 'psi', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 190)
)
INSERT INTO product_spec_definitions (
    product_specification_template_id,
    "group",
    name,
    slug,
    field_type,
    presentation,
    unit,
    is_required,
    is_filterable,
    is_visible,
    role,
    selection_mode,
    min_selections,
    max_selections,
    sort_order,
    validation
)
SELECT
    schwalbe_template.id,
    seed."group",
    seed.name,
    seed.slug,
    seed.field_type,
    seed.presentation,
    seed.unit,
    seed.is_required,
    seed.is_filterable,
    seed.is_visible,
    seed.role,
    seed.selection_mode,
    seed.min_selections,
    seed.max_selections,
    seed.sort_order,
    ''
FROM schwalbe_template
CROSS JOIN seed
WHERE NOT EXISTS (
    SELECT 1
    FROM product_spec_definitions existing
    WHERE existing.product_specification_template_id = schwalbe_template.id
      AND existing.slug = seed.slug
);

-- Validate the post-condition after the seed has run. The first version of
-- this migration performed the count before the seed, which made a fresh
-- database fail with zero definitions instead of creating the template.
DO $$
DECLARE
    schwalbe_template_id BIGINT;
BEGIN
    SELECT id
    INTO schwalbe_template_id
    FROM product_specification_templates
    WHERE slug = 'schwalbe_tire';

    IF (
        SELECT COUNT(*)
        FROM product_spec_definitions
        WHERE product_specification_template_id = schwalbe_template_id
    ) <> 19 THEN
        RAISE EXCEPTION 'schwalbe_tire template must contain exactly 19 definitions';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM product_spec_definitions
        WHERE product_specification_template_id = schwalbe_template_id
          AND slug NOT IN (
              'article_no',
              'ean',
              'model_name',
              'etrto',
              'inch_designation',
              'weight_g',
              'version_label',
              'compound',
              'color',
              'bead',
              'e_bike_rating',
              'epi',
              'load_kg',
              'seal',
              'tread',
              'min_pressure_bar',
              'max_pressure_bar',
              'min_pressure_psi',
              'max_pressure_psi'
          )
    ) THEN
        RAISE EXCEPTION 'schwalbe_tire template contains an unknown definition';
    END IF;

    IF EXISTS (
        WITH expected (
            slug,
            expected_group,
            expected_name,
            expected_field_type,
            expected_presentation,
            expected_unit,
            expected_is_required,
            expected_is_filterable,
            expected_is_visible,
            expected_role,
            expected_selection_mode,
            expected_min_selections,
            expected_max_selections,
            expected_sort_order,
            expected_validation
        ) AS (
            VALUES
                ('article_no', 'Official product facts', 'Article No.', 'text', 'text', '', TRUE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 10, ''),
                ('ean', 'Official product facts', 'EAN', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 20, ''),
                ('model_name', 'Official product facts', 'Product name', 'text', 'text', '', TRUE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 30, ''),
                ('etrto', 'Official product facts', 'ETRTO', 'text', 'text', '', TRUE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 40, ''),
                ('inch_designation', 'Official product facts', 'Inch', 'text', 'text', '', FALSE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 50, ''),
                ('weight_g', 'Official product facts', 'Weight', 'number', 'text', 'g', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 60, ''),
                ('version_label', 'Official product facts', 'Version', 'text', 'text', '', FALSE, TRUE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 70, ''),
                ('compound', 'Official product facts', 'Compound', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 80, ''),
                ('color', 'Official product facts', 'Color', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 90, ''),
                ('bead', 'Official product facts', 'Bead', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 100, ''),
                ('e_bike_rating', 'Official product facts', 'E-Bike', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 110, ''),
                ('epi', 'Official product facts', 'Epi', 'number', 'text', 'EPI', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 120, ''),
                ('load_kg', 'Official product facts', 'Load (kg)', 'number', 'text', 'kg', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 130, ''),
                ('seal', 'Official product facts', 'Seal', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 140, ''),
                ('tread', 'Official product facts', 'Tread', 'text', 'text', '', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 150, ''),
                ('min_pressure_bar', 'Official product facts', 'min. Bar', 'number', 'text', 'bar', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 160, ''),
                ('max_pressure_bar', 'Official product facts', 'max. Bar', 'number', 'text', 'bar', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 170, ''),
                ('min_pressure_psi', 'Official product facts', 'min. PSI', 'number', 'text', 'psi', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 180, ''),
                ('max_pressure_psi', 'Official product facts', 'max. PSI', 'number', 'text', 'psi', FALSE, FALSE, TRUE, 'attribute', 'single', 0, NULL::INTEGER, 190, '')
        )
        SELECT 1
        FROM expected
        LEFT JOIN product_spec_definitions actual
          ON actual.product_specification_template_id = schwalbe_template_id
         AND actual.slug = expected.slug
        WHERE actual.id IS NULL
           OR actual."group" IS DISTINCT FROM expected.expected_group
           OR actual.name IS DISTINCT FROM expected.expected_name
           OR actual.field_type IS DISTINCT FROM expected.expected_field_type
           OR actual.presentation IS DISTINCT FROM expected.expected_presentation
           OR actual.unit IS DISTINCT FROM expected.expected_unit
           OR actual.is_required IS DISTINCT FROM expected.expected_is_required
           OR actual.is_filterable IS DISTINCT FROM expected.expected_is_filterable
           OR actual.is_visible IS DISTINCT FROM expected.expected_is_visible
           OR actual.role IS DISTINCT FROM expected.expected_role
           OR actual.selection_mode IS DISTINCT FROM expected.expected_selection_mode
           OR actual.min_selections IS DISTINCT FROM expected.expected_min_selections
           OR actual.max_selections IS DISTINCT FROM expected.expected_max_selections
           OR actual.sort_order IS DISTINCT FROM expected.expected_sort_order
           OR actual.validation IS DISTINCT FROM expected.expected_validation
    ) THEN
        RAISE EXCEPTION 'schwalbe_tire template definition structure does not match the official 19-field contract';
    END IF;
END $$;
