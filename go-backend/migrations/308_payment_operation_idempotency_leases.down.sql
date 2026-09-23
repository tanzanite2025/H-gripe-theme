DROP INDEX IF EXISTS idx_payment_operation_idempotencies_claimable;

UPDATE payment_operation_idempotencies
SET
    status = 'pending',
    updated_at = NOW()
WHERE status = 'reconciling';

ALTER TABLE payment_operation_idempotencies
    DROP CONSTRAINT IF EXISTS ck_payment_operation_idempotencies_status;

ALTER TABLE payment_operation_idempotencies
    DROP COLUMN IF EXISTS reconciliation_started_at,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS claim_token;
