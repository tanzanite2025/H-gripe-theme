-- Add independently sellable hub and spoke product templates.
-- Product-specific values remain on products/SKUs.
INSERT INTO product_types (name, slug, description, sort_order, is_enabled)
VALUES
    ('Hub', 'hub', 'Hub template. Common hub fields are defined here; specific product/SKU values are maintained on the product/SKU.', 50, TRUE),
    ('Spoke', 'spoke', 'Spoke template. Common spoke fields are defined here; specific product/SKU values are maintained on the product/SKU.', 60, TRUE)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO product_spec_definitions (
    product_type_id, "group", name, slug, field_type, unit,
    is_required, is_filterable, is_visible, is_variant_option,
    sort_order, options, validation
)
SELECT pt.id, seed."group", seed.name, seed.slug, seed.field_type, seed.unit,
       seed.is_required, seed.is_filterable, seed.is_visible, seed.is_variant_option,
       seed.sort_order, seed.options, ''
FROM product_types pt
JOIN (
    VALUES
        ('hub', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '[]'),
        ('hub', '规格', 'Wheel Position', 'wheel_position', 'select', '', FALSE, TRUE, TRUE, FALSE, 20, '[]'),
        ('hub', '规格', 'Brake Interface', 'brake_interface', 'select', '', FALSE, TRUE, TRUE, TRUE, 30, '[]'),
        ('hub', '规格', 'Axle Standard', 'axle_standard', 'text', '', FALSE, TRUE, TRUE, TRUE, 40, ''),
        ('hub', '规格', 'Freehub Body', 'freehub_body', 'text', '', FALSE, TRUE, TRUE, TRUE, 50, ''),
        ('hub', '规格', 'Spoke Holes', 'spoke_holes', 'number', 'H', FALSE, TRUE, TRUE, TRUE, 60, ''),
        ('hub', '规格', 'Bearing Type', 'bearing_type', 'text', '', FALSE, TRUE, TRUE, FALSE, 70, ''),

        ('spoke', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '[]'),
        ('spoke', '规格', 'Spoke Type', 'spoke_type', 'text', '', FALSE, TRUE, TRUE, FALSE, 20, ''),
        ('spoke', '规格', 'Spoke Length', 'spoke_length', 'number', 'mm', FALSE, TRUE, TRUE, TRUE, 30, ''),
        ('spoke', '规格', 'Spoke Diameter', 'spoke_diameter', 'number', 'mm', FALSE, TRUE, TRUE, TRUE, 40, ''),
        ('spoke', '规格', 'Nipple Type', 'nipple_type', 'select', '', FALSE, TRUE, TRUE, TRUE, 50, '[]'),
        ('spoke', '规格', 'Nipple Material', 'nipple_material', 'select', '', FALSE, TRUE, TRUE, FALSE, 60, '[]')
) AS seed(type_slug, "group", name, slug, field_type, unit, is_required, is_filterable, is_visible, is_variant_option, sort_order, options)
    ON seed.type_slug = pt.slug
WHERE NOT EXISTS (
    SELECT 1
    FROM product_spec_definitions existing
    WHERE existing.product_type_id = pt.id
      AND existing.slug = seed.slug
);
