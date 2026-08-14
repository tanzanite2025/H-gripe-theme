ALTER TABLE ops_domain_bindings
    ADD COLUMN IF NOT EXISTS project_binding_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_ops_domain_binding_project
    ON ops_domain_bindings (project_binding_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_ops_domain_binding_project'
          AND conrelid = 'ops_domain_bindings'::regclass
    ) THEN
        ALTER TABLE ops_domain_bindings
            ADD CONSTRAINT fk_ops_domain_binding_project
            FOREIGN KEY (project_binding_id) REFERENCES ops_project_bindings (id)
            ON DELETE SET NULL;
    END IF;
END $$;

-- Only the original commerce-platform seed has a deterministic legacy owner.
-- Other same-environment projects remain unbound and are reported as REVIEW
-- until an operator assigns them explicitly.
UPDATE ops_domain_bindings AS domain
SET project_binding_id = project.id
FROM ops_project_bindings AS project
WHERE domain.project_binding_id IS NULL
  AND domain.environment = project.environment
  AND project.name = 'commerce-platform'
  AND (
      domain.target = project.gateway_alias
      OR domain.target LIKE project.gateway_alias || ':%'
      OR domain.target = project.compose_project_name
      OR domain.target LIKE project.compose_project_name || ':%'
  );
