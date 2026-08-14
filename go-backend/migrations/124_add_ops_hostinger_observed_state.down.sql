DROP INDEX IF EXISTS idx_ops_project_binding_observed_state;

ALTER TABLE ops_project_bindings
    DROP COLUMN IF EXISTS observed_healthy_container_count,
    DROP COLUMN IF EXISTS observed_running_container_count,
    DROP COLUMN IF EXISTS observed_container_count,
    DROP COLUMN IF EXISTS observed_source,
    DROP COLUMN IF EXISTS observed_state;

DROP INDEX IF EXISTS idx_ops_vps_binding_observed_state;

ALTER TABLE ops_vps_bindings
    DROP COLUMN IF EXISTS observed_region,
    DROP COLUMN IF EXISTS observed_plan,
    DROP COLUMN IF EXISTS observed_source,
    DROP COLUMN IF EXISTS observed_state;
