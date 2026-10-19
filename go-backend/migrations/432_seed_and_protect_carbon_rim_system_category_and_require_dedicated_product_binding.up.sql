LOCK TABLE product_categories, products IN SHARE ROW EXCLUSIVE MODE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM product_categories
        WHERE slug = 'wheel-components' AND parent_id IS NULL AND depth = 1 AND is_enabled
    ) THEN
        RAISE EXCEPTION 'The enabled wheel-components root category is required before adding the carbon-rim system category';
    END IF;
    IF EXISTS (
        SELECT 1 FROM products p JOIN product_categories c ON c.id = p.product_category_id
        WHERE c.slug = 'carbon-rim' AND p.product_type <> 'carbon_rim'
    ) THEN
        RAISE EXCEPTION 'Non-carbon-rim products already use carbon-rim; explicitly classify these products before installing the dedicated system category';
    END IF;
END $$;

INSERT INTO product_categories (parent_id, name, slug, description, depth, sort_order, is_enabled, created_at, updated_at)
SELECT id, 'Carbon Rim', 'carbon-rim', 'Carbon fiber bicycle rims with dedicated engineering specifications and sales SKUs.', 2, 30, TRUE, NOW(), NOW()
FROM product_categories WHERE slug = 'wheel-components'
ON CONFLICT (slug) DO UPDATE SET
    parent_id = EXCLUDED.parent_id, depth = 2, is_enabled = TRUE, updated_at = NOW();

INSERT INTO product_category_translations (product_category_id, locale, name, description, created_at, updated_at)
SELECT category.id, translation.locale, translation.name, translation.description, NOW(), NOW()
FROM product_categories category
CROSS JOIN (VALUES
    ('en', 'Carbon Rim', 'Carbon fiber bicycle rims with dedicated engineering specifications and sales SKUs.'),
    ('zh_cn', '碳纤维车圈', '具有独立工程规格和销售 SKU 的碳纤维自行车车圈。')
) AS translation(locale, name, description)
WHERE category.slug = 'carbon-rim'
ON CONFLICT (product_category_id, locale) DO NOTHING;

-- Classify all dedicated products, including translations and retired records,
-- once. Ordinary products are never converted by their name or category label.
UPDATE products
SET product_category_id = (SELECT id FROM product_categories WHERE slug = 'carbon-rim'), updated_at = NOW()
WHERE product_type = 'carbon_rim';

CREATE FUNCTION protect_carbon_rim_system_category_identity_and_tree_position()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.slug IN ('wheel-components', 'carbon-rim') THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'System category % cannot be deleted', OLD.slug USING ERRCODE = '23514';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR NEW.slug IS DISTINCT FROM OLD.slug THEN
            RAISE EXCEPTION 'System category % identity cannot be changed', OLD.slug USING ERRCODE = '23514';
        END IF;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    IF NEW.slug = 'wheel-components' AND (NEW.parent_id IS NOT NULL OR NEW.depth <> 1 OR NOT NEW.is_enabled) THEN
        RAISE EXCEPTION 'wheel-components must remain an enabled root category' USING ERRCODE = '23514';
    END IF;
    IF NEW.slug = 'carbon-rim' AND (
        NEW.depth <> 2 OR NOT NEW.is_enabled OR NOT EXISTS (
            SELECT 1 FROM product_categories parent WHERE parent.id = NEW.parent_id
            AND parent.slug = 'wheel-components' AND parent.parent_id IS NULL AND parent.depth = 1 AND parent.is_enabled
        )
    ) THEN
        RAISE EXCEPTION 'carbon-rim must remain enabled directly under wheel-components' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END $$;

CREATE TRIGGER protect_carbon_rim_system_category_identity_and_tree_position
BEFORE INSERT OR UPDATE OR DELETE ON product_categories
FOR EACH ROW EXECUTE FUNCTION protect_carbon_rim_system_category_identity_and_tree_position();

CREATE FUNCTION require_carbon_rim_product_type_and_system_category_binding()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    category_slug TEXT;
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.product_type IS DISTINCT FROM OLD.product_type
       AND (NEW.product_type = 'carbon_rim' OR OLD.product_type = 'carbon_rim') THEN
        RAISE EXCEPTION 'Dedicated carbon_rim product type is immutable' USING ERRCODE = '23514';
    END IF;
    SELECT slug INTO category_slug FROM product_categories WHERE id = NEW.product_category_id;
    IF NEW.product_type = 'carbon_rim' AND category_slug IS DISTINCT FROM 'carbon-rim' THEN
        RAISE EXCEPTION 'carbon_rim products must belong to the carbon-rim system category' USING ERRCODE = '23514';
    END IF;
    IF NEW.product_type <> 'carbon_rim' AND category_slug = 'carbon-rim' THEN
        RAISE EXCEPTION 'carbon-rim is reserved for dedicated carbon_rim products' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END $$;

CREATE TRIGGER require_carbon_rim_product_type_and_system_category_binding
BEFORE INSERT OR UPDATE OF product_type, product_category_id ON products
FOR EACH ROW EXECUTE FUNCTION require_carbon_rim_product_type_and_system_category_binding();
