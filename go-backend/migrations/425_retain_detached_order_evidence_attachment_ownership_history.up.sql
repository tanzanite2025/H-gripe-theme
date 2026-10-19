-- Preserve immutable object ownership after removing a mutable placement.
-- Existing attachment rows and snapshots are unchanged.
ALTER TABLE order_evidence_attachments ADD COLUMN IF NOT EXISTS detached_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_order_evidence_attachments_detached_at ON order_evidence_attachments (detached_at);
