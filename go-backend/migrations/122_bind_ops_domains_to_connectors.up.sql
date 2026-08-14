ALTER TABLE ops_domain_bindings
    ADD COLUMN IF NOT EXISTS connector_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_connector
    ON ops_domain_bindings (connector_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_ops_domain_binding_connector'
    ) THEN
        ALTER TABLE ops_domain_bindings
            ADD CONSTRAINT fk_ops_domain_binding_connector
            FOREIGN KEY (connector_id) REFERENCES ops_connectors (id);
    END IF;
END $$;

UPDATE ops_domain_bindings AS domain
SET connector_id = connector.id
FROM ops_connectors AS connector
WHERE domain.connector_id IS NULL
  AND domain.provider = 'cloudflare'
  AND connector.provider = 'cloudflare'
  AND connector.name = 'Cloudflare Production';
