ALTER TABLE customs_classification_profiles
    DROP COLUMN IF EXISTS source_url_us,
    DROP COLUMN IF EXISTS source_url_eu,
    DROP COLUMN IF EXISTS source_url_uk;
