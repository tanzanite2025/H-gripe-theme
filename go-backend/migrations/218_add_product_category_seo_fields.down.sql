ALTER TABLE product_category_translations
    DROP COLUMN IF EXISTS seo_intro,
    DROP COLUMN IF EXISTS meta_description,
    DROP COLUMN IF EXISTS meta_title;

ALTER TABLE product_categories
    DROP COLUMN IF EXISTS seo_intro,
    DROP COLUMN IF EXISTS meta_description,
    DROP COLUMN IF EXISTS meta_title;
