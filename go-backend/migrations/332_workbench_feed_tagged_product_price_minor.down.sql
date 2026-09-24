-- The financial hard cut intentionally does not restore the removed major
-- column. Rollback only removes the canonical column and its constraint.
ALTER TABLE workbench_feed_tagged_products
    DROP CONSTRAINT IF EXISTS chk_workbench_feed_tagged_products_price_minor_non_negative,
    DROP COLUMN IF EXISTS price_minor;
