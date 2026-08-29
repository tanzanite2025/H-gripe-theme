-- Reconcile the persisted production ledger with the dedicated API ingress
-- network introduced by compose.prod.yml.
UPDATE ops_project_bindings AS project
SET networks = 'db, cache, app, api_ingress, shared-edge',
    updated_at = NOW()
WHERE project.name = 'commerce-platform'
  AND project.environment = 'production'
  AND btrim(COALESCE(project.networks, '')) = 'db, cache, app, shared-edge';
