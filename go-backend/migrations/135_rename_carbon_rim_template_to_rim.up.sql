-- Use a material-neutral product specification template for all rim products.
-- Keep the product specification template ID stable so products, QuickBuy bindings and media
-- relationships continue to point to the same record.
DO $$
DECLARE
    legacy_id BIGINT;
    canonical_id BIGINT;
BEGIN
    SELECT id INTO legacy_id
    FROM product_specification_templates
    WHERE slug = 'carbon_rim'
    LIMIT 1;

    SELECT id INTO canonical_id
    FROM product_specification_templates
    WHERE slug = 'rim'
    LIMIT 1;

    IF legacy_id IS NOT NULL
       AND canonical_id IS NOT NULL
       AND legacy_id <> canonical_id THEN
        RAISE EXCEPTION 'Cannot rename carbon_rim to rim because both product specification templates already exist';
    END IF;

    IF legacy_id IS NOT NULL THEN
        UPDATE product_specification_templates
        SET name = 'Rim',
            slug = 'rim',
            description = 'Rim template. Common rim fields are defined here; specific product/SKU values are maintained on the product/SKU.',
            updated_at = NOW()
        WHERE id = legacy_id;
    ELSIF canonical_id IS NOT NULL THEN
        UPDATE product_specification_templates
        SET name = 'Rim',
            description = 'Rim template. Common rim fields are defined here; specific product/SKU values are maintained on the product/SKU.',
            updated_at = NOW()
        WHERE id = canonical_id;
    END IF;
END $$;

UPDATE product_specification_template_translations
SET name = 'Rim',
    updated_at = NOW()
WHERE product_specification_template_id IN (
    SELECT id
    FROM product_specification_templates
    WHERE slug = 'rim'
)
  AND BTRIM(name) = 'Carbon Rim';
