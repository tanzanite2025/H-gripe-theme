-- A receipt and its product increment commit together. There is intentionally
-- no product FK: deleting a product must not delete its replay protection.
CREATE TABLE product_view_count_flush_receipts (
    batch_id UUID NOT NULL,
    product_id BIGINT NOT NULL CHECK (product_id > 0),
    delta BIGINT NOT NULL CHECK (delta > 0),
    product_increment_applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (batch_id, product_id)
);
COMMENT ON TABLE product_view_count_flush_receipts IS
    'Durable buffered view increment receipts. Preserve across partial recovery; never expire independently of Redis processing batches.';
