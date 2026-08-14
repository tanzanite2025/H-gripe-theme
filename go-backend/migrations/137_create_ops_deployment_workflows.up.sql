CREATE TABLE IF NOT EXISTS ops_deployment_workflow_runs (
    id BIGSERIAL PRIMARY KEY,
    kind VARCHAR(32) NOT NULL DEFAULT 'deployment',
    mode VARCHAR(32) NOT NULL DEFAULT 'dry_run',
    project_id BIGINT NOT NULL,
    project_name VARCHAR(120) NOT NULL DEFAULT '',
    environment VARCHAR(32) NOT NULL DEFAULT '',
    requested_ref VARCHAR(160) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    preflight_status VARCHAR(32) NOT NULL DEFAULT '',
    preflight_snapshot TEXT NOT NULL DEFAULT '',
    created_by_id BIGINT NOT NULL DEFAULT 0,
    created_by VARCHAR(160) NOT NULL DEFAULT '',
    approved_by_id BIGINT NULL,
    approved_by VARCHAR(160) NOT NULL DEFAULT '',
    approved_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ops_deployment_workflow_project
        FOREIGN KEY (project_id) REFERENCES ops_project_bindings (id)
);

CREATE INDEX IF NOT EXISTS idx_ops_deployment_workflow_project
    ON ops_deployment_workflow_runs (project_id);
CREATE INDEX IF NOT EXISTS idx_ops_deployment_workflow_status
    ON ops_deployment_workflow_runs (status);
CREATE INDEX IF NOT EXISTS idx_ops_deployment_workflow_created_at
    ON ops_deployment_workflow_runs (created_at DESC);

CREATE TABLE IF NOT EXISTS ops_deployment_workflow_steps (
    id BIGSERIAL PRIMARY KEY,
    workflow_run_id BIGINT NOT NULL,
    sequence INTEGER NOT NULL,
    key VARCHAR(64) NOT NULL,
    label VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    retryable BOOLEAN NOT NULL DEFAULT FALSE,
    external_effect BOOLEAN NOT NULL DEFAULT FALSE,
    input_snapshot TEXT NOT NULL DEFAULT '',
    output_summary TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ops_deployment_workflow_step_run
        FOREIGN KEY (workflow_run_id) REFERENCES ops_deployment_workflow_runs (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ops_deployment_workflow_step_sequence
    ON ops_deployment_workflow_steps (workflow_run_id, sequence);
