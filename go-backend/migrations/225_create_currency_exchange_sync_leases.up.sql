CREATE TABLE IF NOT EXISTS currency_exchange_sync_leases (
    lease_key VARCHAR(80) PRIMARY KEY,
    owner_id VARCHAR(160) NOT NULL,
    lease_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_currency_exchange_sync_leases_expires_at
    ON currency_exchange_sync_leases(lease_expires_at);
