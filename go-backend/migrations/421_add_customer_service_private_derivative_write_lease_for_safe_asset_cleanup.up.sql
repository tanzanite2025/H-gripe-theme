-- Private derivative conversion can overlap expiry or conversation retention.
-- The lease prevents retiring its asset while bounded object writes are active.
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS customer_service_derivative_write_token VARCHAR(36) NOT NULL DEFAULT '';
ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS customer_service_derivative_write_expires_at TIMESTAMPTZ;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='media_assets'::regclass AND conname='media_assets_customer_service_derivative_write_lease_valid') THEN
        ALTER TABLE media_assets ADD CONSTRAINT media_assets_customer_service_derivative_write_lease_valid
            CHECK ((customer_service_derivative_write_token='' AND customer_service_derivative_write_expires_at IS NULL)
                OR (customer_service_derivative_write_token<>'' AND customer_service_derivative_write_expires_at IS NOT NULL));
    END IF;
END $$;
