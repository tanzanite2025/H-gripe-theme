DROP INDEX IF EXISTS uq_transactions_order_method_attempt;
DROP INDEX IF EXISTS idx_transactions_attempt_key;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS provider_request_key,
    DROP COLUMN IF EXISTS attempt_key;
