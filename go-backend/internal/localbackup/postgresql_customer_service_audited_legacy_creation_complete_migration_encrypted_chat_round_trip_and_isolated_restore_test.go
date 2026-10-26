package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func downloadEncryptedSelectedChatPackageUsingScopedHTTPAndUploadForAuditedCreationTest(t *testing.T, manager *Manager) *Task {
	t.Helper()
	cfg := manager.Configuration
	backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-module-exporter", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/customer_service", []string{"customer_service"})
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-module-exporter", []string{PermissionExport})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+backup.ID+"/download-ticket", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var ticket struct {
		DownloadPath string `json:"download_path"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &ticket))
	request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, ticket.DownloadPath, nil)
	for _, cookie := range response.Result().Cookies() {
		request.AddCookie(cookie)
	}
	response = httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	upload := uploadProductCatalogPackageForIntegrationTest(t, manager, response.Body.Bytes())
	require.Equal(t, "module_package_integrity_verified", upload.State, upload.Error)
	require.Equal(t, []string{"customer_service"}, upload.Manifest.SelectedModules)
	return upload
}

func TestPostgreSQLCustomerServiceAuditedLegacyCreationCompleteMigrationEncryptedChatRoundTripAndIsolatedRestore(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete application migrations")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_test_"+newTaskID()[:12]
	cfg.ApplicationDatabaseRole = "synthetic_audited_chat_app_" + newTaskID()[:12]
	cfg.MaxFiles, cfg.MaxExpandedBytes, cfg.MaxWorkspaceBytes = 10000, 256<<20, 1<<30
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
	require.NoError(t, err)
	defer func() { _, _ = control.ExecContext(ctx, `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)) }()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
	require.NoError(t, err)
	defer func() {
		_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	}()
	_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` SET timezone='Asia/Shanghai'`)
	require.NoError(t, err)
	database, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(database, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
	require.NoError(t, err)
	err = migrator.Up()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	_, err = migrator.Close()
	require.NoError(t, err)
	database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	defer database.Close()
	require.NoError(t, database.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	_, err = database.ExecContext(ctx, `INSERT INTO users(id,email,username,password,role) VALUES(900001,'audited-customer@example.invalid','audited-customer','synthetic-unused-password','user'),(900002,'audited-staff@example.invalid','audited-staff','synthetic-unused-password','admin');
INSERT INTO tickets(id,ticket_number,conversation_id,user_id,customer_user_id,assigned_to,subject,category) VALUES(900001,'synthetic-audited-chat','synthetic-audited-chat',900001,900001,900002,'Synthetic preserved conversation','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Synthetic private retained customer message','[]','media_asset_bindings');
INSERT INTO customer_service_inbox_states(ticket_id,recipient_user_id,last_read_message_id,assignment_version,unread_count) VALUES(900001,900002,0,1,1);
INSERT INTO media_assets(id,filename,url,storage_key,size,visibility,library_partition,customer_service_media_purpose,uploader_id,uploader_actor_type,classification_method) VALUES(900001,'audited-chat.txt','/api/customer-service/attachments/900001','customer-service/attachments/originals/audited-chat.txt',6,'private','customer_service','conversation_attachment',900001,'customer','recorded_upload_context');
INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,object_key,media_asset_id,status,expires_at) VALUES('synthetic-audited-chat-upload','customer_service',900001,'customer',900001,'customer-service/attachments/originals/audited-chat.txt',900001,'bound',now()+interval '1 day');
INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(900001,900001,0,'synthetic-audited-chat-upload');`)
	require.NoError(t, err)
	attachment := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service/attachments/originals/audited-chat.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(attachment), 0700))
	require.NoError(t, os.WriteFile(attachment, []byte("backup"), 0600))
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	repo := repository.NewOutboxRepository(orm)
	attemptedAt := time.Now().UTC().Add(-time.Hour)
	event := outbox.Event{EventKey: "customer_service.conversation.created:900001", EventType: outbox.EventTypeCustomerServiceRealtime, AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "900001", Payload: []byte(`{"status":"open","assigned_to":900002}`), Status: outbox.EventStatusDeadLetter, Attempts: 10, MaxAttempts: 10, LastAttemptAt: &attemptedAt, LastError: legacyCustomerServiceCreationIncompleteEnvelopeError}
	require.NoError(t, repo.CreateEvent(&event))
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	oldUpload := downloadEncryptedSelectedChatPackageUsingScopedHTTPAndUploadForAuditedCreationTest(t, manager)
	result, err := repo.ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "Synthetic private audit reason for preserved creation", time.Now().UTC(), &audit.AuditLog{UserID: 900002, Method: "POST", Path: "/api/admin/ops/outbox/failures/" + strconv.FormatUint(uint64(event.ID), 10) + "/ignore"})
	require.NoError(t, err)
	require.NotZero(t, result.AuditID)
	newUpload := downloadEncryptedSelectedChatPackageUsingScopedHTTPAndUploadForAuditedCreationTest(t, manager)
	verifyUploadWithoutChangingCurrentDatabase := func(upload *Task) *Task {
		t.Helper()
		_, before, _, err := captureDatabaseInventory(ctx, database)
		require.NoError(t, err)
		_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(upload.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+upload.ID)
		require.NoError(t, err)
		upload = waitForIndependentRecoveryTask(t, manager, upload.ID)
		require.True(t, upload.SelectedModuleRestorePreview.TargetUnchanged, upload.Error)
		require.True(t, upload.SelectedModuleRestorePreview.DisposableStageRemoved, upload.Error)
		require.False(t, upload.SelectedModuleRestorePreview.OnlineOverwriteAvailable)
		require.Nil(t, upload.SelectedModuleDisposableStage)
		require.Nil(t, upload.Journal)
		_, after, _, err := captureDatabaseInventory(ctx, database)
		require.NoError(t, err)
		require.Equal(t, before, after)
		return upload
	}
	oldUpload = verifyUploadWithoutChangingCurrentDatabase(oldUpload)
	require.Equal(t, "module_restore_conflicts_found", oldUpload.State, oldUpload.Error)
	newUpload = verifyUploadWithoutChangingCurrentDatabase(newUpload)
	require.Equal(t, "module_isolated_restore_verified", newUpload.State, newUpload.Error)
	coverage := newUpload.SelectedModuleRestorePreview.CustomerServiceOverwritePreview
	require.NotNil(t, coverage)
	for _, strategy := range []*SelectedCustomerServiceModuleOverwriteStrategyInspection{coverage.PackageRows, coverage.PreserveCurrentRows} {
		require.True(t, strategy.Complete)
		require.True(t, strategy.DomainChecksSatisfied, "%v", strategy.IssueCounts)
		require.Zero(t, strategy.BlockingIssues)
		require.Equal(t, 1, strategy.IssueCounts["audited_legacy_creation_notification_preserved_without_replay"])
	}
	var deliveries atomic.Int32
	dispatcher := service.NewOutboxService(repo)
	dispatcher.RegisterHandler(outbox.EventTypeCustomerServiceRealtime, func(context.Context, outbox.Event) error { deliveries.Add(1); return nil })
	dispatched, err := dispatcher.ProcessPending(ctx, time.Now().UTC(), 10)
	require.NoError(t, err)
	require.Zero(t, dispatched.Claimed)
	require.Zero(t, deliveries.Load())
	stored, err := repo.FindEventByID(event.ID)
	require.NoError(t, err)
	require.Equal(t, outbox.EventStatusIgnored, stored.Status)
	require.Equal(t, legacyCustomerServiceCreationIncompleteEnvelopeError, stored.LastError)
	require.Equal(t, 10, stored.Attempts)
	actualAttachment, err := os.ReadFile(attachment)
	require.NoError(t, err)
	require.Equal(t, "backup", string(actualAttachment))
	public, err := json.Marshal(coverage)
	require.NoError(t, err)
	for _, privateValue := range []string{event.EventKey, "Synthetic private audit reason", "Synthetic private retained customer message", "audited-customer@example.invalid"} {
		require.NotContains(t, string(public), privateValue)
	}
	// The module carries protected event rows, not authority to recreate audit
	// facts. Losing the target audit must block even with an intact new package.
	_, err = database.ExecContext(ctx, `DELETE FROM audit_logs WHERE id=$1`, result.AuditID)
	require.NoError(t, err)
	newUpload = verifyUploadWithoutChangingCurrentDatabase(newUpload)
	require.Equal(t, "module_restore_conflicts_found", newUpload.State, newUpload.Error)
	for _, strategy := range []*SelectedCustomerServiceModuleOverwriteStrategyInspection{newUpload.SelectedModuleRestorePreview.CustomerServiceOverwritePreview.PackageRows, newUpload.SelectedModuleRestorePreview.CustomerServiceOverwritePreview.PreserveCurrentRows} {
		require.Positive(t, strategy.IssueCounts["audited_legacy_creation_preservation_receipt_missing_or_ambiguous"])
		require.False(t, strategy.DomainChecksSatisfied)
	}
}
