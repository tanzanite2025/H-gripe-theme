ALTER TABLE tracking_events DROP COLUMN IF EXISTS carrier_code;
ALTER TABLE tracking_events ADD COLUMN IF NOT EXISTS provider_carrier_code TEXT;
CREATE INDEX IF NOT EXISTS idx_tracking_events_provider_carrier_code ON tracking_events (provider_carrier_code);
