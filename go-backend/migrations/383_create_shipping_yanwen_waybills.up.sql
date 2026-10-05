CREATE TABLE IF NOT EXISTS shipping_yanwen_waybills (
    id BIGSERIAL PRIMARY KEY,
    environment VARCHAR(16) NOT NULL,
    order_id BIGINT NOT NULL,
    order_number VARCHAR(80) NOT NULL,
    product_code VARCHAR(80) NOT NULL,
    channel_name VARCHAR(160) NOT NULL DEFAULT '',
    warehouse_code VARCHAR(80) NOT NULL,
    destination_country VARCHAR(16) NOT NULL,
    consignee_name VARCHAR(200) NOT NULL,
    declared_description VARCHAR(255) NOT NULL,
    total_quantity INTEGER NOT NULL,
    total_weight_grams INTEGER NOT NULL,
    waybill_number VARCHAR(100) NOT NULL,
    yanwen_order_number VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    request_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_shipping_yanwen_waybills_order_product UNIQUE (order_id, product_code),
    CONSTRAINT uq_shipping_yanwen_waybills_waybill_number UNIQUE (waybill_number)
);

CREATE INDEX IF NOT EXISTS idx_shipping_yanwen_waybills_environment_status
    ON shipping_yanwen_waybills (environment, status);

CREATE INDEX IF NOT EXISTS idx_shipping_yanwen_waybills_order_number
    ON shipping_yanwen_waybills (order_number);
