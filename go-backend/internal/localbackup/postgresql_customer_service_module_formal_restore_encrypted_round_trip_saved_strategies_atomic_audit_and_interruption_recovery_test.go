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

func TestPostgreSQLCustomerServiceModuleFormalRestoreSavedStrategiesAtomicAuditAndInterruptedAttachments(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL and complete application migrations")
	}
	for _, scenario := range []string{"use_package_authenticated_restore", "preserve_current_authenticated_restore", "restore_intent_persisted", "attachment_linked_before_acknowledgement", "rows_and_sequences_before_commit", "rows_committed_before_acknowledgement", "verified_before_fence_removal", "fence_removed_before_acknowledgement", "stale_database_after_rehearsal", "changed_file_after_rehearsal", "changed_package_after_rehearsal", "invalid_shared_dependency", "invalid_upload_actor", "invalid_current_event"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_chat_restore_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_chat_restore_" + newTaskID()[:12]
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
			_, err = database.ExecContext(ctx, `
INSERT INTO users(id,email,username,password,role) VALUES(900001,'chat-customer@example.invalid','chat-customer','synthetic-unused-password','user'),(900002,'chat-staff@example.invalid','chat-staff','synthetic-unused-password','admin');
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900001,'Preserved product','chat-restore-preserved-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'chat-restore-preserved-sku',123,'USD',19);
INSERT INTO orders(id,order_number,user_id,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'preserved-customer-order',900001,'pending','USD','USD',123,123,123,0,0,0,0);
INSERT INTO loyalty_transactions(user_id,type,source,source_id,points,balance,description) VALUES(900001,'earn','referral',900001,5,5,'Preserved registration reward');
INSERT INTO tickets(id,ticket_number,conversation_id,user_id,customer_user_id,assigned_to,subject,category) VALUES(900001,'synthetic-chat-restore','synthetic-chat-restore',900001,900001,900002,'Original conversation','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Original historical searchable customer text','[]','media_asset_bindings'),(900003,900001,900001,'Original historical attachment message','[]','media_asset_bindings');
INSERT INTO customer_service_inbox_states(ticket_id,recipient_user_id,last_read_message_id,assignment_version,unread_count) VALUES(900001,900002,0,1,2);
INSERT INTO media_assets(id,filename,url,storage_key,size,visibility,library_partition,customer_service_media_purpose,uploader_id,uploader_actor_type,classification_method) VALUES(900001,'chat-proof.txt','/api/customer-service/attachments/900001','customer-service/attachments/originals/chat-proof.txt',6,'private','customer_service','conversation_attachment',900001,'customer','recorded_upload_context');
INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,object_key,media_asset_id,status,expires_at) VALUES('synthetic-chat-restore-upload','customer_service',900001,'customer',900001,'customer-service/attachments/originals/chat-proof.txt',900001,'bound',now()+interval '1 day');
INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(900003,900001,0,'synthetic-chat-restore-upload');
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES(900001,'synthetic-chat-restore-realtime','customer_service.realtime','customer_service_conversation','900001','{"ticket_id":900001,"conversation_id":"synthetic-chat-restore","type":"conversation.message.created","event_id":"synthetic-chat-restore-realtime","payload":{"message_id":900003}}','processed',now());
`)
			require.NoError(t, err)
			attachment := filepath.Join(cfg.StorageRoots["uploads_private"], "customer-service/attachments/originals/chat-proof.txt")
			require.NoError(t, os.MkdirAll(filepath.Dir(attachment), 0700))
			require.NoError(t, os.WriteFile(attachment, []byte("backup"), 0600))
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			task := downloadEncryptedSelectedChatPackageUsingScopedHTTPAndUploadForAuditedCreationTest(t, manager)
			_, err = database.ExecContext(ctx, `UPDATE ticket_messages SET content='Newer current customer text' WHERE id=900001;
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900002,900001,900001,'Current additional message must survive','[]','media_asset_bindings');
DELETE FROM ticket_message_media_attachments WHERE ticket_message_id=900003;
DELETE FROM ticket_messages WHERE id=900003`)
			require.NoError(t, err)
			require.NoError(t, os.Remove(attachment))
			invalidQueries := map[string]string{
				"invalid_shared_dependency": `UPDATE users SET username='Changed shared customer' WHERE id=900001`,
				"invalid_upload_actor":      `UPDATE media_upload_sessions SET upload_actor_account_id=900002 WHERE id='synthetic-chat-restore-upload'`,
				"invalid_current_event":     `UPDATE outbox_events SET status='pending',processed_at=NULL WHERE id=900001`,
			}
			if query := invalidQueries[scenario]; query != "" {
				_, err = database.ExecContext(ctx, query)
				require.NoError(t, err)
			}
			_, originalFingerprint, _, err := captureDatabaseInventory(ctx, database)
			require.NoError(t, err)
			require.NoError(t, database.Close())
			_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID)
			require.NoError(t, err)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			if strings.HasPrefix(scenario, "invalid_") {
				require.Equal(t, "module_restore_conflicts_found", task.State, task.Error)
				_, err = manager.SaveSelectedCustomerServiceModuleRestoreChoice(ctx, task.ID, "synthetic-chat-restorer", task.SelectedModuleRestorePreview.SHA256, "use_package_conversation_rows")
				require.Error(t, err)
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, originalFingerprint, actual)
				require.NoError(t, database.Close())
				return
			}
			policy := "use_package_conversation_rows"
			if scenario == "preserve_current_authenticated_restore" {
				policy = "preserve_current_conflicting_conversation_rows"
			}
			task, err = manager.SaveSelectedCustomerServiceModuleRestoreChoice(ctx, task.ID, "synthetic-chat-restorer", task.SelectedModuleRestorePreview.SHA256, policy)
			require.NoError(t, err)
			choice := task.SelectedCustomerServiceRestoreChoice
			_, err = manager.StartSelectedCustomerServiceModuleCandidateRehearsal(task.ID, choice.SHA256, "VERIFY CHAT CHOICE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID+"/"+choice.SHA256)
			require.NoError(t, err)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			require.Equal(t, "module_chat_candidate_rehearsal_verified", task.State, task.Error)
			rehearsal := task.SelectedCustomerServiceCandidateRehearsal
			require.Equal(t, customerServiceModuleFormalRestorePolicy, rehearsal.CustomerServiceFormalRestorePolicy)
			if scenario == "use_package_authenticated_restore" {
				t.Run("old_rehearsal_without_formal_contract_requires_new_rehearsal", func(t *testing.T) {
					oldTask, oldRehearsal := *task, *rehearsal
					oldRehearsal.CustomerServiceFormalRestorePolicy, oldRehearsal.SHA256 = "", ""
					oldRehearsal.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(oldRehearsal)
					require.NoError(t, err)
					oldTask.SelectedCustomerServiceCandidateRehearsal = &oldRehearsal
					require.Error(t, validateCustomerServiceModuleFormalRestoreAuthorization(&oldTask, choice.SHA256, oldRehearsal.SHA256))
				})
			}
			request := CustomerServiceModuleFormalRestoreRequest{Confirmation: "RESTORE CHAT MODULE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID + "/" + choice.SHA256, ExpectedChoiceSHA256: choice.SHA256, ExpectedRehearsalSHA256: rehearsal.SHA256}
			invokeChatRestoreHTTP := func(action, token string, body any, expectedStatus int) {
				t.Helper()
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				httpRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+task.ID+"/"+action, strings.NewReader(string(encoded)))
				httpRequest.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, httpRequest)
				require.Equal(t, expectedStatus, response.Code, response.Body.String())
			}
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			for _, permission := range []string{PermissionView, PermissionUpload, PermissionExport} {
				token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-unprivileged", []string{permission})
				require.NoError(t, err)
				invokeChatRestoreHTTP("customer-service-module-formal-restore", token, request, http.StatusForbidden)
				invokeChatRestoreHTTP("customer-service-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "invalid"}, http.StatusForbidden)
			}
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-chat-restorer", []string{PermissionRestore})
			require.NoError(t, err)
			stale := request
			stale.ExpectedRehearsalSHA256 = strings.Repeat("0", 64)
			invokeChatRestoreHTTP("customer-service-module-formal-restore", token, stale, http.StatusConflict)
			if strings.HasSuffix(scenario, "after_rehearsal") {
				switch scenario {
				case "stale_database_after_rehearsal":
					database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
					require.NoError(t, err)
					_, err = database.ExecContext(ctx, `UPDATE ticket_messages SET content='Changed after rehearsal' WHERE id=900002`)
					require.NoError(t, err)
					require.NoError(t, database.Close())
				case "changed_file_after_rehearsal":
					require.NoError(t, os.WriteFile(attachment, []byte("other!"), 0600))
				case "changed_package_after_rehearsal":
					packagePath, err := manager.taskPackagePath(task.ID)
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(packagePath, []byte("invalid ciphertext"), 0600))
				}
				invokeChatRestoreHTTP("customer-service-module-formal-restore", token, request, http.StatusOK)
				task = waitForIndependentRecoveryTask(t, manager, task.ID)
				require.Equal(t, "failed", task.State, task.Error)
				require.Nil(t, task.CustomerServiceModuleRestore)
				fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
				require.NoError(t, err)
				require.False(t, fenced)
				return
			}
			authenticated := strings.HasSuffix(scenario, "authenticated_restore")
			if authenticated {
				invokeChatRestoreHTTP("customer-service-module-formal-restore", token, request, http.StatusOK)
			} else {
				require.NoError(t, manager.startTaskOperation(task, "synthetic_chat_formal_restore", func(ctx context.Context, task *Task) error {
					return manager.restoreCustomerServiceModuleWithSavedStrategyAtomicReceiptAndRecoverableFiles(ctx, task, "synthetic-chat-restorer", func(point string) error {
						if point == scenario {
							return errors.New("synthetic interrupted chat restore")
						}
						return nil
					})
				}))
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			committed := authenticated || scenario == "rows_committed_before_acknowledgement" || scenario == "verified_before_fence_removal" || scenario == "fence_removed_before_acknowledgement"
			if !authenticated {
				require.Equal(t, "failed", task.State, task.Error)
				require.Equal(t, "synthetic interrupted chat restore", task.Error)
				require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
				_, err = manager.VerifyUploadedPackage(task.ID)
				require.Error(t, err)
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, manager.ExpireWorkCopies(ctx))
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				require.False(t, task.WorkCopyExpired)
				if !committed {
					database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
					require.NoError(t, err)
					_, actual, _, err := captureDatabaseInventory(ctx, database)
					require.NoError(t, err)
					require.Equal(t, originalFingerprint, actual, "chat rows, audit and sequences must roll back together")
					require.NoError(t, database.Close())
				}
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				if scenario != "fence_removed_before_acknowledgement" {
					require.Equal(t, "interrupted", task.State, task.Error)
					invokeChatRestoreHTTP("customer-service-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "RECOVER CHAT RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, http.StatusOK)
					task = waitForIndependentRecoveryTask(t, manager, task.ID)
				}
			}
			require.Equal(t, map[bool]string{true: "completed", false: "rolled_back"}[committed], task.State, task.Error)
			require.NoError(t, validateTransactionalBusinessModuleRestoreProgress(task))
			require.Equal(t, "released", task.CustomerServiceModuleRestore.Phase)
			if committed {
				require.Equal(t, 2, task.CustomerServiceModuleRestore.RestoredRows)
				require.Equal(t, map[bool]int{true: 0, false: 1}[policy == "preserve_current_conflicting_conversation_rows"], task.CustomerServiceModuleRestore.ReplacedRows)
			}
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			staged, err := filepath.Glob(filepath.Join(cfg.StorageRoots["uploads_private"], ".local-backup-customer-service-restore-*"))
			require.NoError(t, err)
			require.Empty(t, staged)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer database.Close()
			var originalMessage, additionalMessage string
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&originalMessage))
			if committed && policy == "use_package_conversation_rows" {
				require.Equal(t, "Original historical searchable customer text", originalMessage)
			} else {
				require.Equal(t, "Newer current customer text", originalMessage)
			}
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900002`).Scan(&additionalMessage))
			require.Equal(t, "Current additional message must survive", additionalMessage)
			var messages, audits, stocks, points, pendingEvents int
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM ticket_messages WHERE id=900003`).Scan(&messages))
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM audit_logs WHERE resource='customer_service_module_formal_restore'`).Scan(&audits))
			require.Equal(t, map[bool]int{true: 1, false: 0}[committed], messages)
			require.Equal(t, messages, audits)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900001`).Scan(&stocks))
			require.Equal(t, 19, stocks)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT points FROM loyalty_transactions WHERE source='referral'`).Scan(&points))
			require.Equal(t, 5, points)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE status<>'processed'`).Scan(&pendingEvents))
			require.Zero(t, pendingEvents)
			if committed {
				actualFile, err := os.ReadFile(attachment)
				require.NoError(t, err)
				require.Equal(t, "backup", string(actualFile))
				orm, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
				require.NoError(t, err)
				ticketRepository := repository.NewTicketRepository(orm)
				messages, err := ticketRepository.FindMessagesByTicketID(900001, 20, 0)
				require.NoError(t, err)
				require.Len(t, messages, 3)
				conversations, count, err := ticketRepository.FindCustomerServiceConversations(1, 20, repository.CustomerServiceConversationFilters{Search: "Original historical attachment message"})
				require.NoError(t, err)
				require.Equal(t, int64(1), count)
				require.Len(t, conversations, 1)
				require.Equal(t, uint(900001), conversations[0].ID)
				dispatcher := service.NewOutboxService(repository.NewOutboxRepository(orm))
				dispatcher.RegisterHandler(outbox.EventTypeCustomerServiceRealtime, func(context.Context, outbox.Event) error {
					t.Error("restored history must not replay realtime delivery")
					return nil
				})
				dispatched, err := dispatcher.ProcessPending(ctx, time.Now().UTC(), 10)
				require.NoError(t, err)
				require.Zero(t, dispatched.Claimed)
			} else {
				_, err := os.Stat(attachment)
				require.ErrorIs(t, err, os.ErrNotExist)
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, originalFingerprint, actual)
			}
			public, err := json.Marshal(task.CustomerServiceModuleRestore)
			require.NoError(t, err)
			for _, privateText := range []string{"Original historical", "Newer current", "chat-customer@example.invalid", "synthetic-chat-restore-upload"} {
				require.NotContains(t, string(public), privateText)
			}
		})
	}
}
