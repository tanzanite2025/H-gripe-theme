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

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/repository"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func downloadEncryptedCustomerProfileModulePackageThroughScopedHTTPAndUploadForFormalRestoreTest(t *testing.T, manager *Manager) *Task {
	t.Helper()
	cfg := manager.Configuration
	backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-customer-exporter", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/customers", []string{"customers"})
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-customer-exporter", []string{PermissionExport})
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
	require.Equal(t, []string{"customers"}, upload.Manifest.SelectedModules)
	return upload
}

func TestPostgreSQLCustomerProfileModuleFormalRestoreOriginalLoginClosedRewardsAndAtomicInterruptionRecovery(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL and complete application migrations")
	}
	for _, scenario := range []string{"authenticated_restore", "restore_intent_persisted", "rows_and_sequences_before_commit", "rows_committed_before_acknowledgement", "verified_before_fence_removal", "fence_removed_before_acknowledgement", "stale_database_after_preview", "existing_account_missing_balance_blocked", "missing_account_current_loyalty_blocked", "missing_account_current_order_blocked", "case_insensitive_email_identity_conflict", "visitor_identity_collision", "subscription_email_collision", "source_missing_balance_summary", "unfinished_legacy_referral_blocked"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_customers_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_customers_" + newTaskID()[:12]
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
			passwordHash, err := bcrypt.GenerateFromPassword([]byte("synthetic-original-customer-password"), bcrypt.MinCost)
			require.NoError(t, err)
			_, err = database.ExecContext(ctx, `INSERT INTO users(id,email,username,password,role,first_name) VALUES(900001,'preserved-customer@example.invalid','preserved-customer','synthetic-old-password','user','Original profile'),(900002,'preserved-admin@example.invalid','preserved-admin','synthetic-admin-password','admin','Preserved staff'),(900010,'missing-customer@example.invalid','missing-customer',$1,'user','Original missing customer'),(900011,'missing-referee@example.invalid','missing-referee',$1,'user','Original missing referee')`, string(passwordHash))
			require.NoError(t, err)
			_, err = database.ExecContext(ctx, `
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900001,'Preserved product','customer-restore-preserved-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'customer-restore-preserved-sku',123,'USD',19);
INSERT INTO orders(id,order_number,user_id,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'preserved-customer-order',900001,'pending','USD','USD',123,123,123,0,0,0,0);
INSERT INTO tickets(id,ticket_number,conversation_id,user_id,customer_user_id,assigned_to,subject,category) VALUES(900001,'preserved-customer-chat','preserved-customer-chat',900001,900001,900002,'Preserved customer conversation','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Current chat must survive customer restore','[]','media_asset_bindings');
INSERT INTO user_loyalty(id,user_id,total_points,available_points) VALUES(900001,900001,5,5),(900010,900010,6,6),(900011,900011,5,5);
INSERT INTO loyalty_transactions(id,user_id,type,source,source_id,points,balance,description) VALUES(900010,900010,'earn','referral',900010,5,5,'Original registration reward'),(900011,900011,'earn','referral',900010,5,5,'Original referred registration reward'),(900012,900010,'earn','checkin',900010,1,6,'Original checkin');
INSERT INTO check_ins(id,user_id,check_in_date,points_earned) VALUES(900010,900010,'2026-10-01',1);
INSERT INTO referrals(id,referrer_id,referred_id,referral_code,status,points_earned,completed_at) VALUES(900010,900010,900011,'SYNTHETICCUSTOMER','completed',5,now());
INSERT INTO user_referral_identities(id,user_id,referral_code) VALUES(900010,900010,'ABCD23');
INSERT INTO visitor_profiles(id,user_id,customer_service_visitor_hash,email) VALUES(900010,900010,'synthetic-customer-visitor','missing-customer@example.invalid');
INSERT INTO wishlist_items(id,user_id,product_id) VALUES(900010,900010,900001);
INSERT INTO subscriptions(id,email,status) VALUES(900001,'preserved-customer@example.invalid','subscribed');
`)
			require.NoError(t, err)
			if scenario == "source_missing_balance_summary" {
				_, err = database.ExecContext(ctx, `DELETE FROM user_loyalty WHERE user_id=900010`)
				require.NoError(t, err)
			}
			if scenario == "unfinished_legacy_referral_blocked" {
				_, err = database.ExecContext(ctx, `UPDATE referrals SET status='pending' WHERE id=900010`)
				require.NoError(t, err)
			}
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			task := downloadEncryptedCustomerProfileModulePackageThroughScopedHTTPAndUploadForFormalRestoreTest(t, manager)
			_, err = database.ExecContext(ctx, `
UPDATE users SET password='synthetic-current-password',status='suspended',first_name='Current preserved profile' WHERE id=900001;
UPDATE user_loyalty SET total_points=99,available_points=99 WHERE user_id=900001;
UPDATE subscriptions SET status='unsubscribed',unsubscribed_at=now() WHERE id=900001;
DELETE FROM user_referral_identities WHERE user_id IN (900010,900011);
DELETE FROM loyalty_transactions WHERE user_id IN (900010,900011);
DELETE FROM user_loyalty WHERE user_id IN (900010,900011);
DELETE FROM check_ins WHERE user_id IN (900010,900011);
DELETE FROM referrals WHERE id=900010;
DELETE FROM visitor_profiles WHERE id=900010;
DELETE FROM wishlist_items WHERE id=900010;
DELETE FROM users WHERE id IN (900010,900011);
INSERT INTO users(id,email,username,password,role) VALUES(900020,'current-additional@example.invalid','current-additional','synthetic-current-additional-password','user');
`)
			require.NoError(t, err)
			invalidQueries := map[string]string{
				"existing_account_missing_balance_blocked": `DELETE FROM user_loyalty WHERE user_id=900001`,
				"missing_account_current_loyalty_blocked":  `INSERT INTO user_loyalty(user_id,total_points,available_points) VALUES(900010,100,100)`,
				"missing_account_current_order_blocked":    `UPDATE orders SET user_id=900010 WHERE id=900001`,
				"case_insensitive_email_identity_conflict": `UPDATE users SET email='MISSING-CUSTOMER@example.invalid' WHERE id=900020`,
				"visitor_identity_collision":               `INSERT INTO visitor_profiles(id,customer_service_visitor_hash) VALUES(900020,'synthetic-customer-visitor')`,
				"subscription_email_collision":             `DELETE FROM subscriptions WHERE id=900001;INSERT INTO subscriptions(id,email,status) VALUES(900020,'PRESERVED-CUSTOMER@example.invalid','unsubscribed')`,
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
			if invalidQueries[scenario] != "" || scenario == "unfinished_legacy_referral_blocked" || scenario == "source_missing_balance_summary" {
				require.Equal(t, "module_restore_conflicts_found", task.State, task.Error)
				require.False(t, task.SelectedModuleRestorePreview.CustomerProfileRestoreEligibility.Eligible)
				require.Error(t, validateCustomerProfileModuleFormalRestorePreview(task, task.SelectedModuleRestorePreview.SHA256))
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, originalFingerprint, actual)
				require.NoError(t, database.Close())
				return
			}
			previewSummary, err := json.Marshal(task.SelectedModuleRestorePreview)
			require.NoError(t, err)
			require.Equal(t, "module_isolated_restore_verified", task.State, "%s %s", task.Error, string(previewSummary))
			require.True(t, task.SelectedModuleRestorePreview.CustomerProfileRestoreEligibility.Eligible)
			require.Equal(t, 12, task.SelectedModuleRestorePreview.RestoredRows)
			request := CustomerProfileModuleFormalRestoreRequest{Confirmation: "RESTORE CUSTOMER PROFILES " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256}
			invokeCustomerProfileRestoreHTTP := func(action, token string, body any, expectedStatus int) {
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
				invokeCustomerProfileRestoreHTTP("customer-profile-module-formal-restore", token, request, http.StatusForbidden)
				invokeCustomerProfileRestoreHTTP("customer-profile-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "invalid"}, http.StatusForbidden)
			}
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-customer-restorer", []string{PermissionRestore})
			require.NoError(t, err)
			staleRequest := request
			staleRequest.ExpectedPreviewSHA256 = strings.Repeat("0", 64)
			invokeCustomerProfileRestoreHTTP("customer-profile-module-formal-restore", token, staleRequest, http.StatusConflict)
			if scenario == "stale_database_after_preview" {
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `UPDATE users SET first_name='Changed after preview' WHERE id=900020`)
				require.NoError(t, err)
				require.NoError(t, database.Close())
				invokeCustomerProfileRestoreHTTP("customer-profile-module-formal-restore", token, request, http.StatusOK)
				task = waitForIndependentRecoveryTask(t, manager, task.ID)
				require.Equal(t, "failed", task.State, task.Error)
				require.Nil(t, task.CustomerProfileModuleRestore)
				return
			}
			authenticated := scenario == "authenticated_restore"
			if authenticated {
				invokeCustomerProfileRestoreHTTP("customer-profile-module-formal-restore", token, request, http.StatusOK)
			} else {
				require.NoError(t, manager.startTaskOperation(task, "synthetic_customer_profile_formal_restore", func(ctx context.Context, task *Task) error {
					return manager.restoreCustomerProfilesWithAtomicReceiptAndRecoverableFiles(ctx, task, "synthetic-customer-restorer", func(point string) error {
						if point == scenario {
							return errors.New("synthetic interrupted customer profile restore")
						}
						return nil
					})
				}))
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			committed := authenticated || scenario == "rows_committed_before_acknowledgement" || scenario == "verified_before_fence_removal" || scenario == "fence_removed_before_acknowledgement"
			if !authenticated {
				require.Equal(t, "failed", task.State, task.Error)
				require.Equal(t, "synthetic interrupted customer profile restore", task.Error)
				require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
				_, err = manager.VerifyUploadedPackage(task.ID)
				require.Error(t, err)
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, manager.ExpireWorkCopies(ctx))
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				require.False(t, task.WorkCopyExpired)
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				if scenario != "fence_removed_before_acknowledgement" {
					require.Equal(t, "interrupted", task.State, task.Error)
					invokeCustomerProfileRestoreHTTP("customer-profile-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "RECOVER CUSTOMER PROFILES " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, http.StatusOK)
					task = waitForIndependentRecoveryTask(t, manager, task.ID)
				}
			}
			require.Equal(t, map[bool]string{true: "completed", false: "rolled_back"}[committed], task.State, task.Error)
			require.NoError(t, validateTransactionalBusinessModuleRestoreProgress(task))
			require.Equal(t, "released", task.CustomerProfileModuleRestore.Phase)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer database.Close()
			if !committed {
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, originalFingerprint, actual)
				return
			}
			require.Equal(t, 12, task.CustomerProfileModuleRestore.RestoredRows)
			orm, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			users := repository.NewUserRepository(orm)
			restored, err := users.FindByEmail("missing-customer@example.invalid")
			require.NoError(t, err)
			require.Equal(t, uint(900010), restored.ID)
			require.True(t, restored.CheckPassword("synthetic-original-customer-password"))
			current, err := users.FindByID(900001)
			require.NoError(t, err)
			require.Equal(t, "synthetic-current-password", current.Password)
			require.Equal(t, "suspended", current.Status)
			require.Equal(t, "Current preserved profile", current.FirstName)
			var currentPoints, restoredPoints, rewards, audits, stock, additional int
			var subscription, chat string
			require.NoError(t, database.QueryRowContext(ctx, `SELECT available_points FROM user_loyalty WHERE user_id=900001`).Scan(&currentPoints))
			require.Equal(t, 99, currentPoints)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT available_points FROM user_loyalty WHERE user_id=900010`).Scan(&restoredPoints))
			require.Equal(t, 6, restoredPoints)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM loyalty_transactions WHERE source='referral'`).Scan(&rewards))
			require.Equal(t, 2, rewards)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM audit_logs WHERE resource='customer_profile_module_formal_restore'`).Scan(&audits))
			require.Equal(t, 1, audits)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900001`).Scan(&stock))
			require.Equal(t, 19, stock)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE id=900020`).Scan(&additional))
			require.Equal(t, 1, additional)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT status FROM subscriptions WHERE id=900001`).Scan(&subscription))
			require.Equal(t, "unsubscribed", subscription)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&chat))
			require.Equal(t, "Current chat must survive customer restore", chat)
			public, err := json.Marshal(task.CustomerProfileModuleRestore)
			require.NoError(t, err)
			for _, privateText := range []string{"missing-customer@example.invalid", string(passwordHash), "Original missing customer"} {
				require.NotContains(t, string(public), privateText)
			}
		})
	}
}
