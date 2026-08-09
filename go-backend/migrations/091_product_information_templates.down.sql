ALTER TABLE products DROP COLUMN IF EXISTS after_sales_template_id;
ALTER TABLE products DROP COLUMN IF EXISTS packaging_template_id;
DROP TABLE IF EXISTS product_information_templates;
