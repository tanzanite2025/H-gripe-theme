package localbackup

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLSelectedBusinessModulesExportOnlyChosenFullRowsAndActualFilesWithoutDatabaseWrites(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and all application migrations")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_test_"+newTaskID()[:12]
	cfg.ApplicationVersion = "sha-" + strings.Repeat("a", 40)
	cfg.ApplicationDatabaseRole = "synthetic_module_app_" + newTaskID()[:12]
	cfg.MaxFiles, cfg.MaxExpandedBytes, cfg.MaxWorkspaceBytes = 10000, 256<<20, 1<<30
	cfg.ProductStaticAssetsDirectory = filepath.Join(t.TempDir(), "static")
	require.NoError(t, os.MkdirAll(filepath.Join(cfg.ProductStaticAssetsDirectory, "company"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename), []byte(cfg.ApplicationVersion+"\n"), 0600))
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
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(db, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
	require.NoError(t, err)
	err = migrator.Up()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	_, err = migrator.Close()
	require.NoError(t, err)
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	_, err = db.ExecContext(ctx, `DO $$ DECLARE customer bigint; staff bigint; product bigint; variant bigint; conversation bigint; other_ticket bigint; asset bigint; message bigint; order_id bigint; BEGIN
INSERT INTO users(email,username,password,role) VALUES('module-customer@example.invalid','module-customer','synthetic-customer-password','user') RETURNING id INTO customer;
INSERT INTO users(email,username,password,role) VALUES('module-staff@example.invalid','module-staff','synthetic-admin-password','admin') RETURNING id INTO staff;
INSERT INTO social_oauth_connections(provider,access_token_encrypted) VALUES('facebook','synthetic-global-connector-token');
INSERT INTO products(name,slug,price_minor,currency) VALUES('Selected module product','module-product',123,'USD') RETURNING id INTO product;
INSERT INTO product_variants(product_id,sku,price_minor,currency,stock) VALUES(product,'module-sku',123,'USD',7) RETURNING id INTO variant;
INSERT INTO media_assets(filename,url,storage_key,size,visibility) VALUES('photo.bin','/uploads/products/module-photo.bin','products/module-photo.bin',4,'public') RETURNING id INTO asset;
INSERT INTO product_media(product_id,variant_id,media_asset_id,media_type,url) VALUES(product,variant,asset,'image','/uploads/products/module-photo.bin');
INSERT INTO tickets(ticket_number,user_id,customer_user_id,conversation_id,subject,category) VALUES('module-chat',customer,customer,'module-chat','Selected conversation','customer_service') RETURNING id INTO conversation;
INSERT INTO ticket_messages(ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(conversation,customer,'Retained deleted customer text','[]','media_asset_bindings') RETURNING id INTO message;
UPDATE tickets SET deleted_at=now() WHERE id=conversation;
INSERT INTO tickets(ticket_number,user_id,subject,category) VALUES('unselected-ticket',customer,'Unselected support case','general') RETURNING id INTO other_ticket;
INSERT INTO ticket_messages(ticket_id,user_id,content,attachments) VALUES(other_ticket,customer,'Unselected other ticket content','[]');
INSERT INTO media_assets(filename,url,storage_key,size,visibility,library_partition,customer_service_media_purpose,uploader_id,uploader_actor_type,classification_method) VALUES('chat.bin','/api/customer-service/attachments/1','customer-service/attachments/originals/module-chat.bin',4,'private','customer_service','conversation_attachment',customer,'customer','recorded_upload_context') RETURNING id INTO asset;
INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,object_key,media_asset_id,status,expires_at) VALUES('synthetic-module-chat-upload','customer_service',conversation,'customer',customer,'customer-service/attachments/originals/module-chat.bin',asset,'bound',now()+interval '1 day');
INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(message,asset,0,'synthetic-module-chat-upload');
INSERT INTO customer_service_inbox_states(ticket_id,recipient_user_id,last_read_message_id,assignment_version,unread_count) VALUES(conversation,staff,0,1,0);
INSERT INTO orders(order_number,user_id,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES('module-order',customer,'USD','USD',123,123,123,0,0,0,0) RETURNING id INTO order_id;
INSERT INTO order_items(order_id,product_id,variant_id,product_name,quantity,currency,price_minor,subtotal_minor,total_minor,tax_amount_minor,discount_minor) VALUES(order_id,product,variant,'Full historical product snapshot',1,'USD',123,123,123,0,0);
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('module-order-completed','order.completed','order',order_id::text,'{}','processed',now());
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('module-chat-realtime','customer_service.realtime','customer_service_conversation',conversation::text,jsonb_build_object('ticket_id',conversation,'conversation_id','module-chat','type','conversation.message.created','event_id','module-chat-realtime','payload',jsonb_build_object('message_id',message)),'processed',now());
END $$`)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "unreferenced-media-only.bin"), []byte("data"), 0600))
	_, before, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	allModules := []string{}
	for _, option := range availableBusinessModuleBackupOptions() {
		allModules = append(allModules, option.ID)
	}
	sort.Strings(allModules)
	for _, modules := range append([][]string{{"products"}, {"media"}, {"orders"}, {"customers"}, {"customer_service"}, {"configuration"}, {"customer_service", "orders"}}, allModules) {
		t.Run(strings.Join(modules, "_"), func(t *testing.T) {
			manager := &Manager{Configuration: cfg}
			preview, err := manager.PreviewSelectedBusinessModuleBackup(ctx, modules)
			require.NoError(t, err)
			require.False(t, preview.OnlineOverwriteAvailable)
			tx, err := db.BeginTx(ctx, nil)
			require.NoError(t, err)
			snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, modules)
			require.NoError(t, err)
			require.NoError(t, tx.Rollback())
			serialized, err := json.Marshal(snapshot)
			require.NoError(t, err)
			selected := map[string]bool{}
			for _, module := range modules {
				selected[module] = true
			}
			if !selected["customers"] {
				require.NotContains(t, snapshot.Tables, "users")
				require.NotContains(t, string(serialized), "synthetic-customer-password")
			} else {
				require.Contains(t, string(serialized), "synthetic-customer-password")
				require.NotContains(t, string(serialized), "synthetic-admin-password")
			}
			if !selected["orders"] {
				require.NotContains(t, snapshot.Tables, "orders")
			}
			if !selected["configuration"] {
				require.NotContains(t, string(serialized), "synthetic-global-connector-token")
			}
			if !selected["customer_service"] {
				require.NotContains(t, snapshot.Tables, "ticket_messages")
			} else {
				require.Contains(t, string(serialized), "Retained deleted customer text")
				require.NotContains(t, string(serialized), "Unselected other ticket content")
				require.Contains(t, snapshot.RequiredFilePaths, "media/uploads_private/customer-service/attachments/originals/module-chat.bin")
			}
			for _, path := range snapshot.RequiredFilePaths {
				var filename string
				for _, root := range mediaRoots {
					if strings.HasPrefix(path, "media/"+root+"/") {
						filename = filepath.Join(cfg.StorageRoots[root], filepath.FromSlash(strings.TrimPrefix(path, "media/"+root+"/")))
					}
				}
				if strings.HasPrefix(path, "product-static/") {
					filename = filepath.Join(cfg.ProductStaticAssetsDirectory, filepath.FromSlash(strings.TrimPrefix(path, "product-static/")))
				}
				require.NotEmpty(t, filename, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					require.NoError(t, os.WriteFile(filename, []byte("data"), 0600))
				}
			}
		})
	}
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	exporter, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-module-operator", []string{PermissionExport})
	require.NoError(t, err)
	restorer, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-module-restorer", []string{PermissionRestore})
	require.NoError(t, err)
	var chatUpload *Task
	for _, modules := range [][]string{{"orders"}, {"customer_service"}, {"customers"}, {"configuration"}, {"media"}, {"customer_service", "orders"}, allModules} {
		phrase := "BACKUP MODULES " + cfg.SiteID + "/" + cfg.Environment + "/" + strings.Join(modules, ",")
		body, err := json.Marshal(map[string]any{"modules": modules, "confirmation": phrase})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/module-backups", bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+exporter)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var created struct {
			Task *Task `json:"task"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
		backup := created.Task
		backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
		require.Equal(t, "ready_to_download", backup.State, backup.Error)
		includesAllMedia, foundUnreferencedFile := false, false
		for _, module := range modules {
			includesAllMedia = includesAllMedia || module == "media"
		}
		for _, file := range backup.Manifest.Files {
			foundUnreferencedFile = foundUnreferencedFile || file.Path == "media/uploads_private/unreferenced-media-only.bin"
		}
		require.Equal(t, includesAllMedia, foundUnreferencedFile)
		request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+backup.ID+"/download-ticket", strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer "+exporter)
		response = httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var ticket struct {
			DownloadPath string `json:"download_path"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &ticket))
		require.NotEmpty(t, ticket.DownloadPath)
		downloadRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, ticket.DownloadPath, nil)
		for _, cookie := range response.Result().Cookies() {
			downloadRequest.AddCookie(cookie)
		}
		response = httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, downloadRequest)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		require.Contains(t, response.Header().Get("Content-Disposition"), "storefront-modules-")
		upload := uploadProductCatalogPackageForIntegrationTest(t, manager, response.Body.Bytes())
		require.Equal(t, "module_package_integrity_verified", upload.State, upload.Error)
		require.Equal(t, modules, upload.Manifest.SelectedModules)
		require.Empty(t, upload.StageDatabase)
		require.Nil(t, upload.Journal)
		phrase = "VERIFY MODULE RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + upload.ID
		body, err = json.Marshal(map[string]string{"confirmation": phrase})
		require.NoError(t, err)
		request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+upload.ID+"/selected-module-isolated-restore-preview", bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+restorer)
		response = httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		upload = waitForIndependentRecoveryTask(t, manager, upload.ID)
		reportData, _ := json.Marshal(upload.SelectedModuleRestorePreview)
		require.Equal(t, "module_isolated_restore_verified", upload.State, "%s: %s", upload.Error, string(reportData))
		require.True(t, upload.SelectedModuleRestorePreview.IsolatedRestoreVerified)
		require.True(t, upload.SelectedModuleRestorePreview.TargetUnchanged)
		require.True(t, upload.SelectedModuleRestorePreview.DisposableStageRemoved)
		require.False(t, upload.SelectedModuleRestorePreview.OnlineOverwriteAvailable)
		require.Zero(t, upload.SelectedModuleRestorePreview.RestoredRows)
		require.Empty(t, upload.StageDatabase)
		_, err = manager.ApplyVerifiedRecovery(upload.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+upload.Manifest.PackageID)
		require.Error(t, err, "module isolation cannot authorize whole-storefront apply")
		if len(modules) == 1 && modules[0] == "customer_service" {
			chatUpload = upload
		}
	}
	require.NotNil(t, chatUpload)
	verifyChatPreview := func() *Task {
		t.Helper()
		_, err := manager.StartSelectedBusinessModuleIsolatedRestorePreview(chatUpload.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+chatUpload.ID)
		require.NoError(t, err)
		return waitForIndependentRecoveryTask(t, manager, chatUpload.ID)
	}
	postChatRecoveryAction := func(task *Task, action string, payload any) (*Task, int, string) {
		t.Helper()
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+task.ID+"/"+action, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+restorer)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		var result struct {
			Task *Task `json:"task"`
		}
		if response.Code == http.StatusOK {
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
		}
		return result.Task, response.Code, response.Body.String()
	}
	saveChatChoice := func(task *Task, policy string) *Task {
		t.Helper()
		result, status, body := postChatRecoveryAction(task, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": task.SelectedModuleRestorePreview.SHA256, "policy": policy})
		require.Equal(t, http.StatusOK, status, body)
		require.NotNil(t, result.SelectedCustomerServiceRestoreChoice)
		require.Equal(t, "synthetic-module-restorer", result.SelectedCustomerServiceRestoreChoice.Operator)
		require.False(t, result.SelectedCustomerServiceRestoreChoice.ExecutionAvailable)
		require.Nil(t, result.SelectedCustomerServiceCandidateRehearsal)
		require.NotEqual(t, "module_chat_candidate_rehearsal_verified", result.State, "saving a new choice invalidates the previous candidate result and success state")
		return result
	}
	rehearseChatChoice := func(task *Task) *Task {
		t.Helper()
		digest := task.SelectedCustomerServiceRestoreChoice.SHA256
		_, status, body := postChatRecoveryAction(task, "selected-customer-service-candidate-rehearsal", map[string]string{"choice_sha256": digest, "confirmation": "VERIFY CHAT CHOICE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID + "/" + digest})
		require.Equal(t, http.StatusOK, status, body)
		result := waitForIndependentRecoveryTask(t, manager, task.ID)
		require.Equal(t, "module_chat_candidate_rehearsal_verified", result.State, result.Error)
		require.NotNil(t, result.SelectedCustomerServiceCandidateRehearsal)
		require.Equal(t, digest, result.SelectedCustomerServiceCandidateRehearsal.SavedChoiceSHA256)
		require.True(t, result.SelectedCustomerServiceCandidateRehearsal.IsolatedRestoreVerified)
		require.True(t, result.SelectedCustomerServiceCandidateRehearsal.TargetUnchanged)
		require.True(t, result.SelectedCustomerServiceCandidateRehearsal.DisposableStageRemoved)
		require.False(t, result.SelectedCustomerServiceCandidateRehearsal.OnlineOverwriteAvailable)
		require.NotEmpty(t, result.SelectedCustomerServiceCandidateRehearsal.CandidateDatabaseFingerprint)
		require.Empty(t, result.StageDatabase)
		require.Nil(t, result.SelectedModuleDisposableStage)
		require.Nil(t, result.Journal)
		_, err := manager.ApplyVerifiedRecovery(result.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+result.Manifest.PackageID)
		require.Error(t, err, "a disposable chat candidate must never authorize online apply")
		return result
	}
	for _, scenario := range []struct {
		name, mutateSQL, cleanupSQL, issue string
		preservedStrategySatisfied         bool
	}{
		{"status_version_rewind", `UPDATE tickets SET status_version=6 WHERE ticket_number='module-chat'`, `UPDATE tickets SET status_version=1 WHERE ticket_number='module-chat'`, "realtime_version_counter_would_rewind", true},
		{"assignment_version_rewind", `UPDATE customer_service_inbox_states SET assignment_version=6`, `UPDATE customer_service_inbox_states SET assignment_version=1`, "realtime_version_counter_would_rewind", true},
		{"read_cursor_rewind", `UPDATE customer_service_inbox_states SET last_read_message_id=(SELECT m.id FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.ticket_number='module-chat')`, `UPDATE customer_service_inbox_states SET last_read_message_id=0`, "realtime_read_event_identity_would_be_reused", true},
		{"message_primary_key_moved_to_another_conversation", `UPDATE ticket_messages SET ticket_id=(SELECT id FROM tickets WHERE ticket_number='unselected-ticket') WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`, `UPDATE ticket_messages SET ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat') WHERE content='Retained deleted customer text'`, "module_conversation_row_ownership_conflict", false},
		{"preserved_ignored_soft_deleted_event_key_ahead", `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,deleted_at) SELECT 900002,'customer_service.conversation.status.changed:'||id||':4','customer_service.realtime','customer_service_conversation',id::text,jsonb_build_object('ticket_id',id,'conversation_id',conversation_id,'type','conversation.status.changed','event_id','customer_service.conversation.status.changed:'||id||':4','payload',jsonb_build_object('status_version',4)),'ignored',now() FROM tickets WHERE ticket_number='module-chat'`, `DELETE FROM outbox_events WHERE event_key LIKE 'customer_service.conversation.status.changed:%'`, "realtime_preserved_event_version_ahead_of_effective_row", false},
	} {
		t.Run("chat_domain_"+scenario.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, scenario.mutateSQL)
			require.NoError(t, err)
			t.Cleanup(func() { _, err := db.ExecContext(ctx, scenario.cleanupSQL); require.NoError(t, err) })
			_, expectedTarget, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			result := verifyChatPreview()
			require.Equal(t, "module_restore_conflicts_found", result.State, result.Error)
			coverage := result.SelectedModuleRestorePreview.CustomerServiceOverwritePreview
			require.NotNil(t, coverage)
			require.True(t, coverage.PrerequisitesSatisfied)
			require.NotNil(t, coverage.PackageRows)
			require.Positive(t, coverage.PackageRows.IssueCounts[scenario.issue], "%+v", coverage.PackageRows)
			require.False(t, coverage.PackageRows.DomainChecksSatisfied)
			require.Equal(t, scenario.preservedStrategySatisfied, coverage.PreserveCurrentRows.DomainChecksSatisfied, "%+v", coverage.PreserveCurrentRows)
			require.False(t, coverage.ExecutionAvailable)
			require.True(t, result.SelectedModuleRestorePreview.TargetUnchanged)
			require.True(t, result.SelectedModuleRestorePreview.DisposableStageRemoved)
			_, actualTarget, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			require.Equal(t, expectedTarget, actualTarget)
			encoded, err := json.Marshal(coverage)
			require.NoError(t, err)
			for _, private := range []string{"Retained deleted customer text", "synthetic-module-chat-upload", "module-customer@example.invalid", "customer-service/attachments/originals/module-chat.bin", "customer_service.conversation.status.changed:"} {
				require.NotContains(t, string(encoded), private)
			}
			_, status, _ := postChatRecoveryAction(result, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": result.SelectedModuleRestorePreview.SHA256, "policy": "use_package_conversation_rows"})
			require.NotEqual(t, http.StatusOK, status, "a failed domain strategy cannot be saved")
			if scenario.name == "status_version_rewind" {
				result = rehearseChatChoice(saveChatChoice(result, "preserve_current_conflicting_conversation_rows"))
				require.Equal(t, expectedTarget, result.SelectedCustomerServiceCandidateRehearsal.CandidateDatabaseFingerprint)
			}
		})
	}
	t.Run("chat_domain_preserves_target_only_message", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) SELECT 900001,id,user_id,'Synthetic current-only message','[]','legacy_urls' FROM tickets WHERE ticket_number='module-chat'`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, err := db.ExecContext(ctx, `DELETE FROM ticket_messages WHERE content='Synthetic current-only message'`)
			require.NoError(t, err)
		})
		result := verifyChatPreview()
		require.Equal(t, "module_isolated_restore_verified", result.State, result.Error)
		coverage := result.SelectedModuleRestorePreview.CustomerServiceOverwritePreview
		require.True(t, coverage.PackageRows.DomainChecksSatisfied, "%+v", coverage.PackageRows)
		require.Equal(t, 1, coverage.PackageRows.TargetOnlyRowsPreserved)
		require.Equal(t, 1, coverage.PreserveCurrentRows.TargetOnlyRowsPreserved)
		_, expectedTarget, _, err := captureDatabaseInventory(ctx, db)
		require.NoError(t, err)
		result = rehearseChatChoice(saveChatChoice(result, "use_package_conversation_rows"))
		require.Equal(t, expectedTarget, result.SelectedCustomerServiceCandidateRehearsal.CandidateDatabaseFingerprint, "current-only message and all event facts must survive the candidate")
	})
	for _, scenario := range []struct{ name, mutateSQL, cleanupSQL, issue string }{
		{"private_upload_registered_actor_conflict", `UPDATE media_assets SET uploader_id=(SELECT id FROM users WHERE email='module-staff@example.invalid') WHERE visibility='private'`, `UPDATE media_assets SET uploader_id=(SELECT id FROM users WHERE email='module-customer@example.invalid') WHERE visibility='private'`, "upload_receipt_asset_actor_identity_conflict"},
		{"attachment_upload_receipt_belongs_to_another_conversation", `UPDATE media_upload_sessions SET business_context_id=(SELECT id FROM tickets WHERE ticket_number='unselected-ticket')`, `UPDATE media_upload_sessions SET business_context_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`, "attachment_binding_conversation_conflict"},
	} {
		t.Run("chat_domain_authenticated_package_"+scenario.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, scenario.mutateSQL)
			require.NoError(t, err)
			t.Cleanup(func() { _, err := db.ExecContext(ctx, scenario.cleanupSQL); require.NoError(t, err) })
			backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-chat-domain-operator", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/customer_service", []string{"customer_service"})
			require.NoError(t, err)
			backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
			require.Equal(t, "ready_to_download", backup.State, backup.Error)
			packagePath, err := manager.taskPackagePath(backup.ID)
			require.NoError(t, err)
			packageBytes, err := os.ReadFile(packagePath)
			require.NoError(t, err)
			upload := uploadProductCatalogPackageForIntegrationTest(t, manager, packageBytes)
			require.Equal(t, "module_package_integrity_verified", upload.State, upload.Error)
			_, expectedTarget, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(upload.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+upload.ID)
			require.NoError(t, err)
			result := waitForIndependentRecoveryTask(t, manager, upload.ID)
			require.Equal(t, "module_restore_conflicts_found", result.State, result.Error)
			coverage := result.SelectedModuleRestorePreview.CustomerServiceOverwritePreview
			require.NotNil(t, coverage)
			require.True(t, coverage.PrerequisitesSatisfied)
			for _, strategy := range []*SelectedCustomerServiceModuleOverwriteStrategyInspection{coverage.PackageRows, coverage.PreserveCurrentRows} {
				require.NotNil(t, strategy)
				require.False(t, strategy.DomainChecksSatisfied)
				require.Positive(t, strategy.IssueCounts[scenario.issue], "%+v", strategy)
			}
			require.True(t, result.SelectedModuleRestorePreview.DisposableStageRemoved)
			_, actualTarget, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			require.Equal(t, expectedTarget, actualTarget)
		})
	}
	t.Run("restore_missing_chat_full_row_and_private_file_only_in_disposable_clone", func(t *testing.T) {
		var original string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT to_jsonb(m)::text FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.ticket_number='module-chat'`).Scan(&original))
		// Keep the selected identity and relationships, but remove the leaf binding
		// and message from the live synthetic target to exercise real insertion.
		var binding string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT to_jsonb(b)::text FROM ticket_message_media_attachments b JOIN ticket_messages m ON m.id=b.ticket_message_id JOIN tickets t ON t.id=m.ticket_id WHERE t.ticket_number='module-chat'`).Scan(&binding))
		_, err := db.ExecContext(ctx, `DELETE FROM ticket_message_media_attachments WHERE ticket_message_id IN (SELECT m.id FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.ticket_number='module-chat'); DELETE FROM ticket_messages WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, err := db.ExecContext(ctx, `INSERT INTO ticket_messages SELECT * FROM jsonb_populate_record(NULL::ticket_messages,$1::jsonb)`, original)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, `INSERT INTO ticket_message_media_attachments SELECT * FROM jsonb_populate_record(NULL::ticket_message_media_attachments,$1::jsonb)`, binding)
			require.NoError(t, err)
		})
		privateFile := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service", "attachments", "originals", "module-chat.bin")
		privateBytes, err := os.ReadFile(privateFile)
		require.NoError(t, err)
		require.NoError(t, os.Remove(privateFile))
		t.Cleanup(func() { require.NoError(t, os.WriteFile(privateFile, privateBytes, 0600)) })
		_, changedTarget, _, err := captureDatabaseInventory(ctx, db)
		require.NoError(t, err)
		result := verifyChatPreview()
		require.Equal(t, "module_isolated_restore_verified", result.State, result.Error)
		require.Equal(t, 2, result.SelectedModuleRestorePreview.RestoredRows)
		require.True(t, result.SelectedModuleRestorePreview.DisposableStageRemoved)
		_, afterPreview, _, err := captureDatabaseInventory(ctx, db)
		require.NoError(t, err)
		require.Equal(t, changedTarget, afterPreview)
		_, err = os.Stat(privateFile)
		require.True(t, os.IsNotExist(err), "the live private file must stay absent")
		result = rehearseChatChoice(saveChatChoice(result, "use_package_conversation_rows"))
		require.Equal(t, 2, result.SelectedCustomerServiceCandidateRehearsal.RestoredRows)
		require.Zero(t, result.SelectedCustomerServiceCandidateRehearsal.ReplacedRows)
		_, afterCandidate, _, err := captureDatabaseInventory(ctx, db)
		require.NoError(t, err)
		require.Equal(t, changedTarget, afterCandidate)
		_, err = os.Stat(privateFile)
		require.True(t, os.IsNotExist(err), "candidate files must never be copied into the live root")
	})
	t.Run("changed_message_blocks_all_module_writes", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `UPDATE ticket_messages SET content='newer live content' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, err := db.ExecContext(ctx, `UPDATE ticket_messages SET content='Retained deleted customer text' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`)
			require.NoError(t, err)
		})
		result := verifyChatPreview()
		require.Equal(t, "module_restore_conflicts_found", result.State, result.Error)
		require.False(t, result.SelectedModuleRestorePreview.IsolatedRestoreVerified)
		require.True(t, result.SelectedModuleRestorePreview.TargetUnchanged)
		found := false
		for _, table := range result.SelectedModuleRestorePreview.Tables {
			found = found || (table.Table == "ticket_messages" && table.ConflictingRows == 1)
		}
		require.True(t, found)
		_, expectedCurrent, _, err := captureDatabaseInventory(ctx, db)
		require.NoError(t, err)
		for _, policy := range []string{"use_package_conversation_rows", "preserve_current_conflicting_conversation_rows"} {
			result = rehearseChatChoice(saveChatChoice(result, policy))
			candidate := result.SelectedCustomerServiceCandidateRehearsal
			if policy == "use_package_conversation_rows" {
				require.Equal(t, 1, candidate.ReplacedRows)
				require.Equal(t, before, candidate.CandidateDatabaseFingerprint, "complete package body must replace only the conflicting chat row")
			} else {
				require.Zero(t, candidate.ReplacedRows)
				require.Equal(t, expectedCurrent, candidate.CandidateDatabaseFingerprint, "preserve-current policy must retain the complete current row")
			}
			_, actualCurrent, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			require.Equal(t, expectedCurrent, actualCurrent)
		}
	})
	t.Run("chat_saved_strategy_invalidates_on_body_sequence_event_media_configuration_or_package_change", func(t *testing.T) {
		result := verifyChatPreview()
		require.Equal(t, "module_isolated_restore_verified", result.State, result.Error)
		result = saveChatChoice(result, "use_package_conversation_rows")
		privateFile := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service", "attachments", "originals", "module-chat.bin")
		for _, scenario := range []struct{ name, mutateSQL, cleanupSQL string }{
			{"body", `UPDATE ticket_messages SET content='changed after saved choice' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`, `UPDATE ticket_messages SET content='Retained deleted customer text' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='module-chat')`},
			{"event", `UPDATE outbox_events SET status='ignored' WHERE event_key='module-chat-realtime'`, `UPDATE outbox_events SET status='processed' WHERE event_key='module-chat-realtime'`},
			{"sequence", `SELECT setval('ticket_messages_id_seq',800000,true)`, ""},
		} {
			var sequenceValue int64
			var sequenceCalled bool
			if scenario.name == "sequence" {
				require.NoError(t, db.QueryRowContext(ctx, `SELECT last_value,is_called FROM ticket_messages_id_seq`).Scan(&sequenceValue, &sequenceCalled))
			}
			_, err := db.ExecContext(ctx, scenario.mutateSQL)
			require.NoError(t, err)
			_, status, body := postChatRecoveryAction(result, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": result.SelectedModuleRestorePreview.SHA256, "policy": "use_package_conversation_rows"})
			require.NotEqual(t, http.StatusOK, status, scenario.name)
			require.Contains(t, body, "已变化")
			if scenario.name == "body" {
				digest := result.SelectedCustomerServiceRestoreChoice.SHA256
				_, status, body = postChatRecoveryAction(result, "selected-customer-service-candidate-rehearsal", map[string]string{"choice_sha256": digest, "confirmation": "VERIFY CHAT CHOICE " + cfg.SiteID + "/" + cfg.Environment + "/" + result.ID + "/" + digest})
				require.Equal(t, http.StatusOK, status, body)
				failed := waitForIndependentRecoveryTask(t, manager, result.ID)
				require.Equal(t, "failed", failed.State)
				require.Contains(t, failed.Error, "已变化")
				require.Nil(t, failed.SelectedModuleDisposableStage)
				// Restore only the synthetic task state so subsequent mutations test
				// the same immutable saved snapshot, rather than the failed-state gate.
				failed.State = result.State
				require.NoError(t, manager.saveTask(failed))
			}
			if scenario.name == "sequence" {
				_, err = db.ExecContext(ctx, `SELECT setval('ticket_messages_id_seq',$1,$2)`, sequenceValue, sequenceCalled)
			} else {
				_, err = db.ExecContext(ctx, scenario.cleanupSQL)
			}
			require.NoError(t, err)
		}
		original, err := os.ReadFile(privateFile)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(privateFile, []byte("changed after saved choice"), 0600))
		_, status, body := postChatRecoveryAction(result, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": result.SelectedModuleRestorePreview.SHA256, "policy": "use_package_conversation_rows"})
		require.NotEqual(t, http.StatusOK, status)
		require.Contains(t, body, "已变化")
		require.NoError(t, os.WriteFile(privateFile, original, 0600))
		manager.Configuration.PublicOrigins = append(manager.Configuration.PublicOrigins, "https://synthetic-other.invalid")
		_, status, body = postChatRecoveryAction(result, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": result.SelectedModuleRestorePreview.SHA256, "policy": "use_package_conversation_rows"})
		require.NotEqual(t, http.StatusOK, status)
		require.Contains(t, body, "已变化")
		manager.Configuration = cfg
		packagePath, err := manager.taskPackagePath(result.ID)
		require.NoError(t, err)
		packageBytes, err := os.ReadFile(packagePath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(packagePath, []byte("changed ciphertext"), 0600))
		_, status, body = postChatRecoveryAction(result, "selected-customer-service-restore-choice", map[string]string{"preview_sha256": result.SelectedModuleRestorePreview.SHA256, "policy": "use_package_conversation_rows"})
		require.NotEqual(t, http.StatusOK, status)
		require.Contains(t, body, "加密包已变化")
		require.NoError(t, os.WriteFile(packagePath, packageBytes, 0600))
	})
	t.Run("changed_shared_customer_and_private_file_report_distinct_conflicts", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `UPDATE users SET username='newer live customer' WHERE email='module-customer@example.invalid'`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, err := db.ExecContext(ctx, `UPDATE users SET username='module-customer' WHERE email='module-customer@example.invalid'`)
			require.NoError(t, err)
		})
		privateFile := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service", "attachments", "originals", "module-chat.bin")
		original, err := os.ReadFile(privateFile)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(privateFile, []byte("newer live file"), 0600))
		t.Cleanup(func() { require.NoError(t, os.WriteFile(privateFile, original, 0600)) })
		result := verifyChatPreview()
		require.Equal(t, "module_restore_conflicts_found", result.State, result.Error)
		require.Positive(t, result.SelectedModuleRestorePreview.DependencyConflicts)
		require.Equal(t, 1, result.SelectedModuleRestorePreview.FileConflicts)
		require.Zero(t, result.SelectedModuleRestorePreview.RestoredRows)
		require.True(t, result.SelectedModuleRestorePreview.DisposableStageRemoved)
	})
	t.Run("missing_processed_event_requires_execution_fact_reconciliation", func(t *testing.T) {
		var original string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT to_jsonb(e)::text FROM outbox_events e WHERE event_key='module-chat-realtime'`).Scan(&original))
		_, err := db.ExecContext(ctx, `DELETE FROM outbox_events WHERE event_key='module-chat-realtime'`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, err := db.ExecContext(ctx, `INSERT INTO outbox_events SELECT * FROM jsonb_populate_record(NULL::outbox_events,$1::jsonb)`, original)
			require.NoError(t, err)
		})
		result := verifyChatPreview()
		require.Equal(t, "module_restore_conflicts_found", result.State, result.Error)
		require.Zero(t, result.SelectedModuleRestorePreview.RestoredRows)
		found := false
		for _, table := range result.SelectedModuleRestorePreview.Tables {
			found = found || (table.Table == "outbox_events" && table.ProtectedNewRows == 1)
		}
		require.True(t, found, "a receipt cannot silently be re-created as an executable event")
	})
	missingFile := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service", "attachments", "originals", "module-chat.bin")
	require.NoError(t, os.Remove(missingFile))
	failed, err := manager.CreateSelectedBusinessModuleBackup("synthetic-module-operator", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/customer_service", []string{"customer_service"})
	require.NoError(t, err)
	failed = waitForIndependentRecoveryTask(t, manager, failed.ID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "module-chat.bin")
	packagePath, err := manager.taskPackagePath(failed.ID)
	require.NoError(t, err)
	_, err = os.Stat(packagePath)
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(cfg.MaintenanceDirectory, "application-maintenance.json"))
	require.True(t, os.IsNotExist(err), "failed export must release its own maintenance isolation")
	_, after, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	require.Equal(t, before, after, "exports, preview and upload verification cannot mutate rows, event statuses or sequences")
	t.Run("restart_cleans_only_the_persisted_disposable_database_identity", func(t *testing.T) {
		task, err := manager.createTask("upload", "synthetic-interrupted-module-rehearsal")
		require.NoError(t, err)
		stageName := "local_backup_stage_" + newTaskID()[:16]
		_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(stageName)+` OWNER local_backup_isolated_owner`)
		require.NoError(t, err)
		defer func() {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(stageName)+` WITH (FORCE)`)
		}()
		var oid int64
		require.NoError(t, readDatabaseOID(ctx, control, stageName, &oid))
		task.State, task.Phase = "running", "checking_selected_module_target_dependencies"
		task.SelectedModuleDisposableStage = &SelectedBusinessModuleDisposableStageIdentity{DatabaseName: stageName, DatabaseOID: oid + 1}
		require.NoError(t, manager.saveTask(task))
		require.ErrorContains(t, manager.cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(ctx, task), "OID")
		var retainedOID int64
		require.NoError(t, readDatabaseOID(ctx, control, stageName, &retainedOID))
		require.Equal(t, oid, retainedOID)
		task.SelectedModuleDisposableStage.DatabaseOID = 0
		require.ErrorContains(t, manager.cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(ctx, task), "OID")
		task.SelectedModuleDisposableStage.DatabaseOID = oid
		require.NoError(t, manager.saveTask(task))
		directory, err := manager.taskDirectory(task.ID)
		require.NoError(t, err)
		for _, name := range []string{"selected-module-disposable-restore-interrupted", "selected-module-authenticated-package-interrupted"} {
			require.NoError(t, os.Mkdir(filepath.Join(directory, name), 0700))
		}
		require.NoError(t, manager.Close())
		restarted, err := NewManager(cfg)
		require.NoError(t, err)
		defer restarted.Close()
		actual, err := restarted.GetTask(task.ID)
		require.NoError(t, err)
		require.Equal(t, "interrupted", actual.State)
		require.Nil(t, actual.SelectedModuleDisposableStage)
		require.NoError(t, readDatabaseOID(ctx, control, stageName, &retainedOID))
		require.Zero(t, retainedOID)
		for _, name := range []string{"selected-module-disposable-restore-interrupted", "selected-module-authenticated-package-interrupted"} {
			_, err := os.Stat(filepath.Join(directory, name))
			require.True(t, os.IsNotExist(err))
		}
	})
}
