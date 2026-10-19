DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM order_evidence_attachments WHERE detached_at IS NOT NULL) THEN
        RAISE EXCEPTION 'Cannot discard detached immutable evidence ownership or restore detached placements';
    END IF;
END $$;
DROP INDEX IF EXISTS idx_order_evidence_attachments_detached_at;
ALTER TABLE order_evidence_attachments DROP COLUMN IF EXISTS detached_at;
