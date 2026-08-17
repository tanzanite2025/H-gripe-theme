ALTER TABLE feedback
    ADD COLUMN IF NOT EXISTS source_hash VARCHAR(80);

CREATE INDEX IF NOT EXISTS idx_feedback_source_hash_created_at ON feedback(source_hash, created_at);
