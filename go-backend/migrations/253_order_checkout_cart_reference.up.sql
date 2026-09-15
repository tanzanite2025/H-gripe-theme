ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS checkout_cart_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_checkout_cart_id
    ON orders(checkout_cart_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_orders_checkout_cart'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT fk_orders_checkout_cart
            FOREIGN KEY (checkout_cart_id)
            REFERENCES carts(id)
            ON DELETE SET NULL;
    END IF;
END $$;
