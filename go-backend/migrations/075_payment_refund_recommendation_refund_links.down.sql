DROP INDEX IF EXISTS idx_payment_refund_recommendations_refund_created_by;
DROP INDEX IF EXISTS idx_payment_refund_recommendations_linked_refund;

ALTER TABLE payment_refund_recommendations
    DROP COLUMN IF EXISTS refund_created_at,
    DROP COLUMN IF EXISTS refund_created_by_id,
    DROP COLUMN IF EXISTS linked_refund_id;
