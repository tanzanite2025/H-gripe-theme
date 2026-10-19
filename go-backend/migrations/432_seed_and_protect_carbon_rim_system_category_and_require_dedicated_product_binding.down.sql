DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM products p JOIN product_categories c ON c.id = p.product_category_id
        WHERE c.slug = 'carbon-rim'
    ) OR EXISTS (
        SELECT 1 FROM product_categories child JOIN product_categories parent ON parent.id = child.parent_id
        WHERE parent.slug = 'carbon-rim'
    ) OR EXISTS (
        SELECT 1 FROM quick_buy_step_product_categories binding JOIN product_categories category ON category.id = binding.product_category_id
        WHERE category.slug = 'carbon-rim'
    ) THEN
        RAISE EXCEPTION 'Cannot remove the carbon-rim system category while products, child categories or Quick Buy configuration use it';
    END IF;
END $$;

DROP TRIGGER require_carbon_rim_product_type_and_system_category_binding ON products;
DROP FUNCTION require_carbon_rim_product_type_and_system_category_binding();
DROP TRIGGER protect_carbon_rim_system_category_identity_and_tree_position ON product_categories;
DROP FUNCTION protect_carbon_rim_system_category_identity_and_tree_position();
DELETE FROM product_category_translations WHERE product_category_id IN (SELECT id FROM product_categories WHERE slug = 'carbon-rim');
DELETE FROM product_categories WHERE slug = 'carbon-rim';
