-- Verification metadata and built-in catalog rows are shared master data. Do not
-- delete or restore old classifications automatically during a rollback.
DROP INDEX IF EXISTS idx_customs_classification_profiles_review_due_at;

ALTER TABLE customs_classification_profiles
    DROP COLUMN IF EXISTS review_due_at,
    DROP COLUMN IF EXISTS verified_at;

