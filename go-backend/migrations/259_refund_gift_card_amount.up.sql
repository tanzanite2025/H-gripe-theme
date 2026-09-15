ALTER TABLE refunds
    ADD COLUMN IF NOT EXISTS gift_card_refund_amount NUMERIC(12,2) NOT NULL DEFAULT 0;

ALTER TABLE gift_card_transactions
    ADD COLUMN IF NOT EXISTS refund_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conname = 'gift_card_transactions_refund_fk'
    ) THEN
        ALTER TABLE gift_card_transactions
            ADD CONSTRAINT gift_card_transactions_refund_fk
            FOREIGN KEY (refund_id) REFERENCES refunds(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_refund_id
    ON gift_card_transactions(refund_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gift_card_transactions_refund_card
    ON gift_card_transactions(refund_id, gift_card_id)
    WHERE type = 'refund' AND refund_id IS NOT NULL;
