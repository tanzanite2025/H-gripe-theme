package localbackup

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

const independentServerLargeFixtureName = "whole-storefront-gibibyte-package-downloaded-to-computer-before-source-server-removal.age"
const independentServerLargeFixtureMediaBytes int64 = (1 << 30) + (16 << 20)

func TestPostgreSQLGibibyteComputerPackageSurvivesSourceRemovalAndRestoresToIndependentServer(t *testing.T) {
	phase := os.Getenv("LOCAL_BACKUP_FRESH_SERVER_DRILL_PHASE")
	directory := os.Getenv("LOCAL_BACKUP_COMPUTER_PACKAGE_DIRECTORY")
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" || directory == "" || (phase != "capture" && phase != "recover") {
		t.Skip("requires explicit disposable source/target phases and a computer transfer directory")
	}
	started := time.Now()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	cfg.DatabaseName = "local_backup_fresh_synthetic"
	cfg.MaxPackageBytes = 3 << 30
	cfg.MaxExpandedBytes = 4 << 30
	cfg.MaxWorkspaceBytes = 12 << 30
	if capacityDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_POSTGRESQL_CAPACITY_DIRECTORY"); capacityDirectory != "" {
		cfg.PostgreSQLCapacityDirectory = capacityDirectory
	}
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, syntheticExternalProjectionBoundaryForRecoveryReleaseTests()))
	ctx := context.Background()
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	packagePath := filepath.Join(directory, independentServerLargeFixtureName)
	if phase == "capture" {
		_, err := control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
		require.NoError(t, err)
		db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `CREATE TABLE schema_migrations(version bigint,dirty boolean);INSERT INTO schema_migrations VALUES(434,false);CREATE TABLE users(id bigint PRIMARY KEY);INSERT INTO users VALUES(1);CREATE TABLE orders(id bigserial PRIMARY KEY,user_id bigint REFERENCES users(id),total_minor bigint);INSERT INTO orders(user_id,total_minor) VALUES(1,12500),(1,12500);CREATE TABLE tickets(id bigint PRIMARY KEY);INSERT INTO tickets VALUES(1);CREATE TABLE ticket_messages(id bigint PRIMARY KEY,ticket_id bigint REFERENCES tickets(id),attachments text);INSERT INTO ticket_messages VALUES(1,1,'["/uploads-private/evidence/large-original.bin","/uploads-private/customer-receipt.txt"]');CREATE TABLE media_assets(id bigint PRIMARY KEY,storage_key text,visibility text,url text);INSERT INTO media_assets VALUES(1,'evidence/large-original.bin','private','/uploads-private/evidence/large-original.bin');CREATE TABLE outbox_events(id bigint PRIMARY KEY,status text);INSERT INTO outbox_events VALUES(1,'pending');`)
		require.NoError(t, err)
		db.Close()
		mediaPath := filepath.Join(cfg.StorageRoots["uploads_private"], "evidence", "large-original.bin")
		require.NoError(t, os.Mkdir(filepath.Dir(mediaPath), 0700))
		media, err := os.OpenFile(mediaPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		require.NoError(t, err)
		_, err = io.CopyN(media, rand.Reader, independentServerLargeFixtureMediaBytes)
		require.NoError(t, err)
		require.NoError(t, media.Close())
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "customer-receipt.txt"), []byte("synthetic-customer-original-receipt"), 0600))
		manager, err := NewManager(cfg)
		require.NoError(t, err)
		defer manager.Close()
		backup, err := manager.CreateBackup("synthetic-computer-drill", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
		require.NoError(t, err)
		backup = waitForGibibyteRecoveryTask(t, manager, backup.ID)
		require.Equal(t, "ready_to_download", backup.State, backup.Error)
		require.Greater(t, backup.Size, int64(1<<30), "exercise a real encrypted package larger than one GiB")
		server := httptest.NewServer(manager.HTTPHandler())
		defer server.Close()
		credential, err := readProtectedCredential(cfg.AuthenticationKeyFile)
		require.NoError(t, err)
		computerFile, err := os.OpenFile(packagePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		require.NoError(t, err)
		for _, interval := range []string{"bytes=0-16777215", "bytes=16777216-"} {
			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/downloads/"+backup.ID, nil)
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer "+credential)
			request.Header.Set("Range", interval)
			response, err := server.Client().Do(request)
			require.NoError(t, err)
			require.Equal(t, http.StatusPartialContent, response.StatusCode)
			_, err = io.Copy(computerFile, response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
		}
		require.NoError(t, computerFile.Sync())
		require.NoError(t, computerFile.Close())
		identity, err := fileIdentity(packagePath, "")
		require.NoError(t, err)
		require.Equal(t, backup.CiphertextSHA256, identity.SHA256)
		require.Equal(t, backup.Size, identity.Size)
		recordIndependentServerDrillMetrics(t, directory, phase, started, identity)
		return
	}
	identity, err := fileIdentity(packagePath, "")
	require.NoError(t, err)
	// A separate cluster/database stands for another project's authoritative
	// data. Both target database cases must leave it and gateway bytes intact.
	_, err = control.ExecContext(ctx, `CREATE DATABASE local_backup_erp_sentinel`)
	require.NoError(t, err)
	erp, err := cfg.connectDatabase(ctx, "local_backup_erp_sentinel")
	require.NoError(t, err)
	defer erp.Close()
	_, err = erp.ExecContext(ctx, `CREATE TABLE independent_order(id bigint PRIMARY KEY,reference text);INSERT INTO independent_order VALUES(999,'synthetic-erp-remains-independent')`)
	require.NoError(t, err)
	gatewayFile := filepath.Join(t.TempDir(), "independent-shared-gateway-configuration.txt")
	require.NoError(t, os.WriteFile(gatewayFile, []byte("synthetic-shared-gateway-unchanged"), 0600))
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer func() { manager.Close() }()
	for _, targetCase := range []string{"absent_database", "older_whole_machine_baseline"} {
		t.Logf("starting target case %s", targetCase)
		if targetCase == "older_whole_machine_baseline" {
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			baseline, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			_, err = baseline.ExecContext(ctx, `CREATE TABLE schema_migrations(version bigint,dirty boolean);INSERT INTO schema_migrations VALUES(434,false);CREATE TABLE users(id bigint PRIMARY KEY);INSERT INTO users VALUES(1);CREATE TABLE orders(id bigserial PRIMARY KEY,user_id bigint REFERENCES users(id),total_minor bigint);INSERT INTO orders(user_id,total_minor) VALUES(1,99),(1,88),(1,77)`)
			require.NoError(t, err)
			baseline.Close()
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "older-baseline-only.txt"), []byte("must-be-retained-with-old-database-only"), 0600))
		}
		var uploaded *Task
		manager, uploaded = uploadLargeComputerFileWithMaintenanceRestart(t, cfg, manager, packagePath)
		if targetCase == "absent_database" {
			// Model a real filesystem with less free space than the fixed safety
			// reserve. No isolated database may be created on this failed attempt.
			lowCapacityDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_LOW_CAPACITY_DIRECTORY")
			require.NotEmpty(t, lowCapacityDirectory)
			lowConfiguration := cfg
			lowConfiguration.PostgreSQLCapacityDirectory = lowCapacityDirectory
			require.NoError(t, manager.Close())
			manager, err = NewManager(lowConfiguration)
			require.NoError(t, err)
			_, err = manager.VerifyUploadedPackage(uploaded.ID)
			require.NoError(t, err)
			rejected := waitForGibibyteRecoveryTask(t, manager, uploaded.ID)
			require.Equal(t, "failed", rejected.State)
			require.Contains(t, rejected.Error, "insufficient PostgreSQL volume capacity")
			var stageExists bool
			require.NoError(t, control.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, rejected.StageDatabase).Scan(&stageExists))
			require.False(t, stageExists)
			require.NoError(t, manager.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
		}
		_, err = manager.VerifyUploadedPackage(uploaded.ID)
		require.NoError(t, err)
		verified := waitForGibibyteRecoveryTask(t, manager, uploaded.ID)
		require.Equal(t, "verified", verified.State, verified.Error)
		if targetCase == "absent_database" {
			require.Empty(t, verified.PreviewTargetTables)
		} else {
			require.EqualValues(t, 3, verified.PreviewTargetTables["orders"])
		}
		_, err = manager.ApplyVerifiedRecovery(verified.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+verified.Manifest.PackageID)
		require.NoError(t, err)
		switched := waitForGibibyteRecoveryTask(t, manager, verified.ID)
		require.Equal(t, "awaiting_reconciliation", switched.State, switched.Error)
		report, err := manager.InspectReconciliation(switched.ID)
		require.NoError(t, err)
		require.Empty(t, report.Gaps)
		require.EqualValues(t, 2, report.Tables["orders"])
		require.Equal(t, "25000", report.FinancialTotalsByCurrency["orders.total_minor"]["unspecified"])
		if targetCase == "absent_database" {
			_, err = manager.RollbackRecovery(switched.ID, "ROLLBACK "+cfg.SiteID+"/"+switched.ID)
			require.NoError(t, err)
			rolledBack := waitForGibibyteRecoveryTask(t, manager, switched.ID)
			require.Equal(t, "rolled_back", rolledBack.State, rolledBack.Error)
			var databaseExists bool
			require.NoError(t, control.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, cfg.DatabaseName).Scan(&databaseExists))
			require.False(t, databaseExists)
		} else {
			evidence := ReconciliationEvidence{ReconciledDatabaseFingerprint: report.DatabaseFingerprint, PaymentAndRefundReference: "synthetic-no-external-payments", ShippingReference: "synthetic-no-external-shipments", DurableEventsAndJobsReference: "synthetic-pending-events-checked", DedicatedRedisRebuildReference: "synthetic-no-redis-connection", AttachmentAccessAndIdentityReference: "synthetic-private-attachment-hash-checked", Operator: "synthetic-independent-server-drill"}
			evidence.ExternalProjections = syntheticExternalProjectionAttestationsForReconciliationReport(report)
			require.NoError(t, manager.CompleteReconciledRecovery(switched.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+switched.ID, evidence))
			restored, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			var newID int64
			require.NoError(t, restored.QueryRowContext(ctx, `INSERT INTO orders(user_id,total_minor) VALUES(1,500) RETURNING id`).Scan(&newID))
			require.EqualValues(t, 3, newID, "restored orders must accept the next unique identifier")
			var eventState string
			require.NoError(t, restored.QueryRowContext(ctx, `SELECT status FROM outbox_events WHERE id=1`).Scan(&eventState))
			require.Equal(t, "pending", eventState, "maintenance never starts a business event consumer")
			restored.Close()
			_, err = os.Stat(filepath.Join(cfg.StorageRoots["uploads_private"], "older-baseline-only.txt"))
			require.True(t, os.IsNotExist(err), "old baseline bytes must not leak into recovered media")
			oldBytes, err := os.ReadFile(filepath.Join(cfg.StorageRoots["uploads_private"], ".local-backup-old-"+switched.ID, "older-baseline-only.txt"))
			require.NoError(t, err)
			require.Equal(t, "must-be-retained-with-old-database-only", string(oldBytes))
		}
		var erpReference string
		require.NoError(t, erp.QueryRowContext(ctx, `SELECT reference FROM independent_order WHERE id=999`).Scan(&erpReference))
		require.Equal(t, "synthetic-erp-remains-independent", erpReference)
		gatewayBytes, err := os.ReadFile(gatewayFile)
		require.NoError(t, err)
		require.Equal(t, "synthetic-shared-gateway-unchanged", string(gatewayBytes))
	}
	recordIndependentServerDrillMetrics(t, directory, phase, started, identity)
}

func waitForGibibyteRecoveryTask(t *testing.T, manager *Manager, id string) *Task {
	t.Helper()
	deadline := time.Now().Add(15 * time.Minute)
	previousPhase := ""
	for time.Now().Before(deadline) {
		task, err := manager.GetTask(id)
		require.NoError(t, err)
		if task.Phase != previousPhase {
			t.Logf("large package task: %s", task.Phase)
			previousPhase = task.Phase
		}
		if task.State != "queued" && task.State != "running" && manager.operationMutex.TryLock() {
			manager.operationMutex.Unlock()
			return task
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal("large package recovery task exceeded fifteen-minute test deadline")
	return nil
}

func uploadLargeComputerFileWithMaintenanceRestart(t *testing.T, cfg Configuration, manager *Manager, packagePath string) (*Manager, *Task) {
	t.Helper()
	file, err := os.Open(packagePath)
	require.NoError(t, err)
	defer file.Close()
	info, err := file.Stat()
	require.NoError(t, err)
	credential, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	server := httptest.NewServer(manager.HTTPHandler())
	defer func() { server.Close() }()
	requestJSON := func(method, path string, body io.Reader, headers map[string]string) *Task {
		t.Helper()
		request, err := http.NewRequestWithContext(context.Background(), method, server.URL+path, body)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+credential)
		for name, value := range headers {
			request.Header.Set(name, value)
		}
		response, err := server.Client().Do(request)
		require.NoError(t, err)
		defer response.Body.Close()
		var result struct {
			Task  *Task  `json:"task"`
			Error string `json:"error"`
		}
		require.NoError(t, json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result))
		require.Equal(t, http.StatusOK, response.StatusCode, result.Error)
		return result.Task
	}
	task := requestJSON(http.MethodPost, "/api/uploads", strings.NewReader(fmt.Sprintf(`{"size":%d}`, info.Size())), map[string]string{"Content-Type": "application/json"})
	buffer := make([]byte, ChunkBytes)
	var offset int64
	for offset < info.Size() {
		size, err := file.ReadAt(buffer, offset)
		require.True(t, err == nil || err == io.EOF)
		digest := sha256.Sum256(buffer[:size])
		headers := map[string]string{"Upload-Offset": fmt.Sprint(offset), "X-Chunk-SHA256": hex.EncodeToString(digest[:]), "Content-Type": "application/octet-stream"}
		task = requestJSON(http.MethodPut, "/api/uploads/"+task.ID+"/chunks", bytes.NewReader(buffer[:size]), headers)
		offset += int64(size)
		if offset == 4*ChunkBytes {
			server.Close()
			require.NoError(t, manager.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			server = httptest.NewServer(manager.HTTPHandler())
			persisted := requestJSON(http.MethodGet, "/api/tasks/"+task.ID, nil, nil)
			require.Equal(t, offset, persisted.UploadedBytes)
			// Retry the already committed fourth chunk after executor restart.
			task = requestJSON(http.MethodPut, "/api/uploads/"+task.ID+"/chunks", bytes.NewReader(buffer[:size]), headers)
			require.Equal(t, offset, task.UploadedBytes)
			t.Logf("resumed encrypted upload at %d bytes after executor restart", offset)
		}
	}
	task = requestJSON(http.MethodPost, "/api/uploads/"+task.ID+"/commit", nil, nil)
	require.Equal(t, "uploaded", task.State)
	return manager, task
}

func recordIndependentServerDrillMetrics(t *testing.T, directory, phase string, started time.Time, packageIdentity FileIdentity) {
	t.Helper()
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	report := map[string]any{"phase": phase, "completed_at": time.Now().UTC(), "elapsed_seconds": time.Since(started).Seconds(), "media_bytes": independentServerLargeFixtureMediaBytes, "package_bytes": packageIdentity.Size, "package_sha256": packageIdentity.SHA256, "go_heap_system_bytes": memory.HeapSys, "go_heap_in_use_bytes": memory.HeapInuse, "source_and_target_are_disposable": true}
	require.NoError(t, writeDurableJSON(filepath.Join(directory, "gibibyte-independent-server-"+phase+"-validation-metrics.json"), report))
	t.Logf("completed %s: %d encrypted bytes in %s; Go heap system %d bytes", phase, packageIdentity.Size, time.Since(started).Round(time.Millisecond), memory.HeapSys)
}
