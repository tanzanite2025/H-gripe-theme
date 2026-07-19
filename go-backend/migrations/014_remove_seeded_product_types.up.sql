-- Product types are managed by administrators and must not be pre-populated.
DELETE FROM product_spec_values
WHERE spec_definition_id IN (
    SELECT id
    FROM product_spec_definitions
    WHERE product_type_id IN (
        SELECT id FROM product_types WHERE slug IN ('carbon_rim', 'frame')
    )
);

DELETE FROM product_spec_definitions
WHERE product_type_id IN (
    SELECT id FROM product_types WHERE slug IN ('carbon_rim', 'frame')
);

DELETE FROM product_types
WHERE slug IN ('carbon_rim', 'frame');
