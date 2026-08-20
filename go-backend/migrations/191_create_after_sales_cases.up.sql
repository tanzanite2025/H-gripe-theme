CREATE TABLE IF NOT EXISTS after_sales_cases (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'requested',
    reason TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    resolution TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL DEFAULT 0,
    updated_by BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS after_sales_case_items (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES after_sales_cases(id) ON DELETE CASCADE,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id BIGINT NOT NULL REFERENCES order_items(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL,
    variant_id BIGINT NULL,
    product_name VARCHAR(255) NOT NULL,
    sku VARCHAR(255) NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_after_sales_case_item UNIQUE (case_id, order_item_id)
);

CREATE INDEX IF NOT EXISTS idx_after_sales_cases_order_id
    ON after_sales_cases(order_id);
CREATE INDEX IF NOT EXISTS idx_after_sales_cases_status
    ON after_sales_cases(status);
CREATE INDEX IF NOT EXISTS idx_after_sales_cases_type
    ON after_sales_cases(type);
CREATE INDEX IF NOT EXISTS idx_after_sales_case_items_order_item_id
    ON after_sales_case_items(order_item_id);
