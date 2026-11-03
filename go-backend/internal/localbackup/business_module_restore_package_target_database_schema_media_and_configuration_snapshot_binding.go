package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Snapshot binding records authentication and current target observations; domain restore authorization remains with each executor. JSON field order and digest inputs are preserved.
type BusinessModuleRestoreSnapshotBinding struct {
	Version                   int       `json:"version"`
	TaskID                    string    `json:"task_id"`
	PackageID                 string    `json:"package_id"`
	SiteID                    string    `json:"site_id"`
	Environment               string    `json:"environment"`
	ApplicationVersion        string    `json:"application_version"`
	MigrationVersion          int64     `json:"migration_version"`
	ManifestSHA256            string    `json:"manifest_sha256"`
	CiphertextSHA256          string    `json:"ciphertext_sha256"`
	TargetDatabaseOID         int64     `json:"target_database_oid"`
	TargetDatabaseFingerprint string    `json:"target_database_fingerprint"`
	TargetSchemaSHA256        string    `json:"target_schema_sha256"`
	TargetMediaFingerprint    string    `json:"target_media_fingerprint"`
	ConfigurationSHA256       string    `json:"configuration_sha256"`
	WorkCopyExpiresAt         time.Time `json:"work_copy_expires_at"`
}

func (manager *Manager) captureBusinessModuleRestoreSnapshotBinding(ctx context.Context, live *sql.DB, task *Task, mediaFingerprint string) (*BusinessModuleRestoreSnapshotBinding, error) {
	cfg := manager.Configuration
	binding := &BusinessModuleRestoreSnapshotBinding{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: cfg.ExpectedMigrationVersion, CiphertextSHA256: task.CiphertextSHA256, TargetMediaFingerprint: mediaFingerprint, WorkCopyExpiresAt: task.ExpiresAt}
	var err error
	binding.ManifestSHA256, err = businessModuleRecoveryCanonicalJSONSHA256(task.Manifest)
	if err != nil {
		return nil, err
	}
	binding.ConfigurationSHA256, err = businessModuleRecoveryCanonicalJSONSHA256(cfg)
	if err != nil {
		return nil, err
	}
	if err := live.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=current_database()`).Scan(&binding.TargetDatabaseOID); err != nil {
		return nil, err
	}
	tx, err := live.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return nil, err
	}
	if observation.Migration != cfg.ExpectedMigrationVersion {
		return nil, errors.New("目标数据库迁移版本已变化，请重新预览")
	}
	binding.TargetDatabaseFingerprint = observation.Fingerprint
	binding.TargetSchemaSHA256, err = businessModuleRecoveryCanonicalJSONSHA256(struct{ Shapes, Indexes, WriteSideEffects any }{observation.Shapes, observation.Indexes, observation.WriteSideEffects})
	return binding, err
}
