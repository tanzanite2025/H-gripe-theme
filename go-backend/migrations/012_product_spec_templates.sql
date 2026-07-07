-- Product type and specification templates.
-- Backend is the source of truth for product-specific fields used by admin forms and storefront rendering.

CREATE TABLE IF NOT EXISTS product_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) UNIQUE NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_spec_definitions (
    id BIGSERIAL PRIMARY KEY,
    product_type_id BIGINT NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    "group" VARCHAR(80) NOT NULL DEFAULT 'specs',
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    field_type VARCHAR(32) NOT NULL DEFAULT 'text',
    unit VARCHAR(32),
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    options TEXT,
    validation TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_product_type_spec_slug UNIQUE (product_type_id, slug)
);

ALTER TABLE products ADD COLUMN IF NOT EXISTS product_type_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_product_type'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_product_type
            FOREIGN KEY (product_type_id) REFERENCES product_types(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_product_type_id ON products(product_type_id);

CREATE TABLE IF NOT EXISTS product_spec_values (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    spec_definition_id BIGINT NOT NULL REFERENCES product_spec_definitions(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_product_spec_value UNIQUE (product_id, spec_definition_id)
);

CREATE INDEX IF NOT EXISTS idx_product_spec_values_product_id ON product_spec_values(product_id);
CREATE INDEX IF NOT EXISTS idx_product_spec_values_definition_id ON product_spec_values(spec_definition_id);
CREATE INDEX IF NOT EXISTS idx_product_spec_values_value ON product_spec_values(value);

INSERT INTO product_types (name, slug, description, sort_order, is_enabled, created_at, updated_at)
VALUES
    ('Carbon Fiber Rim', 'carbon_rim', 'Carbon bicycle rim specification template.', 10, TRUE, NOW(), NOW()),
    ('Bicycle Frame', 'frame', 'Bicycle frame geometry and compatibility template.', 20, TRUE, NOW(), NOW())
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    is_enabled = EXCLUDED.is_enabled,
    updated_at = NOW();

WITH carbon AS (
    SELECT id AS product_type_id FROM product_types WHERE slug = 'carbon_rim'
)
INSERT INTO product_spec_definitions (
    product_type_id, "group", name, slug, field_type, unit,
    is_required, is_filterable, is_visible, sort_order, options, created_at, updated_at
)
SELECT carbon.product_type_id, data.group_name, data.name, data.slug, data.field_type, data.unit,
       data.is_required, data.is_filterable, TRUE, data.sort_order, data.options, NOW(), NOW()
FROM carbon
CROSS JOIN (
    VALUES
        ('Dimensions', 'Outer Width', 'outer_width_mm', 'number', 'mm', TRUE, TRUE, 10, NULL),
        ('Dimensions', 'Inner Width', 'inner_width_mm', 'number', 'mm', TRUE, TRUE, 20, NULL),
        ('Dimensions', 'Rim Depth', 'rim_depth_mm', 'number', 'mm', FALSE, TRUE, 30, NULL),
        ('Compatibility', 'Wheel Size', 'wheel_size', 'select', '', TRUE, TRUE, 40, '["700c","650b","29inch","27.5inch","20inch","451"]'),
        ('Compatibility', 'Brake Type', 'brake_type', 'select', '', TRUE, TRUE, 50, '["disc","rim"]'),
        ('Compatibility', 'Tubeless Ready', 'tubeless_ready', 'boolean', '', FALSE, TRUE, 60, NULL),
        ('Compatibility', 'Recommended Tire Width', 'recommended_tire_width', 'text', 'mm', FALSE, FALSE, 70, NULL),
        ('Safety', 'Max Tire Pressure', 'max_tire_pressure_psi', 'number', 'psi', FALSE, FALSE, 80, NULL),
        ('Build', 'ERD', 'erd_mm', 'number', 'mm', FALSE, FALSE, 90, NULL),
        ('Build', 'Spoke Holes', 'spoke_holes', 'number', 'holes', FALSE, TRUE, 100, NULL),
        ('Weight', 'Weight', 'rim_weight_grams', 'number', 'g', FALSE, TRUE, 110, NULL)
) AS data(group_name, name, slug, field_type, unit, is_required, is_filterable, sort_order, options)
ON CONFLICT (product_type_id, slug) DO UPDATE SET
    "group" = EXCLUDED."group",
    name = EXCLUDED.name,
    field_type = EXCLUDED.field_type,
    unit = EXCLUDED.unit,
    is_required = EXCLUDED.is_required,
    is_filterable = EXCLUDED.is_filterable,
    is_visible = EXCLUDED.is_visible,
    sort_order = EXCLUDED.sort_order,
    options = EXCLUDED.options,
    updated_at = NOW();

WITH frame AS (
    SELECT id AS product_type_id FROM product_types WHERE slug = 'frame'
)
INSERT INTO product_spec_definitions (
    product_type_id, "group", name, slug, field_type, unit,
    is_required, is_filterable, is_visible, sort_order, options, created_at, updated_at
)
SELECT frame.product_type_id, data.group_name, data.name, data.slug, data.field_type, data.unit,
       data.is_required, data.is_filterable, TRUE, data.sort_order, data.options, NOW(), NOW()
FROM frame
CROSS JOIN (
    VALUES
        ('Sizing', 'Frame Size', 'frame_size', 'select', '', TRUE, TRUE, 10, '["XS","S","M","L","XL","46","49","52","54","56","58"]'),
        ('Geometry', 'Reach', 'reach_mm', 'number', 'mm', FALSE, TRUE, 20, NULL),
        ('Geometry', 'Stack', 'stack_mm', 'number', 'mm', FALSE, TRUE, 30, NULL),
        ('Geometry', 'Head Tube', 'head_tube_mm', 'number', 'mm', FALSE, FALSE, 40, NULL),
        ('Geometry', 'Seat Tube', 'seat_tube_mm', 'number', 'mm', FALSE, FALSE, 50, NULL),
        ('Geometry', 'Top Tube', 'top_tube_mm', 'number', 'mm', FALSE, FALSE, 60, NULL),
        ('Compatibility', 'Bottom Bracket Standard', 'bb_standard', 'select', '', FALSE, TRUE, 70, '["BSA","BB86","BB30","PF30","T47"]'),
        ('Compatibility', 'Axle Standard', 'axle_standard', 'select', '', FALSE, TRUE, 80, '["QR","12x142","12x148","15x100","15x110"]'),
        ('Compatibility', 'Max Tire Width', 'max_tire_width_mm', 'number', 'mm', FALSE, TRUE, 90, NULL),
        ('Geometry', 'Fork Offset', 'fork_offset_mm', 'number', 'mm', FALSE, FALSE, 100, NULL),
        ('Material', 'Material', 'material', 'select', '', TRUE, TRUE, 110, '["carbon","aluminum","steel","titanium"]')
) AS data(group_name, name, slug, field_type, unit, is_required, is_filterable, sort_order, options)
ON CONFLICT (product_type_id, slug) DO UPDATE SET
    "group" = EXCLUDED."group",
    name = EXCLUDED.name,
    field_type = EXCLUDED.field_type,
    unit = EXCLUDED.unit,
    is_required = EXCLUDED.is_required,
    is_filterable = EXCLUDED.is_filterable,
    is_visible = EXCLUDED.is_visible,
    sort_order = EXCLUDED.sort_order,
    options = EXCLUDED.options,
    updated_at = NOW();
