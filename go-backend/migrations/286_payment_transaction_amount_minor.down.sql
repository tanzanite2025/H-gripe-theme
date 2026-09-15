ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS chk_transactions_amount_minor_non_negative;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS amount_minor;
