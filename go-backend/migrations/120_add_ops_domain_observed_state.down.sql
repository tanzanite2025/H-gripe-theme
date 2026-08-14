DROP INDEX IF EXISTS idx_ops_domain_binding_observed_status;

ALTER TABLE ops_domain_bindings
    DROP COLUMN IF EXISTS observed_status,
    DROP COLUMN IF EXISTS observed_target,
    DROP COLUMN IF EXISTS observed_proxy_mode,
    DROP COLUMN IF EXISTS observed_tls_mode,
    DROP COLUMN IF EXISTS observed_source,
    DROP COLUMN IF EXISTS last_observed_at,
    DROP COLUMN IF EXISTS observed_error;
