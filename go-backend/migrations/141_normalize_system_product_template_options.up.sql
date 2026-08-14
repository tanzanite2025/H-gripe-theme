-- Product templates own the reusable field skeleton.
-- Concrete values belong to products and SKUs, so system select fields must
-- not carry a hard-coded catalog of values.
UPDATE product_spec_definitions definition
SET options = '[]',
    updated_at = NOW()
FROM product_types product_type
WHERE definition.product_type_id = product_type.id
  AND product_type.slug IN ('rim', 'carbon_frame', 'wheelset', 'handlebar', 'hub', 'spoke')
  AND definition.field_type = 'select'
  AND COALESCE(BTRIM(definition.options), '') <> '[]';
