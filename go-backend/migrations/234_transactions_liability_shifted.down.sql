DROP INDEX IF EXISTS idx_transactions_liability_shifted;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS liability_shifted;
