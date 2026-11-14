-- Match TrackingShipment's durable worker ownership and stale-writer fencing.
-- Existing syncing rows remain unchanged and are claimable without a lease.
ALTER TABLE shipping_tracking_shipments
    ADD COLUMN IF NOT EXISTS sync_lease_owner VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sync_lease_generation BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sync_lease_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_sync_lease_owner
    ON shipping_tracking_shipments (sync_lease_owner);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_sync_lease_generation
    ON shipping_tracking_shipments (sync_lease_generation);
CREATE INDEX IF NOT EXISTS idx_shipping_tracking_shipments_sync_lease_expires_at
    ON shipping_tracking_shipments (sync_lease_expires_at);
