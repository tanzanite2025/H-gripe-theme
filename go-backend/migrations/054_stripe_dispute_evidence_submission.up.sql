ALTER TABLE stripe_disputes
    ADD COLUMN IF NOT EXISTS evidence_submitted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS evidence_submission_payload TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS evidence_submission_error TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_stripe_disputes_evidence_submitted_at
    ON stripe_disputes(evidence_submitted_at);
