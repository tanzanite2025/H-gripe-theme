LOCK TABLE product_categories, product_category_translations, products, quick_buy_step_product_categories
    IN SHARE ROW EXCLUSIVE MODE;

DO $$
DECLARE
    spokes_category_id BIGINT;
BEGIN
    SELECT id INTO spokes_category_id
    FROM product_categories
    WHERE slug = 'spokes';

    IF spokes_category_id IS NULL THEN
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM products WHERE product_category_id = spokes_category_id)
       OR EXISTS (SELECT 1 FROM product_categories WHERE parent_id = spokes_category_id)
       OR EXISTS (SELECT 1 FROM quick_buy_step_product_categories WHERE product_category_id = spokes_category_id) THEN
        RAISE EXCEPTION 'Cannot remove the spokes system category while products, child categories or Quick Buy configuration use it';
    END IF;
END $$;

DROP TRIGGER reject_product_binding_to_unimplemented_spokes_category ON products;
DROP FUNCTION reject_product_binding_to_unimplemented_spokes_category();
DROP TRIGGER protect_spokes_system_category_identity_and_tree_position ON product_categories;
DROP FUNCTION protect_spokes_system_category_identity_and_tree_position();

DELETE FROM product_category_translations
WHERE product_category_id IN (SELECT id FROM product_categories WHERE slug = 'spokes');

DELETE FROM product_categories WHERE slug = 'spokes';
