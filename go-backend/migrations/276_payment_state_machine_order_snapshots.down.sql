DROP INDEX IF EXISTS idx_orders_payment_currency;

ALTER TABLE orders
    DROP COLUMN IF EXISTS dispute_previous_hold,
    DROP COLUMN IF EXISTS dispute_previous_status,
    DROP COLUMN IF EXISTS payment_amount,
    DROP COLUMN IF EXISTS payment_currency;
