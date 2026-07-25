-- Unified visitor profile source for storefront anonymous context.
--
-- This table does not replace user accounts, carts, or customer-service
-- conversations. It only binds browser-scoped facts that are otherwise split
-- across HttpOnly cookies:
--   - Public Chat visitor hash
--   - cart session id
--   - optional captured email
--   - locale / coarse region
--   - privacy-preserving request hashes

CREATE TABLE IF NOT EXISTS visitor_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NULL,
    customer_service_visitor_hash VARCHAR(64) NULL,
    cart_session_id VARCHAR(64) NULL,
    email VARCHAR(255) NULL,
    email_source VARCHAR(40) NULL,
    locale VARCHAR(20) NULL,
    locale_source VARCHAR(40) NULL,
    country_code VARCHAR(8) NULL,
    region VARCHAR(80) NULL,
    city VARCHAR(80) NULL,
    timezone VARCHAR(80) NULL,
    ip_hash VARCHAR(64) NULL,
    user_agent_hash VARCHAR(64) NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_visitor_profiles_customer_service_hash_live
    ON visitor_profiles (customer_service_visitor_hash)
    WHERE deleted_at IS NULL AND customer_service_visitor_hash IS NOT NULL AND customer_service_visitor_hash <> '';

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_cart_session_live
    ON visitor_profiles (cart_session_id)
    WHERE deleted_at IS NULL AND cart_session_id IS NOT NULL AND cart_session_id <> '';

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_user_live
    ON visitor_profiles (user_id)
    WHERE deleted_at IS NULL AND user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_email_live
    ON visitor_profiles (email)
    WHERE deleted_at IS NULL AND email IS NOT NULL AND email <> '';

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_country_live
    ON visitor_profiles (country_code)
    WHERE deleted_at IS NULL AND country_code IS NOT NULL AND country_code <> '';

CREATE INDEX IF NOT EXISTS idx_visitor_profiles_last_seen_live
    ON visitor_profiles (last_seen_at)
    WHERE deleted_at IS NULL;
