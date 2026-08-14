ALTER TABLE ops_project_bindings
    DROP CONSTRAINT IF EXISTS fk_ops_project_binding_connector;

ALTER TABLE ops_vps_bindings
    DROP CONSTRAINT IF EXISTS fk_ops_vps_binding_connector;
