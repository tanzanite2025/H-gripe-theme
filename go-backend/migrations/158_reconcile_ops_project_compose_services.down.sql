UPDATE ops_project_bindings AS project
SET services = COALESCE(
        array_to_string(
            ARRAY(
                SELECT btrim(service.name)
                FROM unnest(string_to_array(COALESCE(project.services, ''), ',')) WITH ORDINALITY AS service(name, ordinal)
                WHERE btrim(service.name) <> ''
                  AND btrim(service.name) <> 'edge-config'
                ORDER BY service.ordinal
            ),
            ', '
        ),
        ''
    ),
    updated_at = NOW()
WHERE project.name = 'commerce-platform'
  AND project.environment = 'production'
  AND EXISTS (
      SELECT 1
      FROM unnest(string_to_array(COALESCE(project.services, ''), ',')) AS service(name)
      WHERE btrim(service.name) = 'edge-config'
  );
