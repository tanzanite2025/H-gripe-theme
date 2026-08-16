ALTER TABLE products
    DROP COLUMN IF EXISTS customs_description,
    DROP COLUMN IF EXISTS country_of_origin,
    DROP COLUMN IF EXISTS cn_code,
    DROP COLUMN IF EXISTS hs_code;
