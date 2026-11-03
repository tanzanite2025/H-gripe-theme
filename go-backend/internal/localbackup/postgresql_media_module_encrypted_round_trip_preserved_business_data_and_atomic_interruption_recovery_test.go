package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func downloadEncryptedMediaModulePackageThroughScopedHTTPAndUploadForFormalRestoreTest(t *testing.T, manager *Manager) *Task {
	t.Helper()
	cfg := manager.Configuration
	backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-media-exporter", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/media", []string{"media"})
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-media-exporter", []string{PermissionExport})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+backup.ID+"/download-ticket", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
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
	require.Equal(t, []string{"media"}, upload.Manifest.SelectedModules)
	return upload
}

func TestPostgreSQLMediaModuleEncryptedDownloadUploadFormalRestoreAndAtomicInterruptionRecovery(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL and complete application migrations")
	}
	for _, scenario := range []string{"authenticated_restore", "restore_intent_persisted", "attachment_linked_before_acknowledgement", "rows_and_sequences_before_commit", "rows_committed_before_acknowledgement", "verified_before_fence_removal", "fence_removed_before_acknowledgement", "stale_database_after_preview", "different_file_bytes", "wrong_private_actor", "unfinished_upload", "live_cleanup_hazard", "private_retention_cleanup_hazard", "wrong_product_variant_owner", "different_registration_contents", "company_release_mismatch", "retained_legacy_private_owner_conflict", "all_eleven_media_tables", "repair_public_and_private_files", "repair_media_original_backup_linked_before_acknowledgement", "repair_media_file_replaced_before_acknowledgement", "repair_rows_and_sequences_before_commit", "repair_rows_committed_before_acknowledgement", "repair_after_preview_changed_bytes", "repair_missing_registration_conflict", "repair_changed_registration_conflict", "repair_unregistered_file_conflict"} {
		t.Run(scenario, func(t *testing.T) {
			repairScenario := strings.HasPrefix(scenario, "repair_")
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_media_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_media_" + newTaskID()[:12]
			cfg.MaxFiles, cfg.MaxExpandedBytes, cfg.MaxWorkspaceBytes = 10000, 256<<20, 1<<30
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			var manager *Manager
			t.Cleanup(func() {
				if manager != nil {
					require.NoError(t, manager.Close())
				}
				_, err := control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
				require.NoError(t, err)
				_, err = control.ExecContext(ctx, `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
				require.NoError(t, err)
				require.NoError(t, control.Close())
			})
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
			require.NoError(t, database.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))

			cfg.ProductStaticAssetsDirectory = t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(cfg.ProductStaticAssetsDirectory, "company"), 0700))
			require.NoError(t, os.WriteFile(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", "fixture.png"), []byte("same release file"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename), []byte(cfg.ApplicationVersion), 0600))
			cfg.PublicOrigins = []string{"https://shop.example.test"}
			_, err = database.ExecContext(ctx, `
INSERT INTO users(id,email,username,password,role) VALUES(900001,'media-customer@example.invalid','media-customer','synthetic','user'),(900002,'media-admin@example.invalid','media-admin','synthetic','admin');
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900001,'Preserved media product','preserved-media-product',123,'USD'),(900002,'Other product','other-media-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'preserved-media-sku',123,'USD',19),(900002,900002,'other-media-sku',123,'USD',7);
INSERT INTO orders(id,order_number,user_id,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'preserved-media-order',900001,'pending','USD','USD',123,123,123,0,0,0,0);
INSERT INTO tickets(id,ticket_number,conversation_id,user_id,customer_user_id,assigned_to,subject,category) VALUES(900001,'preserved-media-chat','preserved-media-chat',900001,900001,900002,'Preserved media conversation','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Preserved conversation and original ownership','[]','media_asset_bindings');
INSERT INTO media_assets(id,filename,url,storage_key,size,mime_type,media_type,visibility,status,uploader_id,uploader_actor_type,library_partition,upload_origin,classification_method,customer_service_media_purpose) VALUES
(900001,'public.png','/uploads/products/media-original.png','products/media-original.png',21,'image/png','image','public','active',900002,'administrator','catalog','administrator_catalog_upload','recorded_upload_context',''),
(900002,'private.png','/uploads/customer-service/attachments/originals/media-private.png','customer-service/attachments/originals/media-private.png',21,'image/png','image','private','active',900001,'customer','customer_service','customer_service_conversation_upload','recorded_upload_context','conversation_attachment');
INSERT INTO media_asset_derivatives(id,media_asset_id,preset,preset_version,url,storage_key,mime_type,size) VALUES
(900001,900001,'thumbnail',1,'/uploads/media-derivatives/900001/thumbnail/v1/public.png','media-derivatives/900001/thumbnail/v1/public.png','image/png',21),
(900002,900002,'thumbnail',1,'/uploads/customer-service/attachments/derivatives/900002/thumbnail/v1/private.png','customer-service/attachments/derivatives/900002/thumbnail/v1/private.png','image/png',21);
INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,object_key,media_asset_id,status,expires_at) VALUES('11111111-1111-4111-8111-111111111111','customer_service',900001,'customer',900001,'customer-service/attachments/originals/media-private.png',900002,'bound',now()-interval '1 day');
INSERT INTO ticket_message_media_attachments(id,ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(900001,900001,900002,0,'11111111-1111-4111-8111-111111111111');
INSERT INTO product_media(id,product_id,variant_id,media_asset_id,url,media_type) VALUES(900001,900001,900001,900001,'/uploads/products/media-original.png','image');
`)
			require.NoError(t, err)
			if scenario == "all_eleven_media_tables" {
				digest := transactionalBusinessModuleRestoreBytesSHA256([]byte("synthetic-media-file!"))
				_, err = database.ExecContext(ctx, `
 INSERT INTO media(id,filename,url,mime_type,size,uploader_id) VALUES(900001,'legacy.png','/uploads/legacy-media.png','image/png',21,900002);
 INSERT INTO media_derivative_presets(id,code,label,max_width) VALUES(900010,'restored_custom','Custom restored preset',777);
 INSERT INTO media_derivative_rebuild_jobs(id,status,finished_at) VALUES(900001,'succeeded',now());
 INSERT INTO media_assets(id,filename,url,storage_key,size,mime_type,media_type,visibility,status,uploader_id,uploader_actor_type,library_partition,upload_origin,classification_method) VALUES(900003,'draft.png','/uploads/business-draft-media/22222222-2222-4222-8222-222222222222/original.png','business-draft-media/22222222-2222-4222-8222-222222222222/original.png',21,'image/png','image','public','active',900002,'administrator','catalog','administrator_catalog_upload','recorded_upload_context');
 INSERT INTO product_media(id,product_id,media_asset_id,url) VALUES(900003,900001,900003,'/uploads/business-draft-media/22222222-2222-4222-8222-222222222222/original.png');
 `)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `INSERT INTO business_draft_media_upload_receipts(id,asset_id,uploader_id,upload_origin,original_object_key,object_keys_json,content_sha256,size,mime_type,status,expires_at) VALUES('22222222-2222-4222-8222-222222222222',900003,900002,'administrator_catalog_upload','business-draft-media/22222222-2222-4222-8222-222222222222/original.png','["business-draft-media/22222222-2222-4222-8222-222222222222/original.png"]',$1,21,'image/png','bound',now()-interval '1 day');`, digest)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `INSERT INTO site_logo_assets(id,filename,url,storage_key,size,content_sha256,uploader_id) VALUES(1,'restored-logo.svg','/uploads/site-logo/restored.svg','site-logo/restored.svg',21,$1,900002);`, digest)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `INSERT INTO site_favicon_assets(id,filename,url,storage_key,size,content_sha256,uploader_id) VALUES(1,'restored-favicon.webp','/uploads/site-favicon/restored.webp','site-favicon/restored.webp',21,$1,900002);`, digest)
				require.NoError(t, err)
			}
			if scenario == "wrong_private_actor" {
				_, err = database.ExecContext(ctx, `UPDATE media_upload_sessions SET upload_actor_account_id=900002 WHERE id='11111111-1111-4111-8111-111111111111'`)
				require.NoError(t, err)
			}
			if scenario == "unfinished_upload" {
				_, err = database.ExecContext(ctx, `UPDATE media_upload_sessions SET status='ready' WHERE id='11111111-1111-4111-8111-111111111111'`)
				require.NoError(t, err)
			}
			if scenario == "wrong_product_variant_owner" {
				_, err = database.ExecContext(ctx, `UPDATE product_media SET variant_id=900002 WHERE id=900001`)
				require.NoError(t, err)
			}
			tx, err := database.BeginTx(ctx, nil)
			require.NoError(t, err)
			snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, []string{"media"})
			require.NoError(t, err)
			require.NoError(t, tx.Rollback())
			for _, name := range snapshot.RequiredFilePaths {
				var filename string
				for _, root := range mediaRoots {
					prefix := "media/" + root + "/"
					if strings.HasPrefix(name, prefix) {
						filename = filepath.Join(cfg.StorageRoots[root], filepath.FromSlash(strings.TrimPrefix(name, prefix)))
					}
				}
				if strings.HasPrefix(name, "product-static/") {
					filename = filepath.Join(cfg.ProductStaticAssetsDirectory, filepath.FromSlash(strings.TrimPrefix(name, "product-static/")))
				}
				require.NotEmpty(t, filename)
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
				require.NoError(t, os.WriteFile(filename, []byte("synthetic-media-file"+"!"), 0600))
			}
			orphan := filepath.Join(cfg.StorageRoots["uploads_private"], "unregistered-backup-file.bin")
			require.NoError(t, os.WriteFile(orphan, []byte("unregistered original bytes"), 0600))
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			task := downloadEncryptedMediaModulePackageThroughScopedHTTPAndUploadForFormalRestoreTest(t, manager)
			if !repairScenario || scenario == "repair_missing_registration_conflict" {
				_, err = database.ExecContext(ctx, `DELETE FROM ticket_message_media_attachments WHERE id=900001;DELETE FROM media_upload_sessions WHERE id='11111111-1111-4111-8111-111111111111';DELETE FROM product_media WHERE id=900001;DELETE FROM media_asset_derivatives WHERE id IN(900001,900002);DELETE FROM media_assets WHERE id IN(900001,900002);`)
				require.NoError(t, err)
			}
			_, err = database.ExecContext(ctx, `UPDATE orders SET admin_note='New order note must remain' WHERE id=900001;`)
			require.NoError(t, err)
			if scenario == "all_eleven_media_tables" {
				_, err = database.ExecContext(ctx, `DELETE FROM media WHERE id=900001;DELETE FROM media_derivative_presets WHERE id=900010;DELETE FROM media_derivative_rebuild_jobs WHERE id=900001;DELETE FROM business_draft_media_upload_receipts WHERE id='22222222-2222-4222-8222-222222222222';DELETE FROM product_media WHERE id=900003;DELETE FROM media_assets WHERE id=900003;DELETE FROM site_logo_assets;DELETE FROM site_favicon_assets;`)
				require.NoError(t, err)
			}
			for _, name := range snapshot.RequiredFilePaths {
				if !strings.HasPrefix(name, "media/") {
					continue
				}
				root, key, err := transactionalBusinessModuleRestoreMediaPath(name)
				require.NoError(t, err)
				require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots[root], filepath.FromSlash(key))))
			}
			require.NoError(t, os.Remove(orphan))
			if repairScenario {
				for _, entry := range []struct{ root, key string }{{"uploads", "products/media-original.png"}, {"uploads_private", "customer-service/attachments/originals/media-private.png"}, {"uploads", "media-derivatives/900001/thumbnail/v1/public.png"}, {"uploads_private", "customer-service/attachments/derivatives/900002/thumbnail/v1/private.png"}} {
					filename := filepath.Join(cfg.StorageRoots[entry.root], filepath.FromSlash(entry.key))
					require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
					require.NoError(t, os.WriteFile(filename, []byte("damaged bytes before repair"), 0600))
				}
			}
			if scenario == "repair_changed_registration_conflict" {
				_, err = database.ExecContext(ctx, `UPDATE media_assets SET caption='Current media metadata must remain' WHERE id=900001`)
				require.NoError(t, err)
			}
			if scenario == "repair_unregistered_file_conflict" {
				require.NoError(t, os.WriteFile(orphan, []byte("current unrelated file"), 0600))
			}
			if scenario == "different_registration_contents" {
				_, err = database.ExecContext(ctx, `INSERT INTO media_assets(id,filename,url,storage_key,size,mime_type,visibility,status) VALUES(900001,'other.png','/uploads/other.png','other.png',21,'image/png','public','active')`)
				require.NoError(t, err)
			}
			if scenario == "company_release_mismatch" {
				require.NoError(t, os.WriteFile(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", "fixture.png"), []byte("different deployed release bytes"), 0600))
			}
			if scenario == "retained_legacy_private_owner_conflict" {
				_, err = database.ExecContext(ctx, `INSERT INTO tickets(id,ticket_number,conversation_id,user_id,subject,category) VALUES(900010,'foreign-media-chat','foreign-media-chat',900002,'Other retained conversation','customer_service');INSERT INTO ticket_messages(ticket_id,content,attachments) VALUES(900010,'Must not acquire access','["/uploads/customer-service/attachments/originals/media-private.png"]')`)
				require.NoError(t, err)
			}
			if scenario == "different_file_bytes" {
				require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "products/media-original.png"), []byte("current distinct bytes"), 0600))
			}
			if scenario == "live_cleanup_hazard" {
				_, err = database.ExecContext(ctx, `INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES('media-cleanup-test','storage.object_cleanup_requested','media_asset','900002','{"resource_type":"media_asset","resource_id":"900002","object_keys":["customer-service/attachments/originals/media-private.png"]}','pending')`)
				require.NoError(t, err)
			}
			if scenario == "private_retention_cleanup_hazard" {
				_, err = database.ExecContext(ctx, `INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES('media-retention-test','customer_service.retention_cleanup_requested','customer_service_conversation','900001','{"ticket_id":900001,"attachment_references":["/uploads/customer-service/attachments/originals/media-private.png"]}','processed')`)
				require.NoError(t, err)
			}
			_, beforeDatabase, _, err := captureDatabaseInventory(ctx, database)
			require.NoError(t, err)
			beforeMedia, err := cfg.mediaFingerprint(ctx)
			require.NoError(t, err)
			require.NoError(t, database.Close())
			_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID)
			require.NoError(t, err)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			blocked := strings.Contains(scenario, "hazard") || scenario == "different_file_bytes" || scenario == "wrong_private_actor" || scenario == "unfinished_upload" || scenario == "wrong_product_variant_owner" || scenario == "different_registration_contents" || scenario == "company_release_mismatch" || scenario == "retained_legacy_private_owner_conflict" || scenario == "repair_missing_registration_conflict" || scenario == "repair_changed_registration_conflict" || scenario == "repair_unregistered_file_conflict"
			if blocked {
				require.Equal(t, "module_restore_conflicts_found", task.State, task.Error)
				require.Error(t, validateMediaModuleFormalRestorePreview(task, task.SelectedModuleRestorePreview.SHA256))
				actualMedia, err := cfg.mediaFingerprint(ctx)
				require.NoError(t, err)
				require.Equal(t, beforeMedia, actualMedia)
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, actualDatabase, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, beforeDatabase, actualDatabase)
				require.NoError(t, database.Close())
				return
			}
			summary, _ := json.Marshal(task.SelectedModuleRestorePreview)
			require.Equal(t, "module_isolated_restore_verified", task.State, "%s %s", task.Error, string(summary))
			require.True(t, task.SelectedModuleRestorePreview.MediaRestoreEligibility.Eligible)
			expectedRows := 7
			if repairScenario {
				expectedRows = 0
				require.Equal(t, 4, task.SelectedModuleRestorePreview.MediaRestoreEligibility.ExistingFileRepairs)
				require.NotEmpty(t, task.SelectedModuleRestorePreview.MediaRestoreEligibility.ExistingFileRepairPlanSHA256)
			}
			if scenario == "all_eleven_media_tables" {
				expectedRows = 15
			}
			require.Equal(t, expectedRows, task.SelectedModuleRestorePreview.RestoredRows)
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-media-restorer", []string{PermissionRestore})
			require.NoError(t, err)
			invokeMediaModuleRestoreHTTPAction := func(action, credential string, body any, status int) {
				t.Helper()
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+task.ID+"/"+action, strings.NewReader(string(encoded)))
				request.Header.Set("Authorization", "Bearer "+credential)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, status, response.Code, response.Body.String())
			}
			request := MediaModuleFormalRestoreRequest{Confirmation: "RESTORE MEDIA " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256}
			for _, permission := range []string{PermissionView, PermissionUpload, PermissionExport} {
				unauthorized, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-limited", []string{permission})
				require.NoError(t, err)
				invokeMediaModuleRestoreHTTPAction("media-module-formal-restore", unauthorized, request, http.StatusForbidden)
			}
			stale := request
			stale.ExpectedPreviewSHA256 = strings.Repeat("0", 64)
			invokeMediaModuleRestoreHTTPAction("media-module-formal-restore", token, stale, http.StatusConflict)
			if scenario == "stale_database_after_preview" || scenario == "repair_after_preview_changed_bytes" {
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				if scenario == "stale_database_after_preview" {
					_, err = database.ExecContext(ctx, `UPDATE orders SET admin_note='changed after preview' WHERE id=900001`)
					require.NoError(t, err)
				} else {
					require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "products/media-original.png"), []byte("new bytes since preview"), 0600))
				}
				require.NoError(t, database.Close())
				invokeMediaModuleRestoreHTTPAction("media-module-formal-restore", token, request, http.StatusOK)
				task = waitForIndependentRecoveryTask(t, manager, task.ID)
				require.Equal(t, "failed", task.State)
				require.Nil(t, task.MediaModuleRestore)
				return
			}
			authenticated := scenario == "authenticated_restore" || scenario == "all_eleven_media_tables" || scenario == "repair_public_and_private_files"
			if authenticated {
				invokeMediaModuleRestoreHTTPAction("media-module-formal-restore", token, request, http.StatusOK)
			} else {
				require.NoError(t, manager.startTaskOperation(task, "synthetic_media_restore", func(ctx context.Context, task *Task) error {
					return manager.restoreMediaModuleWithAtomicReceiptAndRecoverableFiles(ctx, task, "synthetic-media-restorer", func(point string) error {
						if point == strings.TrimPrefix(scenario, "repair_") {
							return errors.New("synthetic interrupted media restore")
						}
						return nil
					})
				}))
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			committed := authenticated || scenario == "rows_committed_before_acknowledgement" || scenario == "verified_before_fence_removal" || scenario == "fence_removed_before_acknowledgement" || scenario == "repair_rows_committed_before_acknowledgement"
			if !authenticated {
				require.Equal(t, "failed", task.State, task.Error)
				require.Equal(t, "synthetic interrupted media restore", task.Error)
				require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, manager.ExpireWorkCopies(ctx))
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				if scenario != "fence_removed_before_acknowledgement" {
					invokeMediaModuleRestoreHTTPAction("media-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "RECOVER MEDIA " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, http.StatusOK)
					task = waitForIndependentRecoveryTask(t, manager, task.ID)
				}
			}
			require.Equal(t, map[bool]string{true: "completed", false: "rolled_back"}[committed], task.State, task.Error)
			require.NoError(t, validateTransactionalBusinessModuleRestoreProgress(task))
			require.Equal(t, "released", task.MediaModuleRestore.Phase)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer database.Close()
			if !committed {
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, beforeDatabase, actual)
				actualMedia, err := cfg.mediaFingerprint(ctx)
				require.NoError(t, err)
				require.Equal(t, beforeMedia, actualMedia)
				return
			}
			require.Equal(t, expectedRows, task.MediaModuleRestore.RestoredRows)
			var audits int
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM audit_logs WHERE resource='media_module_formal_restore'`).Scan(&audits))
			require.Equal(t, 1, audits)
			if repairScenario {
				var repairedFiles int
				require.NoError(t, database.QueryRowContext(ctx, `SELECT (new_value::jsonb->>'repaired_files')::int FROM audit_logs WHERE resource='media_module_formal_restore'`).Scan(&repairedFiles))
				require.Equal(t, 4, repairedFiles)
			}
			for _, name := range snapshot.RequiredFilePaths {
				if !strings.HasPrefix(name, "media/") {
					continue
				}
				filename, err := cfg.resolveSelectedBusinessModuleConfinedFileSource(name)
				require.NoError(t, err)
				body, err := os.ReadFile(filename)
				require.NoError(t, err)
				require.Equal(t, "synthetic-media-file!", string(body))
				if strings.HasPrefix(name, "media/uploads_private/") {
					info, err := os.Stat(filename)
					require.NoError(t, err)
					require.Equal(t, os.FileMode(0600), info.Mode().Perm())
				}
			}
			body, err := os.ReadFile(orphan)
			require.NoError(t, err)
			require.Equal(t, "unregistered original bytes", string(body))
			var notes, chat string
			require.NoError(t, database.QueryRowContext(ctx, `SELECT admin_note FROM orders WHERE id=900001`).Scan(&notes))
			require.Equal(t, "New order note must remain", notes)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&chat))
			require.Equal(t, "Preserved conversation and original ownership", chat)
			verifyRestoredMediaModulePublicImagesAndPrivateAttachmentAuthorization(t, cfg, database)
		})
	}
}
