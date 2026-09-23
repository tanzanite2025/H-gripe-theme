ALTER TABLE payment_operation_idempotencies
    ADD COLUMN IF NOT EXISTS claim_token VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reconciliation_started_at TIMESTAMPTZ;

-- Existing read-only confirmations are immediately reclaimable. Existing
-- PayPal capture claims have an unknown provider outcome and therefore enter
-- reconciliation instead of becoming executable mutations again.
UPDATE payment_operation_idempotencies
SET
    status = CASE
        WHEN scope = 'paypal_capture' THEN 'reconciling'
        ELSE status
    END,
    lease_expires_at = NOW(),
    reconciliation_started_at = CASE
        WHEN scope = 'paypal_capture' THEN COALESCE(reconciliation_started_at, NOW())
        ELSE reconciliation_started_at
    END,
    updated_at = NOW()
WHERE status = 'pending';

ALTER TABLE payment_operation_idempotencies
    DROP CONSTRAINT IF EXISTS ck_payment_operation_idempotencies_status;

ALTER TABLE payment_operation_idempotencies
    ADD CONSTRAINT ck_payment_operation_idempotencies_status
        CHECK (status IN ('pending', 'reconciling', 'completed'));

CREATE INDEX IF NOT EXISTS idx_payment_operation_idempotencies_claimable
    ON payment_operation_idempotencies (status, lease_expires_at, id)
    WHERE status IN ('pending', 'reconciling');
