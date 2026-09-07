DROP INDEX IF EXISTS idx_post_categories_post_category;
DROP INDEX IF EXISTS idx_categories_sort_order;
ALTER TABLE categories DROP COLUMN IF EXISTS sort_order;
