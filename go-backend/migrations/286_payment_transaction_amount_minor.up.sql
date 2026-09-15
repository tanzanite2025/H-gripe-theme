ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS amount_minor BIGINT;

UPDATE transactions
SET amount_minor = CASE
        WHEN UPPER(COALESCE(currency, 'USD')) IN ('JPY', 'KRW', 'CLP')
            THEN ROUND(amount)::BIGINT
        ELSE ROUND(amount * 100)::BIGINT
    END
WHERE amount_minor IS NULL;

ALTER TABLE transactions
    ALTER COLUMN amount_minor SET NOT NULL,
    ADD CONSTRAINT chk_transactions_amount_minor_non_negative CHECK (amount_minor >= 0);
