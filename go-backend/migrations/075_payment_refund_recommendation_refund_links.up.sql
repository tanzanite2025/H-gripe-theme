ALTER TABLE payment_refund_recommendations
    ADD COLUMN IF NOT EXISTS linked_refund_id BIGINT REFERENCES refunds(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS refund_created_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS refund_created_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_linked_refund
    ON payment_refund_recommendations(linked_refund_id);

CREATE INDEX IF NOT EXISTS idx_payment_refund_recommendations_refund_created_by
    ON payment_refund_recommendations(refund_created_by_id);
