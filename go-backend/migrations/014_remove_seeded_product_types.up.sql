-- Product types are managed by administrators and must not be pre-populated.
DELETE FROM product_types
WHERE slug IN ('carbon_rim', 'frame');
