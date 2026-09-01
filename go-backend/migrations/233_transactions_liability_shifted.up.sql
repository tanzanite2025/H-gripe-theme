ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS liability_shifted BOOLEAN NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_liability_shifted
    ON transactions(liability_shifted)
    WHERE liability_shifted IS NOT NULL;
