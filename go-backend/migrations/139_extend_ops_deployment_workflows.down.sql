DROP TABLE IF EXISTS ops_deployment_workflow_locks;

ALTER TABLE ops_deployment_workflow_steps
    DROP COLUMN IF EXISTS external_operation_id;

ALTER TABLE ops_deployment_workflow_runs
    DROP COLUMN IF EXISTS previous_ref,
    DROP COLUMN IF EXISTS rollback_ref,
    DROP COLUMN IF EXISTS remote_operation_id,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS health_status,
    DROP COLUMN IF EXISTS health_snapshot;
