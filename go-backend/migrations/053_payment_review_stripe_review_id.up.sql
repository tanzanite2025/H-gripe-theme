ALTER TABLE payment_reviews
    ADD COLUMN IF NOT EXISTS stripe_review_id VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_payment_reviews_stripe_review_id
    ON payment_reviews(stripe_review_id);
