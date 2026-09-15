-- Customer support requests are informational and may coexist with an
-- operator-managed case (for example while a return shipment is in transit).
-- Replace the historical order-wide lock with a one-active-customer-request
-- guard; item-level eligibility checks continue to prevent duplicate refunds.
DROP INDEX IF EXISTS uq_active_after_sales_case_per_order;

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_customer_after_sales_request_per_order
    ON after_sales_cases(order_id)
    WHERE type = 'customer_request'
      AND status NOT IN ('completed', 'rejected', 'cancelled');
