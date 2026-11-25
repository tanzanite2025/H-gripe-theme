DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM products
        WHERE product_type = 'bicycle_tire' OR bicycle_tire_specification IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot remove bicycle tire product schema while bicycle tire products exist';
    END IF;
END $$;

DROP TRIGGER require_bicycle_tire_product_type_and_tire_category ON products;
DROP FUNCTION require_bicycle_tire_product_type_and_tire_category();
DROP INDEX uq_bicycle_tire_brand_article_no;
DROP INDEX idx_products_dedicated_bicycle_tire_type;

ALTER TABLE products
    DROP CONSTRAINT products_dedicated_product_type_specifications_check;

ALTER TABLE products
    ADD CONSTRAINT products_carbon_rim_type_and_specification_check CHECK (
        (product_type = '' AND carbon_rim_specification IS NULL)
        OR (
            product_type = 'carbon_rim'
            AND product_specification_template_id IS NULL
            AND carbon_rim_specification IS NOT NULL
            AND jsonb_typeof(carbon_rim_specification) = 'object'
            AND (carbon_rim_specification->>'schema_version') IS NOT DISTINCT FROM '1'
        )
    );

ALTER TABLE products DROP COLUMN bicycle_tire_specification;
