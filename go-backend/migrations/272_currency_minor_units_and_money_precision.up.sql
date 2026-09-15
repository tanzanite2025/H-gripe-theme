-- Gift-card ledgers are minor units in the card currency. Preserve legacy
-- rows as USD and make the currency explicit for zero-decimal currencies.
ALTER TABLE gift_card_transactions
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD';

UPDATE gift_card_transactions t
SET currency = COALESCE(NULLIF(UPPER(gc.currency), ''), 'USD')
FROM gift_cards gc
WHERE gc.id = t.gift_card_id;

-- Monetary values use exact decimal storage. Go continues to expose float64
-- at the API boundary, but PostgreSQL no longer stores binary IEEE-754 noise.
ALTER TABLE orders
    ALTER COLUMN subtotal_amount TYPE NUMERIC(18,2),
    ALTER COLUMN shipping_fee TYPE NUMERIC(18,2),
    ALTER COLUMN tax_amount TYPE NUMERIC(18,2),
    ALTER COLUMN discount_amount TYPE NUMERIC(18,2),
    ALTER COLUMN total_amount TYPE NUMERIC(18,2),
    ALTER COLUMN points_value TYPE NUMERIC(18,2);

ALTER TABLE currency_exchange_rates
    ALTER COLUMN rate TYPE NUMERIC(20,10);
