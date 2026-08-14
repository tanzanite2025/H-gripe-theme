ALTER TABLE ops_domain_bindings
    DROP CONSTRAINT IF EXISTS fk_ops_domain_binding_project;

DROP INDEX IF EXISTS idx_ops_domain_binding_project;

ALTER TABLE ops_domain_bindings
    DROP COLUMN IF EXISTS project_binding_id;
