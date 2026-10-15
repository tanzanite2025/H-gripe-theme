DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM media_assets WHERE customer_service_derivative_write_token<>'' OR customer_service_derivative_write_expires_at IS NOT NULL) THEN
        RAISE EXCEPTION 'Cannot remove customer-service derivative write leases while leased assets remain';
    END IF;
END $$;
ALTER TABLE media_assets DROP CONSTRAINT IF EXISTS media_assets_customer_service_derivative_write_lease_valid;
ALTER TABLE media_assets DROP COLUMN IF EXISTS customer_service_derivative_write_expires_at;
ALTER TABLE media_assets DROP COLUMN IF EXISTS customer_service_derivative_write_token;
