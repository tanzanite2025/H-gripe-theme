-- Reset product templates to a single database-owned source of truth.
-- The storefront and admin product forms must read these records instead of
-- carrying separate hard-coded template presets in the frontend.

DELETE FROM shipping_template_bindings
WHERE scope IN ('product_type', 'product', 'variant');

DELETE FROM shipping_packaging_rule_applies;

TRUNCATE TABLE product_types RESTART IDENTITY CASCADE;

INSERT INTO product_types (name, slug, description, sort_order, is_enabled)
VALUES
    ('Carbon Rim', 'carbon_rim', 'Carbon rim template. Specific SKU price, stock, shipping weight and option values are maintained on the product/SKU.', 10, TRUE),
    ('Carbon Frame', 'carbon_frame', 'Carbon frame template. Sizes and product-specific values are maintained on the product/SKU.', 20, TRUE),
    ('Wheelset', 'wheelset', 'Wheelset template. Hub, freehub and depth values are maintained on the product/SKU.', 30, TRUE),
    ('Handlebar', 'handlebar', 'Handlebar and cockpit template. Width, stem length and related values are maintained on the product/SKU.', 40, TRUE);

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
        ('carbon_rim', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '["Carbon Fiber","Aluminum"]'),
        ('carbon_rim', '规格', 'Brake Type', 'brake_type', 'select', '', FALSE, TRUE, TRUE, TRUE, 20, '["Disc Brake","Rim Brake"]'),
        ('carbon_rim', '规格', 'Tire Type', 'tire_type', 'select', '', FALSE, TRUE, TRUE, FALSE, 30, '["Clincher","Tubeless","Tubular"]'),
        ('carbon_rim', '规格', 'Wheel Size', 'wheel_size', 'text', '', FALSE, TRUE, TRUE, TRUE, 40, ''),
        ('carbon_rim', '规格', 'Rim Depth', 'rim_depth', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 50, ''),
        ('carbon_rim', '规格', 'Inner Width', 'inner_width', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 60, ''),
        ('carbon_rim', '规格', 'Outer Width', 'outer_width', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 70, ''),
        ('carbon_rim', '规格', 'Spoke Holes', 'spoke_holes', 'number', 'H', FALSE, TRUE, TRUE, TRUE, 80, ''),
        ('carbon_rim', '规格', 'ERD', 'erd', 'number', 'mm', FALSE, FALSE, TRUE, FALSE, 90, ''),

        ('carbon_frame', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '["Carbon Fiber","Aluminum","Titanium","Steel"]'),
        ('carbon_frame', '规格', 'Frame Size', 'frame_size', 'text', '', FALSE, TRUE, TRUE, TRUE, 20, ''),
        ('carbon_frame', '规格', 'Wheel Size', 'wheel_size', 'text', '', FALSE, TRUE, TRUE, FALSE, 30, ''),
        ('carbon_frame', '规格', 'Brake Type', 'brake_type', 'select', '', FALSE, TRUE, TRUE, FALSE, 40, '["Disc Brake","Rim Brake"]'),
        ('carbon_frame', '规格', 'Head Tube Standard', 'headtube_standard', 'text', '', FALSE, FALSE, TRUE, FALSE, 50, ''),
        ('carbon_frame', '规格', 'Bottom Bracket', 'bottom_bracket', 'text', '', FALSE, TRUE, TRUE, FALSE, 60, ''),
        ('carbon_frame', '规格', 'Axle Standard', 'axle_standard', 'text', '', FALSE, TRUE, TRUE, FALSE, 70, ''),
        ('carbon_frame', '规格', 'Seatpost Standard', 'seatpost_standard', 'text', '', FALSE, FALSE, TRUE, FALSE, 80, ''),

        ('wheelset', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '["Carbon Fiber","Aluminum"]'),
        ('wheelset', '规格', 'Brake Type', 'brake_type', 'select', '', FALSE, TRUE, TRUE, FALSE, 20, '["Disc Brake","Rim Brake"]'),
        ('wheelset', '规格', 'Tire Type', 'tire_type', 'select', '', FALSE, TRUE, TRUE, FALSE, 30, '["Clincher","Tubeless","Tubular"]'),
        ('wheelset', '规格', 'Wheel Size', 'wheel_size', 'text', '', FALSE, TRUE, TRUE, TRUE, 40, ''),
        ('wheelset', '规格', 'Front Rim Depth', 'front_rim_depth', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 50, ''),
        ('wheelset', '规格', 'Rear Rim Depth', 'rear_rim_depth', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 60, ''),
        ('wheelset', '规格', 'Hub Interface', 'hub_interface', 'text', '', FALSE, TRUE, TRUE, TRUE, 70, ''),
        ('wheelset', '规格', 'Freehub Body', 'freehub_body', 'text', '', FALSE, TRUE, TRUE, TRUE, 80, ''),
        ('wheelset', '规格', 'Spoke Type', 'spoke_type', 'text', '', FALSE, FALSE, TRUE, FALSE, 90, ''),

        ('handlebar', '规格', 'Material', 'material', 'select', '', FALSE, TRUE, TRUE, FALSE, 10, '["Carbon Fiber","Aluminum"]'),
        ('handlebar', '规格', 'Bar Width', 'bar_width', 'number', 'mm', FALSE, TRUE, TRUE, TRUE, 20, ''),
        ('handlebar', '规格', 'Stem Length', 'stem_length', 'number', 'mm', FALSE, TRUE, TRUE, TRUE, 30, ''),
        ('handlebar', '规格', 'Reach', 'reach', 'number', 'mm', FALSE, FALSE, TRUE, FALSE, 40, ''),
        ('handlebar', '规格', 'Drop', 'drop', 'number', 'mm', FALSE, FALSE, TRUE, FALSE, 50, ''),
        ('handlebar', '规格', 'Clamp Diameter', 'clamp_diameter', 'number', 'mm', FALSE, TRUE, TRUE, FALSE, 60, ''),
        ('handlebar', '规格', 'Cable Routing', 'cable_routing', 'select', '', FALSE, TRUE, TRUE, FALSE, 70, '["Internal","Semi-internal","External"]')
) AS seed(type_slug, "group", name, slug, field_type, unit, is_required, is_filterable, is_visible, is_variant_option, sort_order, options)
    ON seed.type_slug = pt.slug;
