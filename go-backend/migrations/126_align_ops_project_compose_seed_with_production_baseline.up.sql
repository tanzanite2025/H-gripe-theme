UPDATE ops_project_bindings
SET services = 'db, redis, migrate, edge-config, api, storefront, admin, web'
WHERE name = 'commerce-platform'
  AND environment = 'production';
