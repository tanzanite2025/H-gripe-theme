ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS moderated_by BIGINT,
    ADD COLUMN IF NOT EXISTS moderation_reason TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_reviews_moderated_by
    ON reviews(moderated_by);
