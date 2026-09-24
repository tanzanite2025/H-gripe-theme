-- Configuration snapshots are the immutable order-time source for selected
-- options. The free-form attributes column is no longer written or read by
-- production paths and is removed before launch.
ALTER TABLE order_items
    DROP COLUMN IF EXISTS attributes;
