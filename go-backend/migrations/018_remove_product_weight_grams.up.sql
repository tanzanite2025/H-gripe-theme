-- Product weight must be maintained only on product_variants.weight_grams.
-- Drop the old hidden product-level source so admin forms, APIs and database schema cannot disagree.
ALTER TABLE products DROP COLUMN IF EXISTS weight_grams;
