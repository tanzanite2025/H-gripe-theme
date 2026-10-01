-- The following migration originally extended the reviewable catalog created
-- by 354. Later migrations intentionally replace it with a candidate-only
-- table, so replaying the historical chain must skip these review columns when
-- that replacement schema is already present.
DO $migration$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'schwalbe_tire_specifications'
          AND column_name = 'verification_status'
    ) THEN
        EXECUTE $statement$
            ALTER TABLE schwalbe_tire_specifications
                ADD COLUMN IF NOT EXISTS source_submitted_by_id BIGINT,
                ADD COLUMN IF NOT EXISTS reviewed_by_id BIGINT,
                ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
                ADD COLUMN IF NOT EXISTS review_note TEXT NOT NULL DEFAULT '',
                ADD COLUMN IF NOT EXISTS review_revision BIGINT NOT NULL DEFAULT 1;

            -- Keep legacy rows readable while requiring review evidence for
            -- rows created or updated after this migration. Existing rows are
            -- not rewritten.
            ALTER TABLE schwalbe_tire_specifications
                ADD CONSTRAINT chk_schwalbe_tire_verified_review_audit
                CHECK (
                    verification_status <> 'verified'
                    OR (
                        reviewed_by_id IS NOT NULL
                        AND reviewed_at IS NOT NULL
                        AND (source_submitted_by_id IS NULL OR reviewed_by_id <> source_submitted_by_id)
                    )
                ) NOT VALID;
        $statement$;
    END IF;
END
$migration$;
