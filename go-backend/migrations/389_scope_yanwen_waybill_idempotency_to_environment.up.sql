ALTER TABLE shipping_yanwen_waybills
    DROP CONSTRAINT IF EXISTS uq_shipping_yanwen_waybills_order_product;

DROP INDEX IF EXISTS idx_shipping_yanwen_waybill_order_channel;

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipping_yanwen_waybill_environment_order_product
    ON shipping_yanwen_waybills (environment, order_id, product_code);
