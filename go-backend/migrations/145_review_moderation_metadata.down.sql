DROP INDEX IF EXISTS idx_reviews_moderated_by;

ALTER TABLE reviews
    DROP COLUMN IF EXISTS moderation_reason,
    DROP COLUMN IF EXISTS moderated_by,
    DROP COLUMN IF EXISTS moderated_at;
