CREATE TABLE IF NOT EXISTS after_sales_return_shipments (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES after_sales_cases(id) ON DELETE CASCADE,
    warehouse_name VARCHAR(200) NOT NULL DEFAULT '',
    warehouse_address TEXT NOT NULL DEFAULT '',
    carrier VARCHAR(100) NOT NULL DEFAULT '',
    tracking_number VARCHAR(200) NOT NULL DEFAULT '',
    tracking_url TEXT NOT NULL DEFAULT '',
    label_url TEXT NOT NULL DEFAULT '',
    shipped_at TIMESTAMPTZ NULL,
    received_at TIMESTAMPTZ NULL,
    received_by BIGINT NULL,
    created_by BIGINT NOT NULL DEFAULT 0,
    updated_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_after_sales_return_shipments_case_id_created_at
    ON after_sales_return_shipments(case_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_after_sales_return_shipments_tracking_number
    ON after_sales_return_shipments(tracking_number)
    WHERE tracking_number <> '';
CREATE INDEX IF NOT EXISTS idx_after_sales_return_shipments_received_at
    ON after_sales_return_shipments(received_at)
    WHERE received_at IS NULL;
