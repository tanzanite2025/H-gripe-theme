-- Redis may still replay an applied batch. An empty ledger alone is the only
-- supported rollback; operators must also stop writers and reconcile Redis.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM product_view_count_flush_receipts) THEN
        RAISE EXCEPTION 'Cannot remove product view flush replay protection while durable receipts exist';
    END IF;
END $$;
DROP TABLE product_view_count_flush_receipts;
