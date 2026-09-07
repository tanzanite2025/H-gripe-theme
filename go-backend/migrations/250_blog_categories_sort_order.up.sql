ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_categories_sort_order
    ON categories(locale, sort_order, name, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_post_categories_post_category
    ON post_categories(post_id, category_id);
