ALTER TABLE ops_vps_bindings
    ADD COLUMN IF NOT EXISTS observed_state VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_source VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_plan VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_region VARCHAR(120) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_ops_vps_binding_observed_state
    ON ops_vps_bindings (observed_state);

ALTER TABLE ops_project_bindings
    ADD COLUMN IF NOT EXISTS observed_state VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_source VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS observed_container_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS observed_running_container_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS observed_healthy_container_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_ops_project_binding_observed_state
    ON ops_project_bindings (observed_state);
