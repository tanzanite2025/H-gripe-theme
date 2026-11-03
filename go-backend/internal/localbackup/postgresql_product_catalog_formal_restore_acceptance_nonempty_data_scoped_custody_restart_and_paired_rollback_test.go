package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"image/color"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/service"
)

func TestPostgreSQLProductCatalogFormalRestoreAcceptancePreservesOtherModulesAndHandlesCustodyAndReleaseInterruptions(t *testing.T) {
	connectionURL, migrationsDirectory, redisURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY"), os.Getenv("LOCAL_BACKUP_INTEGRATION_REDIS_URL")
	if connectionURL == "" || migrationsDirectory == "" || redisURL == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_REDIS_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL/Redis with complete migrations and pg_dump/pg_restore")
	}
	for _, scenario := range []string{"authenticated_acceptance_and_fresh_application_access", "acceptance_intent_persisted", "database_custody_committed", "accepted_database_open_before_fence_removal", "accepted_fence_removed_before_acknowledgement"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_product_accept_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_product_accept_" + newTaskID()[:12]
			cfg.MaxFiles, cfg.MaxExpandedBytes, cfg.MaxWorkspaceBytes = 10000, 256<<20, 1<<30
			options, err := redis.ParseURL(redisURL)
			require.NoError(t, err)
			cfg.ProductViewRecoveryRedis = &ProductViewRecoveryRedisConnectionConfiguration{Mode: "standalone", Addresses: []string{options.Addr}, Database: options.DB, Username: options.Username, Password: options.Password}
			cfg.ProductViewRecoveryWriterInventory = []string{"synthetic single-host application and all background writers"}
			projection := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
			productProjection := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true)
			projection.ProductCatalog = &productProjection
			projection.RedisDestinationSHA256 = maintenancegate.FingerprintRedisRecoveryDestination("standalone", []string{options.Addr}, options.DB, "")
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, projection))
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` LOGIN PASSWORD 'synthetic-product-acceptance-login'`)
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			var manager *Manager
			var task *Task
			unrelatedDatabase := "local_backup_unrelated_" + newTaskID()[:12]
			t.Cleanup(func() {
				defer control.Close()
				if manager != nil {
					_ = manager.Close()
				}
				names := []string{cfg.DatabaseName, unrelatedDatabase}
				if task != nil {
					names = append(names, task.StageDatabase)
					if task.Journal != nil {
						names = append(names, task.Journal.RollbackDatabase)
					}
				}
				for _, name := range names {
					if name != "" {
						_, cleanupErr := control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
						require.NoError(t, cleanupErr)
					}
				}
				_, cleanupErr := control.ExecContext(context.Background(), `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
				require.NoError(t, cleanupErr)
				client := redis.NewClient(options)
				defer client.Close()
				require.NoError(t, client.Del(context.Background(), maintenancegate.ProductViewRedisOwnershipKey).Err())
			})
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
			image := createDecodableProductCatalogPNGImageFixture(t, 2, color.RGBA{R: 190, B: 100, A: 255})
			_, err = database.ExecContext(ctx, `INSERT INTO users(id,email,username,password) VALUES(900001,'product-accept@example.invalid','synthetic-product-customer','synthetic-unused-password');
INSERT INTO products(id,name,slug,description,price_minor,currency) VALUES(900001,'Backed up product','product-acceptance','Backup description',12345,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'product-acceptance-sku',12345,'USD',19);
INSERT INTO media_assets(id,filename,url,storage_key,size,visibility) VALUES(900001,'acceptance.png','/uploads/products/acceptance.png','products/acceptance.png',1,'public');
INSERT INTO product_media(product_id,variant_id,media_asset_id,media_type,url) VALUES(900001,900001,900001,'image','/uploads/products/acceptance.png');
INSERT INTO tickets(id,ticket_number,user_id,customer_user_id,conversation_id,subject,category) VALUES(900001,'product-acceptance-chat',900001,900001,'product-acceptance-chat','Retained chat','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Original retained chat','[]','media_asset_bindings');
INSERT INTO orders(id,order_number,user_id,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'product-acceptance-order',900001,'USD','USD',12345,12345,12345,0,0,0,0);
INSERT INTO order_items(order_id,product_id,variant_id,product_name,quantity,currency,price_minor,subtotal_minor,total_minor,tax_amount_minor,discount_minor) VALUES(900001,900001,900001,'Retained historical order description',1,'USD',12345,12345,12345,0,0);
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('product-acceptance-processed','order.completed','order','900001','{}','processed',now());`)
			require.NoError(t, err)
			_, err = database.ExecContext(ctx, `UPDATE media_assets SET size=$1 WHERE id=900001`, len(image))
			require.NoError(t, err)
			imagePath := filepath.Join(cfg.StorageRoots["uploads"], "products/acceptance.png")
			require.NoError(t, os.MkdirAll(filepath.Dir(imagePath), 0700))
			require.NoError(t, os.WriteFile(imagePath, image, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-chat.txt"), []byte("retained private chat attachment"), 0600))
			require.NoError(t, database.Close())
			client := redis.NewClient(options)
			require.NoError(t, maintenancegate.RegisterProductViewRedisRuntimeOwnershipBeforeStartingWriters(ctx, cfg.MaintenanceDirectory, cfg.SiteID, cfg.Environment, cfg.DatabaseName, projection.RedisDestinationSHA256, productProjection.ViewCountKeysSHA256, client))
			require.NoError(t, client.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			backup, err := manager.CreateProductCatalogBackup("synthetic-exporter", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
			require.NoError(t, err)
			backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
			require.Equal(t, "ready_to_download", backup.State, backup.Error)
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			exportToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-exporter", []string{PermissionExport})
			require.NoError(t, err)
			download := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/downloads/"+backup.ID, nil)
			download.Header.Set("Authorization", "Bearer "+exportToken)
			downloadResponse := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(downloadResponse, download)
			require.Equal(t, http.StatusOK, downloadResponse.Code)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			_, err = database.ExecContext(ctx, `UPDATE products SET name='Current changed product' WHERE id=900001; UPDATE product_variants SET stock=7 WHERE id=900001; UPDATE ticket_messages SET content='Latest retained chat' WHERE id=900001`)
			require.NoError(t, err)
			require.NoError(t, database.Close())
			task = uploadProductCatalogPackageForIntegrationTest(t, manager, downloadResponse.Body.Bytes())
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(unrelatedDatabase)+` OWNER local_backup_isolated_owner`)
			require.NoError(t, err)
			restoreToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-accepting-operator", []string{PermissionRestore})
			require.NoError(t, err)
			invokeProductAcceptanceHTTP := func(action string, body any, token string, expectedStatus int) {
				t.Helper()
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+task.ID+"/"+action, strings.NewReader(string(encoded)))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, expectedStatus, response.Code, response.Body.String())
			}
			invokeProductAcceptanceHTTP("product-catalog-stopped-handoff-preparation", ProductCatalogStoppedHandoffPreparationRequest{Confirmation: "PREPARE STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, AllListedWritersStoppedAndUseThisSingleHostGate: true}, restoreToken, http.StatusOK)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			require.Equal(t, "product_module_stopped_handoff_prepared", task.State, task.Error)
			invokeProductAcceptanceHTTP("product-catalog-stopped-isolated-switch", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "SWITCH STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, restoreToken, http.StatusOK)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			require.Equal(t, "product_module_switched_still_isolated", task.State, task.Error)
			acceptRequest := ProductCatalogFormalRestoreAcceptanceRequest{Confirmation: "ACCEPT PRODUCT RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}
			invokeProductAcceptanceHTTP("product-catalog-formal-restore-acceptance", acceptRequest, exportToken, http.StatusForbidden)
			invokeProductAcceptanceHTTP("product-catalog-formal-restore-acceptance", acceptRequest, restoreToken, http.StatusConflict)
			invokeProductAcceptanceHTTP("product-catalog-stopped-detail-cache-invalidation", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "INVALIDATE STOPPED PRODUCT CACHE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, restoreToken, http.StatusOK)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			require.Equal(t, "product_module_switched_still_isolated", task.State, task.Error)
			report, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, task.ID)
			require.NoError(t, err)
			attestations := []ExternalProjectionReconciliationAttestation{}
			for _, requirement := range report.Requirements {
				attestations = append(attestations, ExternalProjectionReconciliationAttestation{Projection: requirement.Projection, Action: requirement.Action, EvidenceReference: "synthetic independently retained acceptance record", EvidenceSHA256: strings.Repeat("a", 64)})
			}
			evidence, err := manager.SaveStoppedProductProjectionIndependentOperatorEvidence(ctx, task.ID, "synthetic-independent-observer", ProductCatalogStoppedProjectionEvidenceRequest{Confirmation: "RECORD STOPPED PRODUCT PROJECTIONS " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ContractSHA256: report.ContractSHA256, Attestations: attestations})
			require.NoError(t, err)
			acceptRequest.ExpectedEvidenceSHA256 = strings.Repeat("b", 64)
			invokeProductAcceptanceHTTP("product-catalog-formal-restore-acceptance", acceptRequest, restoreToken, http.StatusConflict)
			acceptRequest.ExpectedEvidenceSHA256 = evidence.SHA256
			if scenario == "authenticated_acceptance_and_fresh_application_access" {
				invokeProductAcceptanceHTTP("product-catalog-formal-restore-acceptance", acceptRequest, restoreToken, http.StatusOK)
			} else {
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				require.NoError(t, manager.startTaskOperation(task, "synthetic_acceptance_interruption", func(ctx context.Context, task *Task) error {
					return manager.acceptStoppedProductCatalogRestoreAndResumeService(ctx, task, "synthetic-accepting-operator", func(point string) error {
						if point == scenario {
							return errors.New("synthetic acceptance interruption at " + point)
						}
						return nil
					})
				}))
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			if scenario == "authenticated_acceptance_and_fresh_application_access" {
				require.Equal(t, "completed", task.State, task.Error)
			} else {
				require.Equal(t, "failed", task.State, task.Error)
				require.Contains(t, task.Error, scenario)
			}
			require.NoError(t, manager.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			accepted := scenario == "authenticated_acceptance_and_fresh_application_access" || scenario == "accepted_fence_removed_before_acknowledgement"
			if accepted {
				require.Equal(t, "completed", task.State, task.Error)
				require.True(t, taskHasDurablyReleasedAcceptedProductCatalog(task))
				require.False(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task))
				_, err = manager.StartVerifiedPairedRollbackOfStoppedProductCatalogSwitch(task.ID, "synthetic", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "ROLLBACK STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID})
				require.Error(t, err, "a historical rollback must not erase resumed business writes")
			} else {
				require.Equal(t, "interrupted", task.State)
				invokeProductAcceptanceHTTP("product-catalog-stopped-switch-rollback", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "ROLLBACK STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, restoreToken, http.StatusOK)
				task = waitForIndependentRecoveryTask(t, manager, task.ID)
				require.Equal(t, "rolled_back", task.State, task.Error)
			}
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer database.Close()
			if accepted {
				_, err = database.ExecContext(ctx, `SET ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
				require.NoError(t, err)
			}
			var name, message, orderDescription, eventStatus string
			var stock int
			require.NoError(t, database.QueryRowContext(ctx, `SELECT name FROM products WHERE id=900001`).Scan(&name))
			if accepted {
				require.Equal(t, "Backed up product", name)
			} else {
				require.Equal(t, "Current changed product", name)
			}
			require.NoError(t, database.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900001`).Scan(&stock))
			require.Equal(t, 7, stock)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&message))
			require.Equal(t, "Latest retained chat", message)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT product_name FROM order_items WHERE order_id=900001`).Scan(&orderDescription))
			require.Equal(t, "Retained historical order description", orderDescription)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT status FROM outbox_events WHERE event_key='product-acceptance-processed'`).Scan(&eventStatus))
			require.Equal(t, "processed", eventStatus)
			if accepted {
				_, err = database.ExecContext(ctx, `UPDATE products SET description=description WHERE id=900001`)
				require.NoError(t, err, "the resumed application role must be able to write")
				applicationURL, err := url.Parse(connectionURL)
				require.NoError(t, err)
				applicationURL.User = url.UserPassword(cfg.ApplicationDatabaseRole, "synthetic-product-acceptance-login")
				applicationURL.Path = "/" + cfg.DatabaseName
				application, err := sql.Open("postgres", applicationURL.String())
				require.NoError(t, err)
				var actualRole, restoredName string
				require.NoError(t, application.QueryRowContext(ctx, `SELECT current_user,name FROM products WHERE id=900001`).Scan(&actualRole, &restoredName))
				require.Equal(t, cfg.ApplicationDatabaseRole, actualRole)
				require.Equal(t, "Backed up product", restoredName)
				require.NoError(t, application.Close())
			}
			actualImage, err := os.ReadFile(imagePath)
			require.NoError(t, err)
			require.Equal(t, image, actualImage)
			privateAttachment, err := os.ReadFile(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-chat.txt"))
			require.NoError(t, err)
			require.Equal(t, "retained private chat attachment", string(privateAttachment))
			var unrelatedOwner string
			require.NoError(t, control.QueryRowContext(ctx, `SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname=$1`, unrelatedDatabase).Scan(&unrelatedOwner))
			require.Equal(t, "local_backup_isolated_owner", unrelatedOwner, "acceptance cannot transfer another candidate database")
		})
	}
}

func TestProductCatalogFormalRestoreCompletedReceiptSurvivesLaterConfigurationAndRejectsCorruptTerminals(t *testing.T) {
	for _, scenario := range []string{"valid_after_later_configuration", "pending_acknowledgement_after_later_configuration", "changed_database_oid", "changed_evidence", "missing_verified_custody", "incomplete_media_move"} {
		t.Run(scenario, func(t *testing.T) {
			previous, task := createSyntheticStoppedProductSwitchWithVerifiedOriginalReleaseIntent(t)
			journal := task.Journal.ProductCatalogStoppedSwitch
			task.Journal.DatabaseSwitched = true
			task.ProductCatalogStoppedProjectionEvidence = &ProductCatalogStoppedProjectionOperatorEvidenceRecord{Version: 1, Scope: productCatalogBackupScope, TaskID: task.ID, WindowID: journal.WindowID, ContractSHA256: strings.Repeat("a", 64), RecordedAt: time.Now().UTC().Add(-time.Minute)}
			evidence := task.ProductCatalogStoppedProjectionEvidence
			var err error
			evidence.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(evidence)
			require.NoError(t, err)
			journal.Acceptance = &ProductCatalogFormalRestoreAcceptance{Version: 1, Operator: "synthetic", ApplicationDatabaseRole: previous.Configuration.ApplicationDatabaseRole, ProjectionEvidenceSHA256: evidence.SHA256, ProjectionContractSHA256: evidence.ContractSHA256, AcceptedAt: time.Now().UTC(), CustodyVerifiedAt: time.Now().UTC(), ServiceReleasedAt: time.Now().UTC()}
			journal.Phase = "product_released"
			task.State, task.Phase = "completed", "product_restore_accepted_and_service_resumed"
			for index := range journal.Moves {
				journal.Moves[index].Started, journal.Moves[index].Completed = true, true
			}
			require.NoError(t, previous.saveProductCatalogStoppedSwitchProgress(task))
			require.True(t, taskHasDurablyReleasedAcceptedProductCatalog(task))
			switch scenario {
			case "pending_acknowledgement_after_later_configuration":
				journal.Phase, task.State = "releasing_product", "running"
				journal.Acceptance.ServiceReleasedAt = time.Time{}
				require.NoError(t, previous.saveProductCatalogStoppedSwitchProgress(task))
				fenced, err := maintenancegate.IsApplicationFenced(previous.Configuration.MaintenanceDirectory)
				require.NoError(t, err)
				require.False(t, fenced)
			case "changed_database_oid":
				task.Journal.NewDatabaseOID++
			case "changed_evidence":
				evidence.Operator = "changed"
			case "missing_verified_custody":
				journal.Acceptance.CustodyVerifiedAt = time.Time{}
			case "incomplete_media_move":
				journal.Moves = append(journal.Moves, moduleRecoveryMediaEntryMove{})
			}
			if scenario != "valid_after_later_configuration" && scenario != "pending_acknowledgement_after_later_configuration" {
				require.False(t, taskHasDurablyReleasedAcceptedProductCatalog(task))
				require.True(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task))
				return
			}
			cfg := previous.Configuration
			cfg.PublicOrigins = []string{"https://later-deployment.example.invalid"}
			manager, err := NewManager(cfg)
			require.NoError(t, err)
			defer manager.Close()
			completed, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "completed", completed.State)
			require.True(t, taskHasDurablyReleasedAcceptedProductCatalog(completed))
			require.False(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(completed))
		})
	}
}

func TestProductCatalogFormalRestoreAcceptanceRejectsOtherScopesAndMissingEvidenceWithoutDatabaseAccess(t *testing.T) {
	manager, task, _ := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
	directory, err := manager.taskDirectory(task.ID)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(directory, 0700))
	for _, scope := range []string{"whole_application", "selected_business_modules", productCatalogBackupScope} {
		t.Run(scope, func(t *testing.T) {
			task.Manifest.Scope = scope
			require.NoError(t, manager.saveTask(task))
			_, err := manager.StartProductCatalogFormalRestoreAcceptance(task.ID, "synthetic", ProductCatalogFormalRestoreAcceptanceRequest{Confirmation: "ACCEPT PRODUCT RESTORE " + manager.Configuration.SiteID + "/" + manager.Configuration.Environment + "/" + task.ID})
			require.Error(t, err)
		})
	}
}
