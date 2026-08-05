CREATE TABLE IF NOT EXISTS refund_line_items (
  id BIGSERIAL PRIMARY KEY,
  refund_id BIGINT NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
  order_id BIGINT NOT NULL REFERENCES orders(id),
  order_item_id BIGINT NOT NULL REFERENCES order_items(id),
  product_id BIGINT NOT NULL,
  variant_id BIGINT NULL,
  product_name TEXT NOT NULL,
  sku TEXT NOT NULL DEFAULT '',
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
  line_subtotal_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  line_tax_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  line_discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  line_total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  restock BOOLEAN NOT NULL DEFAULT FALSE,
  restocked_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refund_line_items_refund_id ON refund_line_items(refund_id);
CREATE INDEX IF NOT EXISTS idx_refund_line_items_order_id ON refund_line_items(order_id);
CREATE INDEX IF NOT EXISTS idx_refund_line_items_order_item_id ON refund_line_items(order_item_id);
