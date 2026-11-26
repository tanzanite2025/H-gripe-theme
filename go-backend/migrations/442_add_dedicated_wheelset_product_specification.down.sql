DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM products WHERE product_type = 'wheelset') THEN
        RAISE EXCEPTION 'Cannot remove wheelset product support while wheelset products exist';
    END IF;
END $$;

DROP TRIGGER protect_wheelset_system_category_identity ON product_categories;
DROP FUNCTION protect_wheelset_system_category_identity();
DROP TRIGGER require_wheelset_product_type_and_system_category_binding ON products;
DROP FUNCTION require_wheelset_product_type_and_system_category_binding();
DROP INDEX idx_products_dedicated_wheelset_type;

ALTER TABLE products
    DROP CONSTRAINT products_wheelset_type_and_specification_check,
    DROP CONSTRAINT products_dedicated_product_type_allowed_check,
    DROP CONSTRAINT products_carbon_rim_type_and_specification_check,
    DROP CONSTRAINT products_bicycle_tire_type_and_specification_check;

ALTER TABLE products
    ADD CONSTRAINT products_dedicated_product_type_allowed_check
        CHECK (product_type IN ('', 'carbon_rim', 'bicycle_tire')),
    ADD CONSTRAINT products_carbon_rim_type_and_specification_check
        CHECK (
            (
                product_type = 'carbon_rim'
                AND product_specification_template_id IS NULL
                AND carbon_rim_specification IS NOT NULL
                AND jsonb_typeof(carbon_rim_specification) = 'object'
                AND (carbon_rim_specification->>'schema_version') IS NOT DISTINCT FROM '1'
            )
            OR (product_type <> 'carbon_rim' AND carbon_rim_specification IS NULL)
        ),
    ADD CONSTRAINT products_bicycle_tire_type_and_specification_check
        CHECK (
            (
                product_type = 'bicycle_tire'
                AND product_specification_template_id IS NULL
                AND bicycle_tire_specification IS NOT NULL
                AND jsonb_typeof(bicycle_tire_specification) = 'object'
                AND (bicycle_tire_specification->>'schema_version') IS NOT DISTINCT FROM '1'
                AND (
                    NULLIF(BTRIM(bicycle_tire_specification->>'article_no'), '') IS NULL
                    OR brand_id IS NOT NULL
                )
            )
            OR (product_type <> 'bicycle_tire' AND bicycle_tire_specification IS NULL)
        );

ALTER TABLE products
    DROP COLUMN wheelset_specification;
