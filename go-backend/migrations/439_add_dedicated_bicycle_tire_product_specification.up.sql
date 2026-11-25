ALTER TABLE products
    ADD COLUMN bicycle_tire_specification JSONB;

ALTER TABLE products
    DROP CONSTRAINT products_carbon_rim_type_and_specification_check;

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

CREATE INDEX idx_products_dedicated_bicycle_tire_type
    ON products (product_type)
    WHERE product_type = 'bicycle_tire' AND deleted_at IS NULL;

CREATE UNIQUE INDEX uq_bicycle_tire_brand_article_no
    ON products (
        brand_id,
        LOWER(BTRIM(bicycle_tire_specification->>'article_no'))
    )
    WHERE product_type = 'bicycle_tire'
      AND parent_id IS NULL
      AND deleted_at IS NULL
      AND NULLIF(BTRIM(bicycle_tire_specification->>'article_no'), '') IS NOT NULL;

CREATE FUNCTION require_bicycle_tire_product_type_and_tire_category()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    category_slug TEXT;
BEGIN
    IF TG_OP = 'UPDATE'
       AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'bicycle_tire' OR OLD.product_type = 'bicycle_tire') THEN
        RAISE EXCEPTION 'Dedicated bicycle_tire product type is immutable' USING ERRCODE = '23514';
    END IF;

    IF NEW.product_type = 'bicycle_tire' THEN
        SELECT slug INTO category_slug
        FROM product_categories
        WHERE id = NEW.product_category_id;

        IF category_slug IS DISTINCT FROM 'tire' THEN
            RAISE EXCEPTION 'bicycle_tire products must belong to the tire category' USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER require_bicycle_tire_product_type_and_tire_category
BEFORE INSERT OR UPDATE OF product_type, product_category_id ON products
FOR EACH ROW EXECUTE FUNCTION require_bicycle_tire_product_type_and_tire_category();

COMMENT ON COLUMN products.bicycle_tire_specification IS
    'Versioned product-level bicycle tire facts; optional technical values are omitted when unknown.';
