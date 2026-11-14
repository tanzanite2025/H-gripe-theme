-- Stop tracking writers before downgrading. Never discard in-flight ownership
-- or a generation that an outstanding worker could still use as a fencing token.
LOCK TABLE shipping_tracking_shipments IN ACCESS EXCLUSIVE MODE;
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM shipping_tracking_shipments
        WHERE sync_status = 'syncing'
           OR sync_lease_owner <> ''
           OR sync_lease_generation <> 0
           OR sync_lease_expires_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot remove tracking sync leases while syncing shipments or lease generations remain';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_shipping_tracking_shipments_sync_lease_expires_at;
DROP INDEX IF EXISTS idx_shipping_tracking_shipments_sync_lease_generation;
DROP INDEX IF EXISTS idx_shipping_tracking_shipments_sync_lease_owner;
ALTER TABLE shipping_tracking_shipments
    DROP COLUMN sync_lease_expires_at,
    DROP COLUMN sync_lease_generation,
    DROP COLUMN sync_lease_owner;
