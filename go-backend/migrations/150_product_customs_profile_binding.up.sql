ALTER TABLE products
    ADD COLUMN IF NOT EXISTS customs_classification_profile_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_products_customs_classification_profile'
    ) THEN
        ALTER TABLE products
            ADD CONSTRAINT fk_products_customs_classification_profile
            FOREIGN KEY (customs_classification_profile_id)
            REFERENCES customs_classification_profiles(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_customs_classification_profile_id
    ON products(customs_classification_profile_id);

UPDATE products p
SET customs_classification_profile_id = matched.profile_id
FROM (
    SELECT
        p2.id AS product_id,
        MAX(profile.id) AS profile_id
    FROM products p2
    JOIN customs_classification_profiles profile
      ON profile.hs_code = COALESCE(p2.hs_code, '')
     AND profile.cn_code = COALESCE(p2.cn_code, '')
     AND profile.country_of_origin = COALESCE(p2.country_of_origin, '')
     AND profile.customs_description = COALESCE(p2.customs_description, '')
     AND (
         profile.product_specification_template_id IS NULL
         OR profile.product_specification_template_id = p2.product_specification_template_id
     )
     AND profile.status = 'active'
    WHERE p2.customs_classification_profile_id IS NULL
    GROUP BY p2.id
    HAVING COUNT(profile.id) = 1
) matched
WHERE p.id = matched.product_id;
