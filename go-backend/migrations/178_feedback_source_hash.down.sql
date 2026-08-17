DROP INDEX IF EXISTS idx_feedback_source_hash_created_at;

ALTER TABLE feedback
    DROP COLUMN IF EXISTS source_hash;
