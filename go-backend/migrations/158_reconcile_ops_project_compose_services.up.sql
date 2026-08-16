-- Reconcile the persisted production ledger with compose.prod.yml.
UPDATE ops_project_bindings AS project
SET services = array_to_string(
        ARRAY(
            SELECT btrim(service.name)
            FROM unnest(string_to_array(COALESCE(project.services, ''), ',')) AS service(name)
            WHERE btrim(service.name) <> ''
        ) || ARRAY['edge-config'],
        ', '
    ),
    updated_at = NOW()
WHERE project.name = 'commerce-platform'
  AND project.environment = 'production'
  AND NOT EXISTS (
      SELECT 1
      FROM unnest(string_to_array(COALESCE(project.services, ''), ',')) AS service(name)
      WHERE btrim(service.name) = 'edge-config'
  );
