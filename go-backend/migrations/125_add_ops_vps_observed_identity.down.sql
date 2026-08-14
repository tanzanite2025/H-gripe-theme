ALTER TABLE ops_vps_bindings
    DROP COLUMN IF EXISTS observed_operating_system,
    DROP COLUMN IF EXISTS observed_ipv4,
    DROP COLUMN IF EXISTS observed_hostname;
