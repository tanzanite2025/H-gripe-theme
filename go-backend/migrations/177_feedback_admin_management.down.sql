DROP INDEX IF EXISTS idx_feedback_reviewed_at;
DROP INDEX IF EXISTS idx_feedback_replied_at;
DROP INDEX IF EXISTS idx_feedback_page_path;

ALTER TABLE feedback
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS replied_by,
    DROP COLUMN IF EXISTS replied_at,
    DROP COLUMN IF EXISTS reply_content,
    DROP COLUMN IF EXISTS page_title,
    DROP COLUMN IF EXISTS page_path;
