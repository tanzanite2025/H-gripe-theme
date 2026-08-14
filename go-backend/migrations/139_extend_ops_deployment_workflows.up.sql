ALTER TABLE ops_deployment_workflow_runs
    ADD COLUMN IF NOT EXISTS previous_ref VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS rollback_ref VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS remote_operation_id VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS health_status VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS health_snapshot TEXT NOT NULL DEFAULT '';

ALTER TABLE ops_deployment_workflow_steps
    ADD COLUMN IF NOT EXISTS external_operation_id VARCHAR(160) NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS ops_deployment_workflow_locks (
    resource_key VARCHAR(160) PRIMARY KEY,
    workflow_run_id BIGINT NOT NULL,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_ops_deployment_workflow_lock_run
        FOREIGN KEY (workflow_run_id) REFERENCES ops_deployment_workflow_runs (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ops_deployment_workflow_lock_expires_at
    ON ops_deployment_workflow_locks (expires_at);
