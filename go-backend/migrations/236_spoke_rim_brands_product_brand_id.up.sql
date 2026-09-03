ALTER TABLE spoke_rim_brands
    ADD COLUMN IF NOT EXISTS product_brand_id BIGINT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_spoke_rim_brands_product_brand_id
    ON spoke_rim_brands(product_brand_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_spoke_rim_brands_product_brand'
    ) THEN
        ALTER TABLE spoke_rim_brands
            ADD CONSTRAINT fk_spoke_rim_brands_product_brand
            FOREIGN KEY (product_brand_id) REFERENCES product_brands (id)
            ON DELETE RESTRICT;
    END IF;
END $$;
