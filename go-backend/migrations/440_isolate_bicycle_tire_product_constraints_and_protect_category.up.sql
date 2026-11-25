LOCK TABLE product_categories, products IN SHARE ROW EXCLUSIVE MODE;

DO $$
DECLARE
    tire_category_id BIGINT;
BEGIN
    SELECT tire.id INTO tire_category_id
    FROM product_categories tire
    JOIN product_categories wheel_components ON wheel_components.id = tire.parent_id
    WHERE tire.slug = 'tire'
      AND tire.depth = 2
      AND tire.is_enabled
      AND wheel_components.slug = 'wheel-components'
      AND wheel_components.parent_id IS NULL
      AND wheel_components.depth = 1
      AND wheel_components.is_enabled;

    IF tire_category_id IS NULL THEN
        RAISE EXCEPTION 'The enabled tire category under the wheel-components root is required';
    END IF;

    IF EXISTS (
        SELECT 1 FROM products
        WHERE product_type = 'bicycle_tire'
          AND product_category_id IS DISTINCT FROM tire_category_id
    ) THEN
        RAISE EXCEPTION 'Existing bicycle_tire products must use the tire system category';
    END IF;
END $$;

ALTER TABLE products
    DROP CONSTRAINT products_dedicated_product_type_specifications_check;

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

CREATE FUNCTION protect_bicycle_tire_system_category_identity_and_tree_position()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'tire' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The tire system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The tire system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'tire' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent
            WHERE parent.id = NEW.parent_id
              AND parent.slug = 'wheel-components'
              AND parent.parent_id IS NULL
              AND parent.depth = 1
              AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'tire must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER protect_bicycle_tire_system_category_identity_and_tree_position
BEFORE INSERT OR UPDATE OR DELETE ON product_categories
FOR EACH ROW EXECUTE FUNCTION protect_bicycle_tire_system_category_identity_and_tree_position();
