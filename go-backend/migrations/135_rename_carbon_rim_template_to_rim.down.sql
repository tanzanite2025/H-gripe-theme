-- Restore the legacy product type name and slug.
DO $$
DECLARE
    legacy_id BIGINT;
    canonical_id BIGINT;
BEGIN
    SELECT id INTO legacy_id
    FROM product_types
    WHERE slug = 'carbon_rim'
    LIMIT 1;

    SELECT id INTO canonical_id
    FROM product_types
    WHERE slug = 'rim'
    LIMIT 1;

    IF legacy_id IS NOT NULL
       AND canonical_id IS NOT NULL
       AND legacy_id <> canonical_id THEN
        RAISE EXCEPTION 'Cannot restore carbon_rim because both product types already exist';
    END IF;

    IF canonical_id IS NOT NULL THEN
        UPDATE product_types
        SET name = 'Carbon Rim',
            slug = 'carbon_rim',
            description = 'Carbon rim template. Specific SKU price, stock, shipping weight and option values are maintained on the product/SKU.',
            updated_at = NOW()
        WHERE id = canonical_id;
    END IF;
END $$;

UPDATE product_type_translations
SET name = 'Carbon Rim',
    updated_at = NOW()
WHERE product_type_id IN (
    SELECT id
    FROM product_types
    WHERE slug = 'carbon_rim'
)
  AND BTRIM(name) = 'Rim';
