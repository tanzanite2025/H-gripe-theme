DROP INDEX IF EXISTS uq_gift_card_transactions_refund_card;
DROP INDEX IF EXISTS idx_gift_card_transactions_refund_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conname = 'gift_card_transactions_refund_fk'
    ) THEN
        ALTER TABLE gift_card_transactions
            DROP CONSTRAINT gift_card_transactions_refund_fk;
    END IF;
END $$;

ALTER TABLE gift_card_transactions
    DROP COLUMN IF EXISTS refund_id;

ALTER TABLE refunds
    DROP COLUMN IF EXISTS gift_card_refund_amount;
