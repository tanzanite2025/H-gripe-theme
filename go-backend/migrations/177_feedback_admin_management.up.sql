ALTER TABLE feedback
    ADD COLUMN IF NOT EXISTS page_path TEXT,
    ADD COLUMN IF NOT EXISTS page_title TEXT,
    ADD COLUMN IF NOT EXISTS reply_content TEXT,
    ADD COLUMN IF NOT EXISTS replied_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS replied_by BIGINT,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reviewed_by BIGINT;

CREATE INDEX IF NOT EXISTS idx_feedback_page_path ON feedback(page_path);
CREATE INDEX IF NOT EXISTS idx_feedback_replied_at ON feedback(replied_at);
CREATE INDEX IF NOT EXISTS idx_feedback_reviewed_at ON feedback(reviewed_at);
