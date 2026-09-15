DROP INDEX IF EXISTS uq_active_customer_after_sales_request_per_order;

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_after_sales_case_per_order
    ON after_sales_cases(order_id)
    WHERE status NOT IN ('completed', 'rejected', 'cancelled');
