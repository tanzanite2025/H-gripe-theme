-- The financial hard cut intentionally does not restore the removed major
-- column. Rollback only removes the canonical column and its constraint.
ALTER TABLE order_items
    DROP CONSTRAINT IF EXISTS chk_order_items_declared_value_minor_non_negative,
    DROP COLUMN IF EXISTS declared_value_minor;
