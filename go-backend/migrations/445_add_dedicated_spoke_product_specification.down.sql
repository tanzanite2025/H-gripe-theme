DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM products
        WHERE product_type = 'spoke' OR spoke_specification IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot remove spoke product support while spoke products or specifications exist';
    END IF;
END $$;

DROP TRIGGER require_spoke_product_type_and_system_category_binding ON products;
DROP FUNCTION require_spoke_product_type_and_system_category_binding();
DROP INDEX idx_products_dedicated_spoke_type;

ALTER TABLE products
    DROP CONSTRAINT products_spoke_type_and_specification_check,
    DROP CONSTRAINT products_dedicated_product_type_allowed_check,
    DROP CONSTRAINT products_carbon_rim_type_and_specification_check,
    DROP CONSTRAINT products_bicycle_tire_type_and_specification_check,
    DROP CONSTRAINT products_wheelset_type_and_specification_check;

ALTER TABLE products
    ADD CONSTRAINT products_dedicated_product_type_allowed_check
        CHECK (product_type IN ('', 'carbon_rim', 'bicycle_tire', 'wheelset')),
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
        ),
    ADD CONSTRAINT products_wheelset_type_and_specification_check
        CHECK (
            (
                product_type = 'wheelset'
                AND product_specification_template_id IS NULL
                AND wheelset_specification IS NOT NULL
                AND jsonb_typeof(wheelset_specification) = 'object'
                AND (wheelset_specification->>'schema_version') IS NOT DISTINCT FROM '1'
                AND jsonb_typeof(wheelset_specification->'front_wheel') = 'object'
                AND NULLIF(BTRIM(wheelset_specification->'front_wheel'->>'wheel_size'), '') IS NOT NULL
                AND CASE
                    WHEN jsonb_typeof(wheelset_specification->'front_wheel'->'spoke_hole_counts') = 'array'
                    THEN jsonb_array_length(wheelset_specification->'front_wheel'->'spoke_hole_counts') > 0
                    ELSE FALSE
                END
                AND CASE
                    WHEN jsonb_typeof(wheelset_specification->'front_wheel'->'lacing_patterns') = 'array'
                    THEN jsonb_array_length(wheelset_specification->'front_wheel'->'lacing_patterns') > 0
                    ELSE FALSE
                END
                AND jsonb_typeof(wheelset_specification->'rear_wheel') = 'object'
                AND NULLIF(BTRIM(wheelset_specification->'rear_wheel'->>'wheel_size'), '') IS NOT NULL
                AND CASE
                    WHEN jsonb_typeof(wheelset_specification->'rear_wheel'->'spoke_hole_counts') = 'array'
                    THEN jsonb_array_length(wheelset_specification->'rear_wheel'->'spoke_hole_counts') > 0
                    ELSE FALSE
                END
                AND CASE
                    WHEN jsonb_typeof(wheelset_specification->'rear_wheel'->'lacing_patterns') = 'array'
                    THEN jsonb_array_length(wheelset_specification->'rear_wheel'->'lacing_patterns') > 0
                    ELSE FALSE
                END
            ) IS TRUE
            OR (product_type <> 'wheelset' AND wheelset_specification IS NULL)
        );

CREATE FUNCTION reject_product_binding_to_unimplemented_spokes_category()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    category_slug TEXT;
BEGIN
    SELECT slug INTO category_slug
    FROM product_categories
    WHERE id = NEW.product_category_id;

    IF category_slug = 'spokes' THEN
        RAISE EXCEPTION 'spokes products require the dedicated spoke product type, which is not implemented yet' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER reject_product_binding_to_unimplemented_spokes_category
BEFORE INSERT OR UPDATE OF product_category_id, product_type ON products
FOR EACH ROW EXECUTE FUNCTION reject_product_binding_to_unimplemented_spokes_category();

ALTER TABLE products
    DROP COLUMN spoke_specification;
