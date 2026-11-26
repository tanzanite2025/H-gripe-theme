LOCK TABLE product_categories, product_category_translations, products, quick_buy_step_product_categories
    IN SHARE ROW EXCLUSIVE MODE;

DO $$
DECLARE
    wheel_components_id BIGINT;
BEGIN
    SELECT id INTO wheel_components_id
    FROM product_categories
    WHERE slug = 'wheel-components'
      AND parent_id IS NULL
      AND depth = 1
      AND is_enabled;

    IF wheel_components_id IS NULL THEN
        RAISE EXCEPTION 'The enabled wheel-components root category is required before adding spokes';
    END IF;

    IF EXISTS (SELECT 1 FROM product_categories WHERE slug = 'spokes') THEN
        RAISE EXCEPTION 'The spokes category slug is already in use; resolve its identity before seeding the system category';
    END IF;
END $$;

INSERT INTO product_categories (
    parent_id,
    name,
    slug,
    description,
    depth,
    sort_order,
    is_enabled,
    created_at,
    updated_at
)
SELECT
    id,
    'Spokes',
    'spokes',
    'System category for individual bicycle spoke products.',
    2,
    25,
    TRUE,
    NOW(),
    NOW()
FROM product_categories
WHERE slug = 'wheel-components';

INSERT INTO product_category_translations (
    product_category_id,
    locale,
    name,
    description,
    created_at,
    updated_at
)
SELECT
    category.id,
    translation.locale,
    translation.name,
    translation.description,
    NOW(),
    NOW()
FROM product_categories category
CROSS JOIN (
    VALUES
        ('en', 'Spokes', 'Individual bicycle spoke products.'),
        ('zh_cn', '辐条', '独立销售的自行车辐条产品。')
) AS translation(locale, name, description)
WHERE category.slug = 'spokes';

CREATE FUNCTION protect_spokes_system_category_identity_and_tree_position()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug = 'spokes' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'The spokes system category cannot be deleted' USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'The spokes system category identity cannot be changed' USING ERRCODE = '23514';
        END IF;
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;

    IF NEW.slug = 'spokes' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent
            WHERE parent.id = NEW.parent_id
              AND parent.slug = 'wheel-components'
              AND parent.parent_id IS NULL
              AND parent.depth = 1
              AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'spokes must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER protect_spokes_system_category_identity_and_tree_position
BEFORE INSERT OR UPDATE OR DELETE ON product_categories
FOR EACH ROW EXECUTE FUNCTION protect_spokes_system_category_identity_and_tree_position();

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
