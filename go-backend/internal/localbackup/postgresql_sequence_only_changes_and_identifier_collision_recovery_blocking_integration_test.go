package localbackup

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLSequenceOnlyChangesBlockStalePreviewAndBrokenSequencesBlockRecovery(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires an explicitly disposable PostgreSQL server")
	}
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	ctx := context.Background()
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
	require.NoError(t, err)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer func() {
		manager.Close()
		tasks, _ := manager.ListTasks()
		for _, task := range tasks {
			if task.StageDatabase != "" {
				_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(task.StageDatabase)+` WITH (FORCE)`)
			}
		}
		_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	}()
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE schema_migrations(version bigint,dirty boolean);INSERT INTO schema_migrations VALUES(434,false);CREATE TABLE users(id bigint PRIMARY KEY);INSERT INTO users VALUES(1);CREATE TABLE orders(id bigserial PRIMARY KEY,user_id bigint REFERENCES users(id),total_minor bigint);INSERT INTO orders(user_id,total_minor) VALUES(1,12500);CREATE TABLE descending_receipts(id bigint PRIMARY KEY);CREATE SEQUENCE descending_receipts_id_seq INCREMENT BY -1 MINVALUE -10 MAXVALUE -1 START -1 OWNED BY descending_receipts.id;ALTER TABLE descending_receipts ALTER COLUMN id SET DEFAULT nextval('descending_receipts_id_seq');INSERT INTO descending_receipts DEFAULT VALUES;`)
	require.NoError(t, err)
	db.Close()
	backup, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	packagePath, err := manager.taskPackagePath(backup.ID)
	require.NoError(t, err)
	data, err := os.ReadFile(packagePath)
	require.NoError(t, err)
	verified := uploadEncryptedRecoveryFixture(t, manager, data)
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `SELECT nextval('orders_id_seq')`)
	require.NoError(t, err)
	db.Close()
	confirmation := "RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + verified.Manifest.PackageID
	_, err = manager.ApplyVerifiedRecovery(verified.ID, confirmation)
	require.NoError(t, err)
	failed := waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "target changed after preview")
	require.Nil(t, failed.Journal)
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(verified.ID))
	_, err = manager.VerifyUploadedPackage(verified.ID)
	require.NoError(t, err)
	verified = waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Equal(t, "verified", verified.State, verified.Error)
	stage, err := cfg.connectDatabase(ctx, verified.StageDatabase)
	require.NoError(t, err)
	_, err = stage.ExecContext(ctx, `SELECT nextval('orders_id_seq')`)
	require.NoError(t, err)
	stage.Close()
	_, err = manager.ApplyVerifiedRecovery(verified.ID, confirmation)
	require.NoError(t, err)
	failed = waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Contains(t, failed.Error, "isolated database changed")
	require.Nil(t, failed.Journal)
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(verified.ID))
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `SELECT setval('orders_id_seq',1,false);SELECT setval('descending_receipts_id_seq',-1,false)`)
	require.NoError(t, err)
	db.Close()
	gapped, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	gapped = waitForIndependentRecoveryTask(t, manager, gapped.ID)
	require.Equal(t, "generated_with_gaps", gapped.State, gapped.Error)
	require.Contains(t, strings.Join(gapped.Manifest.Gaps, " "), "unsafe_owned_unique_sequence:orders_id_seq")
	require.Contains(t, strings.Join(gapped.Manifest.Gaps, " "), "unsafe_owned_unique_sequence:descending_receipts_id_seq")
}
