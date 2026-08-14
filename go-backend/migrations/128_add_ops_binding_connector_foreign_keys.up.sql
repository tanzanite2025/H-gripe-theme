UPDATE ops_vps_bindings AS vps
SET connector_id = NULL
WHERE connector_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM ops_connectors AS connector
      WHERE connector.id = vps.connector_id
  );

UPDATE ops_project_bindings AS project
SET connector_id = NULL
WHERE connector_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM ops_connectors AS connector
      WHERE connector.id = project.connector_id
  );

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_ops_vps_binding_connector'
          AND conrelid = 'ops_vps_bindings'::regclass
    ) THEN
        ALTER TABLE ops_vps_bindings
            ADD CONSTRAINT fk_ops_vps_binding_connector
            FOREIGN KEY (connector_id) REFERENCES ops_connectors (id)
            ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_ops_project_binding_connector'
          AND conrelid = 'ops_project_bindings'::regclass
    ) THEN
        ALTER TABLE ops_project_bindings
            ADD CONSTRAINT fk_ops_project_binding_connector
            FOREIGN KEY (connector_id) REFERENCES ops_connectors (id)
            ON DELETE SET NULL;
    END IF;
END $$;
