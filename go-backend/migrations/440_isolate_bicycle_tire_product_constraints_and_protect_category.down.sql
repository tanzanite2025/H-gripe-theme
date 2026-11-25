DROP TRIGGER protect_bicycle_tire_system_category_identity_and_tree_position ON product_categories;
DROP FUNCTION protect_bicycle_tire_system_category_identity_and_tree_position();

ALTER TABLE products
    DROP CONSTRAINT products_dedicated_product_type_allowed_check,
    DROP CONSTRAINT products_carbon_rim_type_and_specification_check,
    DROP CONSTRAINT products_bicycle_tire_type_and_specification_check;

ALTER TABLE products
    ADD CONSTRAINT products_dedicated_product_type_specifications_check CHECK (
        (product_type = '' AND carbon_rim_specification IS NULL AND bicycle_tire_specification IS NULL)
        OR (
            product_type = 'carbon_rim'
            AND product_specification_template_id IS NULL
            AND carbon_rim_specification IS NOT NULL
            AND jsonb_typeof(carbon_rim_specification) = 'object'
            AND (carbon_rim_specification->>'schema_version') IS NOT DISTINCT FROM '1'
            AND bicycle_tire_specification IS NULL
        )
        OR (
            product_type = 'bicycle_tire'
            AND product_specification_template_id IS NULL
            AND carbon_rim_specification IS NULL
            AND bicycle_tire_specification IS NOT NULL
            AND jsonb_typeof(bicycle_tire_specification) = 'object'
            AND (bicycle_tire_specification->>'schema_version') IS NOT DISTINCT FROM '1'
            AND (
                NULLIF(BTRIM(bicycle_tire_specification->>'article_no'), '') IS NULL
                OR brand_id IS NOT NULL
            )
        )
    );
