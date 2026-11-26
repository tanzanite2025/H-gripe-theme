ALTER TABLE products
    ADD COLUMN wheelset_specification JSONB;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM product_categories
        WHERE slug = 'wheelset' AND is_enabled
    ) THEN
        RAISE EXCEPTION 'The enabled wheelset system category is required';
    END IF;
END $$;

ALTER TABLE products
    DROP CONSTRAINT products_dedicated_product_type_allowed_check,
    DROP CONSTRAINT products_carbon_rim_type_and_specification_check,
    DROP CONSTRAINT products_bicycle_tire_type_and_specification_check;

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
                AND (
                    product_specification_template_id IS NULL
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
            )
            OR (product_type <> 'wheelset' AND wheelset_specification IS NULL)
        );

CREATE INDEX idx_products_dedicated_wheelset_type
    ON products(product_type)
    WHERE product_type = 'wheelset' AND deleted_at IS NULL;

CREATE FUNCTION require_wheelset_product_type_and_system_category_binding()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    category_slug TEXT;
    category_enabled BOOLEAN;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'wheelset' OR OLD.product_type = 'wheelset') THEN
        RAISE EXCEPTION 'Dedicated wheelset product type is immutable' USING ERRCODE = '23514';
    END IF;

    IF NEW.product_type = 'wheelset' THEN
        SELECT slug, is_enabled INTO category_slug, category_enabled
        FROM product_categories
        WHERE id = NEW.product_category_id;

        IF category_slug IS DISTINCT FROM 'wheelset' OR category_enabled IS DISTINCT FROM TRUE THEN
            RAISE EXCEPTION 'wheelset products must belong to the enabled wheelset system category' USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER require_wheelset_product_type_and_system_category_binding
BEFORE INSERT OR UPDATE OF product_category_id, product_type ON products
FOR EACH ROW EXECUTE FUNCTION require_wheelset_product_type_and_system_category_binding();

CREATE FUNCTION protect_wheelset_system_category_identity()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'wheelset' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The wheelset system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The wheelset system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'wheelset' AND NOT NEW.is_enabled THEN
        RAISE EXCEPTION 'The wheelset system category must remain enabled' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER protect_wheelset_system_category_identity
BEFORE INSERT OR UPDATE OR DELETE ON product_categories
FOR EACH ROW EXECUTE FUNCTION protect_wheelset_system_category_identity();

COMMENT ON COLUMN products.wheelset_specification IS
    'Versioned product-level front and rear wheel facts for wheelset products; not a generic product template.';
