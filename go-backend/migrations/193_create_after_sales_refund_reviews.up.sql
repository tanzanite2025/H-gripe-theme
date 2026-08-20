-- A review record is a controlled approval bridge for after-sales refunds.
-- It never creates a Refund record or invokes a payment gateway by itself.

CREATE TABLE IF NOT EXISTS after_sales_refund_reviews (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL UNIQUE REFERENCES after_sales_cases(id) ON DELETE CASCADE,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    proposed_amount NUMERIC(18, 2) NOT NULL CHECK (proposed_amount > 0),
    currency VARCHAR(8) NOT NULL,
    request_notes TEXT NOT NULL DEFAULT '',
    decision_notes TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL DEFAULT 0,
    updated_by BIGINT NOT NULL DEFAULT 0,
    reviewed_by_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    linked_refund_id BIGINT REFERENCES refunds(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_after_sales_refund_reviews_status
        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_after_sales_refund_reviews_status_created
    ON after_sales_refund_reviews(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_after_sales_refund_reviews_linked_refund
    ON after_sales_refund_reviews(linked_refund_id);
