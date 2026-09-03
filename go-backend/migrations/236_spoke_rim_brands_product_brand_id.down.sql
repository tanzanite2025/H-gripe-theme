ALTER TABLE spoke_rim_brands
    DROP CONSTRAINT IF EXISTS fk_spoke_rim_brands_product_brand;

DROP INDEX IF EXISTS idx_spoke_rim_brands_product_brand_id;

ALTER TABLE spoke_rim_brands
    DROP COLUMN IF EXISTS product_brand_id;
