CREATE INDEX IF NOT EXISTS idx_orders_payment_expiration_scan
    ON orders(status, payment_status, payment_method, created_at);

CREATE INDEX IF NOT EXISTS idx_transactions_order_status_updated
    ON transactions(order_id, status, updated_at);
