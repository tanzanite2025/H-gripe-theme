DO $$
BEGIN
    IF EXISTS (
        SELECT order_id, product_code
        FROM shipping_yanwen_waybills
        GROUP BY order_id, product_code
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot remove Yanwen waybill environment idempotency while cross-environment duplicates exist';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_shipping_yanwen_waybill_environment_order_product;

ALTER TABLE shipping_yanwen_waybills
    ADD CONSTRAINT uq_shipping_yanwen_waybills_order_product UNIQUE (order_id, product_code);
