package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func waitForIndependentRecoveryTask(t *testing.T, manager *Manager, id string) *Task {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		task, err := manager.GetTask(id)
		require.NoError(t, err)
		if task.State != "queued" && task.State != "running" {
			if manager.operationMutex.TryLock() {
				manager.operationMutex.Unlock()
				return task
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("independent recovery task timed out")
	return nil
}

func uploadEncryptedRecoveryFixture(t *testing.T, manager *Manager, data []byte) *Task {
	t.Helper()
	task, err := manager.BeginUpload("synthetic-operator", int64(len(data)))
	require.NoError(t, err)
	for offset := 0; offset < len(data); {
		end := offset + int(ChunkBytes)
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]
		digest := sha256.Sum256(chunk)
		_, err := manager.AppendUploadChunk(task.ID, int64(offset), chunk, hex.EncodeToString(digest[:]))
		require.NoError(t, err)
		offset = end
	}
	_, err = manager.CommitUpload(task.ID)
	require.NoError(t, err)
	_, err = manager.VerifyUploadedPackage(task.ID)
	require.NoError(t, err)
	verified := waitForIndependentRecoveryTask(t, manager, task.ID)
	require.Equal(t, "verified", verified.State, verified.Error)
	return verified
}

func TestPostgreSQLWholeStorefrontRoundTripPairedRollbackAndRecoveryWithoutOriginalDatabase(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	if connectionURL == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL 15 test server and local pg_dump/pg_restore")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	t.Setenv("SERVER_BASE_URL", "https://synthetic-backup-media.invalid")
	t.Setenv("STOREFRONT_BASE_URL", "")
	require.NoError(t, cfg.inheritMediaOriginsFromExistingServerAndStorefrontDeploymentURLs())
	cfg.DatabaseURL = connectionURL
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	companyPhotoFile := prepareCompanyStaticReleaseFilesForWholeStorefrontRecoveryTest(t, &cfg)
	requireCurrentCompanyPhotoBytes := func(expected string) {
		t.Helper()
		contents, err := os.ReadFile(companyPhotoFile)
		require.NoError(t, err)
		require.Equal(t, expected, string(contents))
	}
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, syntheticExternalProjectionBoundaryForRecoveryReleaseTests()))
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
		names := []string{cfg.DatabaseName}
		for _, task := range tasks {
			if task.StageDatabase != "" {
				names = append(names, task.StageDatabase)
			}
			if task.Journal != nil {
				names = append(names, task.Journal.RollbackDatabase)
			}
		}
		for _, name := range names {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
		}
	}()
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
CREATE TABLE schema_migrations(version bigint NOT NULL,dirty boolean NOT NULL);INSERT INTO schema_migrations VALUES(434,false);
CREATE TABLE users(id bigint PRIMARY KEY);INSERT INTO users VALUES(7);
CREATE TABLE orders(id bigint PRIMARY KEY,user_id bigint REFERENCES users(id),total_minor bigint NOT NULL);INSERT INTO orders VALUES(11,7,12500);
CREATE TABLE order_items(id bigint PRIMARY KEY,order_id bigint REFERENCES orders(id),total_minor bigint NOT NULL);INSERT INTO order_items VALUES(12,11,12500);
CREATE TABLE tickets(id bigint PRIMARY KEY);INSERT INTO tickets VALUES(21);
CREATE TABLE ticket_messages(id bigint PRIMARY KEY,ticket_id bigint REFERENCES tickets(id),attachments text);INSERT INTO ticket_messages VALUES(22,21,'["/uploads-private/customer-receipt.txt"]');
CREATE TABLE media_assets(id bigint PRIMARY KEY,storage_key text,visibility text,url text);INSERT INTO media_assets VALUES(31,'customer-receipt.txt','private','/uploads-private/customer-receipt.txt');
CREATE TABLE media_asset_derivatives(id bigint PRIMARY KEY,media_asset_id bigint REFERENCES media_assets(id),storage_key text);
CREATE TABLE ticket_message_media_attachments(id bigint PRIMARY KEY,ticket_message_id bigint REFERENCES ticket_messages(id),media_asset_id bigint REFERENCES media_assets(id));INSERT INTO ticket_message_media_attachments VALUES(32,22,31);
CREATE TABLE outbox_events(id bigint PRIMARY KEY,status text);INSERT INTO outbox_events VALUES(41,'pending');`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE product_media(id bigint PRIMARY KEY,url text,thumbnail_url text);INSERT INTO product_media VALUES(61,'/company/products/backup-required-photo.txt','')`)
	require.NoError(t, err)
	db.Close()
	for name, content := range map[string]string{"uploads/order-image.txt": "order-evidence", "uploads_private/customer-receipt.txt": "original-private-conversation", "site_logo_uploads/site-logo.txt": "site-branding"} {
		parts := strings.SplitN(name, "/", 2)
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots[parts[0]], parts[1]), []byte(content), 0600))
	}
	faviconFile := prepareWholeBackupSiteIdentityFilesAndVerifyWrongRootDecoysCannotResolveMissingObjects(t, cfg)
	backup, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	require.Empty(t, backup.Manifest.Gaps)
	require.Equal(t, cfg.ApplicationVersion, backup.Manifest.StorefrontCompanyStaticReleaseVersion)
	packagePath, _ := manager.taskPackagePath(backup.ID)
	ciphertext, err := os.ReadFile(packagePath)
	require.NoError(t, err)
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO orders VALUES(13,7,300)`)
	require.NoError(t, err)
	db.Close()
	privateFile := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-receipt.txt")
	require.NoError(t, os.WriteFile(privateFile, []byte("newer-online-private-attachment"), 0600))
	require.NoError(t, os.WriteFile(faviconFile, []byte("newer-online-site-favicon"), 0600))
	require.NoError(t, os.WriteFile(companyPhotoFile, []byte("newer-company-photo-bytes"), 0600))
	verified := uploadEncryptedRecoveryFixture(t, manager, ciphertext)
	require.EqualValues(t, 2, verified.PreviewTargetTables["orders"])
	// Preflight is isolated: it has not changed current order data or media.
	contents, err := os.ReadFile(privateFile)
	require.NoError(t, err)
	require.Equal(t, "newer-online-private-attachment", string(contents))
	faviconBytes, err := os.ReadFile(faviconFile)
	require.NoError(t, err)
	require.Equal(t, "newer-online-site-favicon", string(faviconBytes), "upload preview must preserve live site identity files")
	requireCurrentCompanyPhotoBytes("newer-company-photo-bytes")
	_, err = manager.ApplyVerifiedRecovery(verified.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+verified.Manifest.PackageID)
	require.NoError(t, err)
	switched := waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Equal(t, "awaiting_reconciliation", switched.State, switched.Error)
	var connectionsAllowed bool
	require.NoError(t, control.QueryRowContext(ctx, `SELECT datallowconn FROM pg_database WHERE datname=$1`, cfg.DatabaseName).Scan(&connectionsAllowed))
	require.False(t, connectionsAllowed)
	contents, err = os.ReadFile(privateFile)
	require.NoError(t, err)
	require.Equal(t, "original-private-conversation", string(contents))
	faviconBytes, err = os.ReadFile(faviconFile)
	require.NoError(t, err)
	require.Equal(t, "original-site-favicon", string(faviconBytes))
	requireCurrentCompanyPhotoBytes("original-company-photo-bytes")
	require.NoError(t, manager.Close())
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
	_, err = manager.RollbackRecovery(switched.ID, "ROLLBACK "+cfg.SiteID+"/"+switched.ID)
	require.NoError(t, err)
	rolledBack := waitForIndependentRecoveryTask(t, manager, switched.ID)
	require.Equal(t, "rolled_back", rolledBack.State, rolledBack.Error)
	contents, err = os.ReadFile(privateFile)
	require.NoError(t, err)
	require.Equal(t, "newer-online-private-attachment", string(contents))
	faviconBytes, err = os.ReadFile(faviconFile)
	require.NoError(t, err)
	require.Equal(t, "newer-online-site-favicon", string(faviconBytes), "paired rollback must restore the newer favicon together with the database")
	requireCurrentCompanyPhotoBytes("newer-company-photo-bytes")
	// A fresh upload can now replace the same target after a complete rollback.
	verified = uploadEncryptedRecoveryFixture(t, manager, ciphertext)
	_, err = manager.ApplyVerifiedRecovery(verified.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+verified.Manifest.PackageID)
	require.NoError(t, err)
	switched = waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Equal(t, "awaiting_reconciliation", switched.State, switched.Error)
	evidence := ReconciliationEvidence{PaymentAndRefundReference: "synthetic-payment-report", ShippingReference: "synthetic-shipping-report", DurableEventsAndJobsReference: "synthetic-jobs-report-no-replay", DedicatedRedisRebuildReference: "synthetic-dedicated-cache-report", AttachmentAccessAndIdentityReference: "synthetic-access-report", Operator: "synthetic-operator"}
	report, err := manager.InspectReconciliation(switched.ID)
	require.NoError(t, err)
	require.Empty(t, report.Gaps)
	evidence.ReconciledDatabaseFingerprint = report.DatabaseFingerprint
	evidence.ExternalProjections = syntheticExternalProjectionAttestationsForReconciliationReport(report)
	require.NoError(t, manager.OpenReconciliationConnections(switched.ID, "RECONCILE "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID))
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE outbox_events SET status='reviewed_without_dispatch' WHERE id=41`)
	require.NoError(t, err)
	db.Close()
	require.Error(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence), "stale reconciliation fingerprints cannot release the service")
	report, err = manager.InspectReconciliation(switched.ID)
	require.NoError(t, err)
	evidence.ReconciledDatabaseFingerprint = report.DatabaseFingerprint
	evidence.ExternalProjections = syntheticExternalProjectionAttestationsForReconciliationReport(report)
	validExternalEvidence := evidence.ExternalProjections
	evidence.ExternalProjections = nil
	require.ErrorContains(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence), "external projection evidence")
	require.NoError(t, control.QueryRowContext(ctx, `SELECT datallowconn FROM pg_database WHERE datname=$1`, cfg.DatabaseName).Scan(&connectionsAllowed))
	require.False(t, connectionsAllowed, "a rejected final attestation must not reopen administrative SQL logins")
	evidence.ExternalProjections = validExternalEvidence
	for _, scenario := range []string{"duplicate_projection", "missing_dedupe_verification", "unsupported_cache_flush"} {
		t.Run(scenario, func(t *testing.T) {
			invalid := syntheticExternalProjectionAttestationsForReconciliationReport(report)
			switch scenario {
			case "duplicate_projection":
				invalid.Attestations[1] = invalid.Attestations[0]
			case "missing_dedupe_verification":
				invalid.Attestations[2].Action = "stream_only"
			case "unsupported_cache_flush":
				invalid.Attestations[1].Action = "flushdb"
			}
			evidence.ExternalProjections = invalid
			require.Error(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence))
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
		})
	}
	evidence.ExternalProjections = validExternalEvidence
	// Database reconciliation does not excuse a subsequently missing or changed
	// private attachment. Service release must recheck the actual media trees.
	require.NoError(t, os.WriteFile(companyPhotoFile, []byte("corrupted-after-static-reconciliation"), 0600))
	require.Error(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence), "changed company static bytes must block final service release")
	require.NoError(t, os.WriteFile(companyPhotoFile, []byte("original-company-photo-bytes"), 0600))
	require.NoError(t, os.Remove(privateFile))
	require.Error(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence))
	fenced, err = maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
	require.NoError(t, os.WriteFile(privateFile, []byte("original-private-conversation"), 0600))
	require.NoError(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence))
	completed, err := manager.GetTask(switched.ID)
	require.NoError(t, err)
	require.Equal(t, evidence.ExternalProjections, completed.Reconciliation.ExternalProjections)
	require.Equal(t, evidence.ExternalProjections.ContractSHA256, completed.ExternalProjectionReconciliationContract.ContractSHA256)
	if evidenceDirectory := os.Getenv("LOCAL_BACKUP_EXTERNAL_PROJECTION_VALIDATION_EVIDENCE_DIRECTORY"); evidenceDirectory != "" {
		require.NoError(t, os.MkdirAll(evidenceDirectory, 0700))
		require.NoError(t, writeDurableJSON(filepath.Join(evidenceDirectory, "synthetic-whole-recovery-completed-task-and-operator-attestations.json"), completed))
		require.NoError(t, writeDurableJSON(filepath.Join(evidenceDirectory, "synthetic-whole-recovery-frozen-reconciliation-report.json"), report))
	}
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	var orderCount, totalMinor int64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*),sum(total_minor) FROM orders`).Scan(&orderCount, &totalMinor))
	require.EqualValues(t, 1, orderCount)
	require.EqualValues(t, 12500, totalMinor)
	var eventState string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM outbox_events WHERE id=41`).Scan(&eventState))
	require.Equal(t, "reviewed_without_dispatch", eventState)
	db.Close()
	// Simulate loss of the original BUSINESS database on this disposable server.
	// The recovery configuration, authentication and stage journal stay outside it.
	_, err = control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	require.NoError(t, err)
	require.NoError(t, os.Remove(companyPhotoFile))
	require.NoError(t, os.Remove(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename)))
	verified = uploadEncryptedRecoveryFixture(t, manager, ciphertext)
	require.Empty(t, verified.PreviewTargetTables)
	_, err = manager.ApplyVerifiedRecovery(verified.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+verified.Manifest.PackageID)
	require.NoError(t, err)
	switched = waitForIndependentRecoveryTask(t, manager, verified.ID)
	require.Equal(t, "awaiting_reconciliation", switched.State, switched.Error)
	require.Zero(t, switched.Journal.OldDatabaseOID)
	report, err = manager.InspectReconciliation(switched.ID)
	require.NoError(t, err)
	require.Empty(t, report.Gaps)
	evidence.ReconciledDatabaseFingerprint = report.DatabaseFingerprint
	// Identical package contents on another recovery task cannot reuse an old
	// projection attestation: the task and switched database OID are different.
	require.ErrorContains(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence), "external projection evidence")
	evidence.ExternalProjections = syntheticExternalProjectionAttestationsForReconciliationReport(report)
	require.NoError(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence))
	requireCurrentCompanyPhotoBytes("original-company-photo-bytes")
	// Legacy malformed attachment JSON and a missing private object are captured
	// as explicit gaps. They can be downloaded for investigation, never applied.
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE ticket_messages SET attachments='{malformed-history' WHERE id=22`)
	require.NoError(t, err)
	db.Close()
	require.NoError(t, os.Remove(privateFile))
	gapped, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	gapped = waitForIndependentRecoveryTask(t, manager, gapped.ID)
	require.Equal(t, "generated_with_gaps", gapped.State, gapped.Error)
	require.GreaterOrEqual(t, len(gapped.Manifest.Gaps), 2)
	gappedFile, _ := manager.taskPackagePath(gapped.ID)
	gappedData, err := os.ReadFile(gappedFile)
	require.NoError(t, err)
	rejected, err := manager.BeginUpload("synthetic-operator", int64(len(gappedData)))
	require.NoError(t, err)
	digest := sha256.Sum256(gappedData)
	_, err = manager.AppendUploadChunk(rejected.ID, 0, gappedData, hex.EncodeToString(digest[:]))
	require.NoError(t, err)
	_, err = manager.CommitUpload(rejected.ID)
	require.NoError(t, err)
	_, err = manager.VerifyUploadedPackage(rejected.ID)
	require.NoError(t, err)
	rejected = waitForIndependentRecoveryTask(t, manager, rejected.ID)
	require.Equal(t, "failed", rejected.State)
	require.Contains(t, rejected.Error, "unresolved")
	require.Empty(t, rejected.StageDatabase)
}

func TestPostgreSQLInterruptedRenameAndMediaMoveJournalsAlwaysRollbackAsAPair(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	if connectionURL == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server")
	}
	for _, phase := range []string{"before_database_rename", "after_old_database_rename", "after_old_media_rename_before_commit", "after_new_media_rename_before_commit", "after_old_company_static_rename_before_commit", "after_new_company_static_rename_before_commit", "after_new_database_rename_before_commit"} {
		t.Run(phase, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL = connectionURL
			cfg.DatabaseName = "local_backup_crash_" + newTaskID()[:12]
			prepareCompanyStaticReleaseFilesForWholeStorefrontRecoveryTest(t, &cfg)
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			defer control.Close()
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, `CREATE TABLE schema_migrations(version bigint,dirty boolean);INSERT INTO schema_migrations VALUES(434,false);CREATE TABLE retained_business_fact(id bigint PRIMARY KEY,value text);INSERT INTO retained_business_fact VALUES(1,'current-before-recovery')`)
			require.NoError(t, err)
			db.Close()
			manager, err := NewManager(cfg)
			require.NoError(t, err)
			task, err := manager.createTask("upload", "synthetic-operator")
			require.NoError(t, err)
			task.StageDatabase = "local_backup_stage_" + task.ID[:16]
			oldName := "local_backup_old_" + task.ID[:16]
			defer func() {
				manager.Close()
				for _, name := range []string{cfg.DatabaseName, task.StageDatabase, oldName} {
					_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH(FORCE)`)
				}
			}()
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` WITH TEMPLATE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			for _, storageRoot := range cfg.wholeApplicationRecoveryStorageRoots(true) {
				name, root := storageRoot.Name, storageRoot.Directory
				require.NoError(t, os.WriteFile(filepath.Join(root, "fact.txt"), []byte("current-original-"+name), 0600))
				oldRoot := filepath.Join(root, ".local-backup-old-"+task.ID)
				stageRoot := filepath.Join(root, ".local-backup-stage-"+task.ID)
				require.NoError(t, os.Mkdir(oldRoot, 0700))
				require.NoError(t, os.Mkdir(stageRoot, 0700))
				require.NoError(t, os.WriteFile(filepath.Join(stageRoot, "fact.txt"), []byte("restored-older-"+name), 0600))
			}
			task.PreviewTargetFingerprint, task.PreviewTargetTables, err = cfg.targetFingerprint(ctx)
			require.NoError(t, err)
			journal := &RecoveryJournal{RollbackDatabase: oldName, Moves: []RecoveryFileMove{}}
			require.NoError(t, readDatabaseOID(ctx, control, cfg.DatabaseName, &journal.OldDatabaseOID))
			require.NoError(t, readDatabaseOID(ctx, control, task.StageDatabase, &journal.NewDatabaseOID))
			for _, storageRoot := range cfg.wholeApplicationRecoveryStorageRoots(true) {
				root := storageRoot.Directory
				journal.Moves = append(journal.Moves, RecoveryFileMove{Source: filepath.Join(root, "fact.txt"), Destination: filepath.Join(root, ".local-backup-old-"+task.ID, "fact.txt")}, RecoveryFileMove{Source: filepath.Join(root, ".local-backup-stage-"+task.ID, "fact.txt"), Destination: filepath.Join(root, "fact.txt")})
			}
			task.Journal = journal
			task.State = "running"
			task.Phase = phase
			require.NoError(t, manager.saveTask(task))
			require.NoError(t, writeDurableJSON(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename), map[string]string{"task_id": task.ID}))
			if phase != "before_database_rename" {
				require.NoError(t, sealDatabase(ctx, control, cfg.DatabaseName))
				_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` RENAME TO `+pq.QuoteIdentifier(oldName))
				require.NoError(t, err)
			}
			moveCount := 0
			switch phase {
			case "after_old_media_rename_before_commit":
				moveCount = 1
			case "after_new_media_rename_before_commit":
				moveCount = 2
			case "after_old_company_static_rename_before_commit":
				moveCount = len(mediaRoots)*2 + 1
			case "after_new_company_static_rename_before_commit":
				moveCount = len(mediaRoots)*2 + 2
			case "after_new_database_rename_before_commit":
				moveCount = len(journal.Moves)
			}
			for index := 0; index < moveCount; index++ {
				journal.Moves[index].Started = true
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, performJournaledFileMove(&journal.Moves[index]))
				if index < moveCount-1 {
					require.NoError(t, manager.saveTask(task))
				}
			}
			if phase == "after_new_database_rename_before_commit" {
				_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName))
				require.NoError(t, err)
			}
			require.NoError(t, manager.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			interrupted, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "interrupted", interrupted.State)
			_, err = manager.RollbackRecovery(task.ID, "ROLLBACK "+cfg.SiteID+"/"+task.ID)
			require.NoError(t, err)
			restored := waitForIndependentRecoveryTask(t, manager, task.ID)
			require.Equal(t, "rolled_back", restored.State, restored.Error)
			fingerprint, _, err := cfg.targetFingerprint(ctx)
			require.NoError(t, err)
			require.Equal(t, task.PreviewTargetFingerprint, fingerprint)
		})
	}
}
