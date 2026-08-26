CREATE TABLE IF NOT EXISTS shipment_records (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    order_number VARCHAR(255) NOT NULL,
    user_id BIGINT,
    customer_name VARCHAR(255),
    customer_email VARCHAR(190),
    tracking_shipment_id BIGINT,
    tracking_number VARCHAR(120),
    shipped_at TIMESTAMPTZ NOT NULL,
    items_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
    product_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    details_bound BOOLEAN NOT NULL DEFAULT FALSE,
    shipping_note TEXT,
    shipping_images JSONB NOT NULL DEFAULT '[]'::jsonb,
    warranty_months INTEGER NOT NULL DEFAULT 12,
    warranty_start_at TIMESTAMPTZ NOT NULL,
    warranty_expires TIMESTAMPTZ NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_shipment_records_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_records_order
    ON shipment_records(order_id);
CREATE INDEX IF NOT EXISTS idx_shipment_records_order_number
    ON shipment_records(order_number);
CREATE INDEX IF NOT EXISTS idx_shipment_records_user_id
    ON shipment_records(user_id);
CREATE INDEX IF NOT EXISTS idx_shipment_records_customer_email
    ON shipment_records(customer_email);
CREATE INDEX IF NOT EXISTS idx_shipment_records_tracking_number
    ON shipment_records(tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipment_records_shipped_at
    ON shipment_records(shipped_at);
CREATE INDEX IF NOT EXISTS idx_shipment_records_warranty_expires
    ON shipment_records(warranty_expires);
CREATE INDEX IF NOT EXISTS idx_shipment_records_status
    ON shipment_records(status);
