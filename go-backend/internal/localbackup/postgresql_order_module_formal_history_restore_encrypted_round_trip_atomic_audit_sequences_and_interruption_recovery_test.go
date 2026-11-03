package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/repository"
)

type orderModuleGlobalPaymentReceiptReplayTestCase struct {
	Scope, Path, Key, ResponseBody string
	StatusCode                     int
}

func registeredOrderModuleGlobalPaymentReceiptReplayTestCases() []orderModuleGlobalPaymentReceiptReplayTestCase {
	return []orderModuleGlobalPaymentReceiptReplayTestCase{
		{"paypal_capture", "/api/v1/payment/paypal/orders/original-provider-order/capture", "original-paypal-key", `{"recorded":"original capture response"}`, 200},
		{"alipay_confirm", "/api/v1/payment/alipay/orders/historical-order-to-restore/confirm", "original-alipay-key", `{"recorded":"original confirmation response"}`, 200},
		{"wechat_confirm", "/api/v1/payment/wechat/orders/historical-order-to-restore/confirm", "original-wechat-key", `{"error":"original cached client error"}`, 422},
	}
}

func verifyOrderModuleRestoredGlobalPaymentReceiptsReplayWithoutEnteringPaymentHandler(t *testing.T, db *sql.DB) {
	t.Helper()
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	readPaymentReceiptRowsDigest := func() string {
		var digest string
		require.NoError(t, db.QueryRowContext(context.Background(), `SELECT md5(jsonb_agg(to_jsonb(record) ORDER BY id)::text) FROM payment_operation_idempotencies record`).Scan(&digest))
		return digest
	}
	before := readPaymentReceiptRowsDigest()
	providerHandlerCalls := 0
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", uint(900001)); c.Next() })
	for _, receipt := range registeredOrderModuleGlobalPaymentReceiptReplayTestCases() {
		policy := middleware.PaymentOperationRetryExpiredQuery
		if receipt.Scope == "paypal_capture" {
			policy = middleware.PaymentOperationReconcileExpiredMutation
		}
		router.POST(receipt.Path, middleware.PaymentOperationIdempotency(repository.NewPaymentOperationIdempotencyRepository(orm), receipt.Scope, policy), func(c *gin.Context) {
			providerHandlerCalls++
			c.JSON(http.StatusInternalServerError, gin.H{"error": "must never execute during historical receipt replay"})
		})
		for _, requestBody := range []string{"{}", "{}", `{"different":"request"}`} {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, receipt.Path, strings.NewReader(requestBody))
			request.Header.Set("Idempotency-Key", receipt.Key)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if requestBody == "{}" {
				require.Equal(t, receipt.StatusCode, response.Code)
				require.Equal(t, receipt.ResponseBody, response.Body.String())
				require.Equal(t, "true", response.Header().Get("Idempotency-Replayed"))
				require.Equal(t, "application/json; charset=utf-8", response.Header().Get("Content-Type"))
			} else {
				require.Equal(t, http.StatusConflict, response.Code)
				require.Contains(t, response.Body.String(), "idempotency_key_conflict")
			}
		}
	}
	require.Zero(t, providerHandlerCalls, "the real durable middleware must return before any payment handler or provider interaction")
	require.Equal(t, before, readPaymentReceiptRowsDigest(), "original receipt identities, responses and execution states stay unchanged; normal retry attempts may consume sequence values")
}

func TestPostgreSQLOrderModuleFormalHistoryRestoreAtomicAuditSequencesAndInterruptedAttachmentRecovery(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL and complete application migrations")
	}
	for _, scenario := range []string{"authenticated_restore", "restore_intent_persisted", "attachment_linked_before_acknowledgement", "rows_and_sequences_before_commit", "rows_committed_before_acknowledgement", "verified_before_fence_removal", "fence_removed_before_acknowledgement", "completed_execution_authenticated_restore", "completed_execution_rows_and_sequences_before_commit", "completed_execution_rows_committed_before_acknowledgement", "newer_order_notes_preserved", "completed_execution_invalid_notification_identity", "completed_execution_invalid_event_order_payload", "global_payment_authenticated_restore", "global_payment_rows_and_sequences_before_commit", "global_payment_rows_committed_before_acknowledgement", "global_payment_invalid_natural_key_identity", "global_payment_invalid_current_response", "global_payment_invalid_user_dependency", "global_payment_invalid_pending_receipt", "global_payment_invalid_recorded_response", "settled_active_paid_authenticated_restore", "settled_active_delivered_authenticated_restore", "settled_active_paid_missing_processed_event_restore", "settled_active_processing_authenticated_restore", "settled_active_shipped_authenticated_restore", "settled_active_paid_newer_order_notes_preserved", "settled_active_paid_rows_and_sequences_before_commit", "settled_active_paid_rows_committed_before_acknowledgement", "settled_active_paid_invalid_pending_transaction", "settled_active_paid_invalid_missing_payment_completion", "settled_active_paid_invalid_stock_dependency", "settled_active_paid_invalid_current_status", "settled_active_paid_invalid_current_payment", "settled_active_paid_invalid_component_allocations", "settled_active_shipped_invalid_missing_shipping_evidence", "settled_active_paid_invalid_new_current_refund", "settled_active_paid_invalid_fulfillment_hold", "settled_active_paid_component_authenticated_restore", "settled_active_paid_component_rows_and_sequences_before_commit", "settled_active_paid_component_rows_committed_before_acknowledgement", "settled_active_paid_primary_master_authenticated_restore", "settled_active_paid_invalid_component_stock_dependency", "settled_active_paid_invalid_master_stock_dependency", "settled_active_paid_invalid_component_master_mapping", "settled_active_paid_invalid_component_product_dependency", "settled_active_paid_invalid_primary_master_stock_dependency"} {
		t.Run(scenario, func(t *testing.T) {
			settledActiveOrder := strings.HasPrefix(scenario, "settled_active_")
			activeOrderStatus, activeScenario, _ := strings.Cut(strings.TrimPrefix(scenario, "settled_active_"), "_")
			configuredInventoryOrder := settledActiveOrder && (strings.Contains(activeScenario, "component_") || strings.Contains(activeScenario, "master_")) && activeScenario != "invalid_component_allocations"
			primaryMasterInventoryOrder := strings.Contains(activeScenario, "primary_master_")
			deliveredActiveOrder := settledActiveOrder && activeOrderStatus == "delivered"
			if deliveredActiveOrder {
				activeOrderStatus = "shipped"
			}
			missingActiveProcessedEvent := settledActiveOrder && activeScenario == "missing_processed_event_restore"
			preserveCurrentOrderNotes := scenario == "newer_order_notes_preserved" || settledActiveOrder && activeScenario == "newer_order_notes_preserved"
			globalPaymentReceipts := strings.HasPrefix(scenario, "global_payment_")
			completedExecutionHistory := strings.HasPrefix(scenario, "completed_execution_") || globalPaymentReceipts
			checkpointScenario := strings.TrimPrefix(scenario, "completed_execution_")
			checkpointScenario = strings.TrimPrefix(checkpointScenario, "global_payment_")
			if settledActiveOrder {
				checkpointScenario = strings.TrimPrefix(activeScenario, "component_")
				checkpointScenario = strings.TrimPrefix(checkpointScenario, "primary_master_")
			}
			if preserveCurrentOrderNotes || missingActiveProcessedEvent {
				checkpointScenario = "authenticated_restore"
			}
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_order_restore_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_order_app_" + newTaskID()[:12]
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
			require.NoError(t, db.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			_, err = db.ExecContext(ctx, `
INSERT INTO users(id,email,username,password) VALUES(900001,'order-restore@example.invalid','synthetic-order-customer','synthetic-unused-password');
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900001,'Current product','order-restore-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'order-restore-sku',123,'USD',19);
INSERT INTO orders(id,order_number,user_id,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'historical-order-to-restore',900001,'completed','USD','USD',123,123,123,0,0,0,0),(900002,'current-order-to-preserve',900001,'pending','USD','USD',123,123,123,0,0,0,0);
INSERT INTO order_items(id,order_id,product_id,variant_id,product_name,quantity,currency,price_minor,subtotal_minor,total_minor,tax_amount_minor,discount_minor) VALUES(900001,900001,900001,900001,'Historical product description',1,'USD',123,123,123,0,0),(900002,900002,900001,900001,'Current product description',1,'USD',123,123,123,0,0);
INSERT INTO transactions(id,order_id,payment_method,amount_minor,currency,status) VALUES(900001,900001,'synthetic-provider',123,'USD','completed');
INSERT INTO refunds(id,order_id,transaction_id,amount_minor,requested_amount_minor,discount_clawback_amount_minor,currency,status) VALUES(900001,900001,900001,10,10,0,'USD','completed');
INSERT INTO payment_operation_idempotencies(user_id,scope,idempotency_key,request_hash,status,status_code,response_body) VALUES(900001,'synthetic-capture','order-restore-payment-key',repeat('a',64),'completed',200,'private completed payment receipt');
INSERT INTO refund_idempotencies(admin_user_id,scope,idempotency_key,request_hash,refund_id) VALUES(900001,'synthetic-refund','order-restore-refund-key',repeat('b',64),900001);
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('order-restore-completed','order.completed','order','900001','{"order_id":900001,"order_number":"historical-order-to-restore","schema_version":1,"idempotency_key":"order-restore-completion"}','processed',now());
INSERT INTO tickets(id,ticket_number,user_id,customer_user_id,conversation_id,subject,category) VALUES(900001,'order-restore-chat',900001,900001,'order-restore-chat','Retained conversation','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Latest retained customer message','[]','media_asset_bindings');
INSERT INTO loyalty_transactions(user_id,type,source,source_id,points,balance,description) VALUES(900001,'earn','referral',900001,5,5,'Retain registration reward');
INSERT INTO order_evidence_snapshots(id,order_id,schema_version,confirmed_at,currency,snapshot_data,snapshot_sha256) VALUES(900001,900001,1,now(),'USD','{}',repeat('a',64));
INSERT INTO order_evidence_packages(id,order_id,snapshot_id,package_version,schema_version) VALUES(900001,900001,900001,1,1);
INSERT INTO order_evidence_items(id,package_id,order_id,item_type,status,required_reason) VALUES(900001,900001,900001,'signed_pod','draft','base');
INSERT INTO order_evidence_attachments(id,evidence_item_id,storage_key,original_filename,mime_type,size_bytes,sha256) VALUES(900001,900001,'order-evidence/900001/900001/synthetic-proof.bin','proof.bin','application/octet-stream',8,encode(sha256('original'::bytea),'hex'));
`)
			require.NoError(t, err)
			if completedExecutionHistory {
				_, err = db.ExecContext(ctx, `
UPDATE transactions SET transaction_id='synthetic-provider-payment',completed_at=now() WHERE id=900001;
UPDATE refunds SET refund_id='synthetic-provider-refund',completed_at=now() WHERE id=900001;
INSERT INTO payment_refund_executions(id,refund_id,order_id,transaction_id,provider,provider_payment_id,merchant_order_number,provider_transaction_id,amount_minor,currency,status,idempotency_key,attempt_count,requested_by_id,requested_at,completed_at,provider_refund_id) VALUES(900001,900001,900001,900001,'synthetic-provider','synthetic-provider-payment','historical-order-to-restore','synthetic-provider-payment',10,'USD','succeeded','synthetic-refund-execution-key',1,900001,now(),now(),'synthetic-provider-refund');
INSERT INTO order_idempotencies(id,user_id,scope,idempotency_key,request_hash,order_id) VALUES(900001,900001,'synthetic-checkout','synthetic-checkout-request',repeat('c',64),900001);
INSERT INTO shipment_records(id,order_id,order_number,shipped_at,warranty_start_at,warranty_expires) VALUES(900001,900001,'historical-order-to-restore',now(),now(),now()+interval '1 year');
INSERT INTO email_delivery_records(id,outbox_event_id,event_key,event_type,template_code,recipient_email,reference_type,reference_number,status,first_attempt_at,last_attempt_at,sent_at) SELECT 900001,id,'notification:order-restore-completion:order_completed','order.completed','order_completed','synthetic-order@example.invalid','order','historical-order-to-restore','sent',now(),now(),now() FROM outbox_events WHERE event_key='order-restore-completed';
`)
				require.NoError(t, err)
				if scenario == "completed_execution_invalid_notification_identity" {
					_, err = db.ExecContext(ctx, `UPDATE email_delivery_records SET event_key='synthetic-other-notification' WHERE id=900001`)
					require.NoError(t, err)
				}
				if scenario == "completed_execution_invalid_event_order_payload" {
					_, err = db.ExecContext(ctx, `UPDATE outbox_events SET payload=payload||'{"order_id":900002}'::jsonb WHERE event_key='order-restore-completed'`)
					require.NoError(t, err)
				}
			}
			if globalPaymentReceipts {
				for position, receipt := range registeredOrderModuleGlobalPaymentReceiptReplayTestCases() {
					digest := sha256.Sum256([]byte("POST\n" + receipt.Path + "\n\n{}"))
					_, err = db.ExecContext(ctx, `INSERT INTO payment_operation_idempotencies(id,user_id,scope,idempotency_key,request_hash,status,status_code,content_type,response_body,created_at,updated_at) VALUES($1,900001,$2,$3,$4,'completed',$5,'application/json; charset=utf-8',$6,now(),now())`, 900010+position, receipt.Scope, receipt.Key, fmt.Sprintf("%x", digest), receipt.StatusCode, receipt.ResponseBody)
					require.NoError(t, err)
				}
				if scenario == "global_payment_invalid_pending_receipt" {
					_, err = db.ExecContext(ctx, `UPDATE payment_operation_idempotencies SET status='pending' WHERE id=900010`)
					require.NoError(t, err)
				}
				if scenario == "global_payment_invalid_recorded_response" {
					_, err = db.ExecContext(ctx, `UPDATE payment_operation_idempotencies SET response_body='private truncated response' WHERE id=900010`)
					require.NoError(t, err)
				}
			}
			attachment := filepath.Join(cfg.StorageRoots["uploads_private"], "order-evidence/900001/900001/synthetic-proof.bin")
			if settledActiveOrder {
				_, err = db.ExecContext(ctx, `UPDATE orders SET status=$1,payment_status='paid',payment_method='synthetic-provider',paid_at=now() WHERE id=900001`, activeOrderStatus)
				require.NoError(t, err)
				_, err = db.ExecContext(ctx, `DELETE FROM refund_idempotencies WHERE refund_id=900001; DELETE FROM refunds WHERE id=900001;
UPDATE transactions SET transaction_id='synthetic-active-provider-payment',completed_at=now() WHERE id=900001;
UPDATE outbox_events SET event_type='order.payment_succeeded' WHERE event_key='order-restore-completed'`)
				require.NoError(t, err)
				if activeOrderStatus == "shipped" {
					_, err = db.ExecContext(ctx, `UPDATE orders SET shipping_status='shipped',shipped_at=now() WHERE id=900001;
INSERT INTO shipment_records(id,order_id,order_number,shipped_at,warranty_start_at,warranty_expires) VALUES(900001,900001,'historical-order-to-restore',now(),now(),now()+interval '1 year')`)
					require.NoError(t, err)
				}
				if deliveredActiveOrder {
					_, err = db.ExecContext(ctx, `UPDATE orders SET shipping_status='delivered',delivered_at=now() WHERE id=900001`)
					require.NoError(t, err)
				}
				if configuredInventoryOrder {
					_, err = db.ExecContext(ctx, `INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900010,'Original component product','order-restore-component-product',11,'USD');
INSERT INTO product_variants(id,product_id,sku,option_values,price_minor,currency,stock) VALUES(900011,900010,'original-component-master','{"kind":"master"}',11,'USD',11),(900013,900010,'other-current-master','{"kind":"alternative-master"}',13,'USD',13);
INSERT INTO product_variants(id,product_id,master_variant_id,sku,option_values,price_minor,currency,stock) VALUES(900010,900010,900011,'original-component-translation','{"kind":"component-translation"}',11,'USD',0),(900012,900001,900011,'original-primary-translation','{"kind":"primary-translation"}',123,'USD',0)`)
					require.NoError(t, err)
					orderItemVariantID := int64(900001)
					orderItemConfigurationSnapshot := `{"inventory_allocations":[{"variant_id":900010,"quantity":2,"group_slug":"original-components","value_key":"original"}]}`
					if primaryMasterInventoryOrder {
						orderItemVariantID = 900012
						orderItemConfigurationSnapshot = `{}`
					}
					if activeScenario == "invalid_component_stock_dependency" {
						orderItemConfigurationSnapshot = `{"inventory_allocations":[{"variant_id":900011,"quantity":2}]}`
					}
					_, err = db.ExecContext(ctx, `UPDATE order_items SET variant_id=$1,configuration_snapshot=$2::jsonb WHERE id=900001`, orderItemVariantID, orderItemConfigurationSnapshot)
					require.NoError(t, err)
				}
				sourceChanges := map[string]string{
					"invalid_pending_transaction":        `UPDATE transactions SET status='pending' WHERE id=900001`,
					"invalid_missing_payment_completion": `UPDATE transactions SET completed_at=NULL WHERE id=900001`,
					"invalid_component_allocations":      `UPDATE order_items SET configuration_snapshot='{"inventory_allocations":[{"variant_id":900002,"quantity":1}]}' WHERE id=900001`,
					"invalid_missing_shipping_evidence":  `DELETE FROM shipment_records WHERE id=900001`,
					"invalid_fulfillment_hold":           `UPDATE orders SET fulfillment_hold=true WHERE id=900001`,
				}
				if query := sourceChanges[activeScenario]; query != "" {
					_, err = db.ExecContext(ctx, query)
					require.NoError(t, err)
				}
			}
			if activeScenario == "component_authenticated_restore" {
				t.Run("older_package_without_original_master_dependency_is_rejected", func(t *testing.T) {
					tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
					require.NoError(t, err)
					defer tx.Rollback()
					snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, []string{"orders"})
					require.NoError(t, err)
					originalDependencies := snapshot.ExternalDependencies
					snapshot.ExternalDependencies = nil
					for _, reference := range originalDependencies {
						if reference.Table != "product_variants" || reference.Values["id"] != "900011" {
							snapshot.ExternalDependencies = append(snapshot.ExternalDependencies, reference)
						}
					}
					require.Len(t, snapshot.ExternalDependencies, len(originalDependencies)-1)
					observation, err := inspectModuleRecoveryDatabaseSchemaWithoutRowFingerprints(ctx, tx)
					require.NoError(t, err)
					actualDependencies, err := captureSelectedBusinessModuleExternalReferenceFingerprintsWithObservedSchema(ctx, tx, snapshot, observation)
					require.NoError(t, err)
					report := &SelectedBusinessModuleIsolatedRestorePreview{}
					classifySelectedBusinessModuleExternalDependencyIssues(snapshot, observation, actualDependencies, report)
					require.Equal(t, 1, report.DependencyConflicts)
					require.Equal(t, "unexpected_target_reference", report.DependencyIssues[0].Reason)
				})
			}
			require.NoError(t, os.MkdirAll(filepath.Dir(attachment), 0700))
			require.NoError(t, os.WriteFile(attachment, []byte("original"), 0600))
			if scenario == "authenticated_restore" {
				tx, err := db.BeginTx(ctx, nil)
				require.NoError(t, err)
				observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
				require.NoError(t, err)
				original, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, []string{"orders"})
				require.NoError(t, err)
				var originalTerminalOrder json.RawMessage
				for _, row := range original.Tables["orders"].Rows {
					id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
					if id == "900001" {
						originalTerminalOrder = row
					}
				}
				require.NotEmpty(t, originalTerminalOrder)
				for _, difference := range []string{"newer_notes_only", "older_notes", "status", "address", "financial_amount", "deleted", "timestamp_without_notes"} {
					t.Run("note_preservation_"+difference, func(t *testing.T) {
						_, err := tx.ExecContext(ctx, `SAVEPOINT order_note_preservation; SET LOCAL session_replication_role='replica'; UPDATE orders SET customer_note='Different note',updated_at=now()+interval '1 second' WHERE id=900001`)
						require.NoError(t, err)
						changes := map[string]string{
							"older_notes":             `UPDATE orders SET updated_at=now()-interval '1 day' WHERE id=900001`,
							"status":                  `UPDATE orders SET status='refunded' WHERE id=900001`,
							"address":                 `UPDATE orders SET shipping_email='changed@example.invalid' WHERE id=900001`,
							"financial_amount":        `UPDATE orders SET payment_amount_minor=payment_amount_minor+1 WHERE id=900001`,
							"deleted":                 `UPDATE orders SET deleted_at=now() WHERE id=900001`,
							"timestamp_without_notes": `UPDATE orders SET customer_note=NULL WHERE id=900001`,
						}
						if query := changes[difference]; query != "" {
							_, err = tx.ExecContext(ctx, query)
							require.NoError(t, err)
						}
						preserve, err := canPreserveCurrentRestorableOrderNotesWithOtherwiseIdenticalFacts(ctx, tx, original, "orders", originalTerminalOrder)
						require.NoError(t, err)
						require.Equal(t, difference == "newer_notes_only", preserve)
						_, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT order_note_preservation; RELEASE SAVEPOINT order_note_preservation`)
						require.NoError(t, err)
					})
				}
				conflicts, err := countOrderHistoryCurrentExecutionScopesWithRowsAbsentFromPackage(ctx, tx, original, observation)
				require.NoError(t, err)
				require.Zero(t, conflicts)
				_, err = tx.ExecContext(ctx, `INSERT INTO orders(id,order_number,user_id,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900003,'unrelated-new-order',900001,'USD','USD',123,123,123,0,0,0,0); INSERT INTO transactions(id,order_id,payment_method,amount_minor,currency,status) VALUES(900003,900003,'synthetic-provider',123,'USD','completed')`)
				require.NoError(t, err)
				conflicts, err = countOrderHistoryCurrentExecutionScopesWithRowsAbsentFromPackage(ctx, tx, original, observation)
				require.NoError(t, err)
				require.Zero(t, conflicts, "unrelated new order and transaction remain independent")
				_, err = tx.ExecContext(ctx, `INSERT INTO transactions(id,order_id,payment_method,amount_minor,currency,status) VALUES(900004,900001,'synthetic-provider',123,'USD','completed')`)
				require.NoError(t, err)
				conflicts, err = countOrderHistoryCurrentExecutionScopesWithRowsAbsentFromPackage(ctx, tx, original, observation)
				require.NoError(t, err)
				require.Positive(t, conflicts, "new execution facts for a package order must block historical resurrection")
				scope := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}}
				require.NoError(t, validateOrderModuleInsertOnlyWriteContract(ctx, tx, scope, "orders", observation), "existing UPDATE-only immutable snapshot guards cannot fire on INSERT")
				_, err = tx.ExecContext(ctx, `CREATE FUNCTION public.synthetic_order_insert_side_effect() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'must not execute'; END $$; CREATE TRIGGER synthetic_order_insert_guard BEFORE INSERT ON public.orders FOR EACH ROW EXECUTE FUNCTION public.synthetic_order_insert_side_effect()`)
				require.NoError(t, err)
				require.ErrorContains(t, validateOrderModuleInsertOnlyWriteContract(ctx, tx, scope, "orders", observation), "INSERT triggers")
				require.NoError(t, tx.Rollback())
			}
			require.NoError(t, db.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-exporter", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/orders", []string{"orders"})
			require.NoError(t, err)
			backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
			require.Equal(t, "ready_to_download", backup.State, backup.Error)
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			exporter, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-exporter", []string{PermissionExport})
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/downloads/"+backup.ID, nil)
			request.Header.Set("Authorization", "Bearer "+exporter)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)
			require.NotContains(t, response.Body.String(), "Historical product description", "download contains encrypted data only")
			db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, `DELETE FROM order_evidence_attachments WHERE id=900001; DELETE FROM order_evidence_items WHERE id=900001; DELETE FROM order_evidence_packages WHERE id=900001; DELETE FROM order_evidence_snapshots WHERE id=900001;`)
			require.NoError(t, err)
			if preserveCurrentOrderNotes {
				_, err = db.ExecContext(ctx, `UPDATE orders SET customer_note='New customer note after backup',admin_note='New operator note after backup',updated_at=now()+interval '1 second' WHERE id=900001`)
			} else if settledActiveOrder && activeScenario == "invalid_current_status" {
				_, err = db.ExecContext(ctx, `UPDATE orders SET status='processing' WHERE id=900001`)
			} else {
				_, err = db.ExecContext(ctx, `DELETE FROM order_items WHERE id=900001; DELETE FROM orders WHERE id=900001;`)
			}
			require.NoError(t, err)
			if completedExecutionHistory {
				_, err = db.ExecContext(ctx, `DELETE FROM email_delivery_records WHERE id=900001; DELETE FROM outbox_events WHERE event_key='order-restore-completed'; DELETE FROM shipment_records WHERE id=900001; DELETE FROM order_idempotencies WHERE id=900001; DELETE FROM payment_refund_executions WHERE id=900001; DELETE FROM refund_idempotencies WHERE refund_id=900001; DELETE FROM refunds WHERE id=900001; DELETE FROM transactions WHERE id=900001;`)
				require.NoError(t, err)
			}
			if globalPaymentReceipts {
				switch scenario {
				case "global_payment_invalid_natural_key_identity":
					_, err = db.ExecContext(ctx, `UPDATE payment_operation_idempotencies SET id=900020 WHERE id=900010; DELETE FROM payment_operation_idempotencies WHERE id IN (900011,900012)`)
				case "global_payment_invalid_current_response":
					_, err = db.ExecContext(ctx, `UPDATE payment_operation_idempotencies SET response_body='{"current":"newer response"}' WHERE id=900010; DELETE FROM payment_operation_idempotencies WHERE id IN (900011,900012)`)
				default:
					_, err = db.ExecContext(ctx, `DELETE FROM payment_operation_idempotencies WHERE id BETWEEN 900010 AND 900012`)
				}
				require.NoError(t, err)
				if scenario == "global_payment_invalid_user_dependency" {
					_, err = db.ExecContext(ctx, `UPDATE users SET username='changed-current-customer' WHERE id=900001`)
					require.NoError(t, err)
				}
			}
			if settledActiveOrder {
				_, err = db.ExecContext(ctx, `DELETE FROM shipment_records WHERE id=900001`)
				require.NoError(t, err)
				switch activeScenario {
				case "invalid_current_payment":
					_, err = db.ExecContext(ctx, `UPDATE transactions SET amount_minor=124 WHERE id=900001`)
				case "invalid_new_current_refund":
					_, err = db.ExecContext(ctx, `INSERT INTO refunds(id,order_id,transaction_id,amount_minor,requested_amount_minor,discount_clawback_amount_minor,currency,status) VALUES(900002,900001,900001,1,1,0,'USD','pending')`)
				default:
					_, err = db.ExecContext(ctx, `DELETE FROM transactions WHERE id=900001`)
				}
				require.NoError(t, err)
				if missingActiveProcessedEvent {
					_, err = db.ExecContext(ctx, `DELETE FROM outbox_events WHERE event_key='order-restore-completed'`)
					require.NoError(t, err)
				}
				if configuredInventoryOrder {
					currentInventoryChanges := map[string]string{
						"invalid_component_stock_dependency":      `UPDATE product_variants SET stock=stock+1 WHERE id=900011`,
						"invalid_master_stock_dependency":         `UPDATE product_variants SET stock=stock+1 WHERE id=900011`,
						"invalid_primary_master_stock_dependency": `UPDATE product_variants SET stock=stock+1 WHERE id=900011`,
						"invalid_component_master_mapping":        `UPDATE product_variants SET master_variant_id=900013 WHERE id=900010`,
						"invalid_component_product_dependency":    `UPDATE products SET name='Changed current component product' WHERE id=900010`,
					}
					if query := currentInventoryChanges[activeScenario]; query != "" {
						_, err = db.ExecContext(ctx, query)
						require.NoError(t, err)
					}
				}
				if activeScenario == "invalid_stock_dependency" {
					_, err = db.ExecContext(ctx, `UPDATE product_variants SET stock=stock+1 WHERE id=900001`)
					require.NoError(t, err)
				}
			}
			require.NoError(t, os.Remove(attachment))
			for directory, count := filepath.Dir(attachment), 0; count < 3; directory, count = filepath.Dir(directory), count+1 {
				require.NoError(t, os.Remove(directory))
			}
			beforeConstraintCatalog := readOrderModuleNormalConstraintTestCatalogIdentity(t, db)
			_, beforeFingerprint, _, err := captureDatabaseInventory(ctx, db)
			require.NoError(t, err)
			require.NoError(t, db.Close())
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, response.Body.Bytes())
			require.Equal(t, "module_package_integrity_verified", task.State, task.Error)
			_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID)
			require.NoError(t, err)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			reportBytes, err := json.Marshal(task.SelectedModuleRestorePreview)
			require.NoError(t, err)
			if globalPaymentReceipts {
				require.Contains(t, strings.Join(task.SelectedModuleRestorePreview.Limitations, "\n"), "已完成全局支付请求凭据")
				require.NotContains(t, strings.Join(task.SelectedModuleRestorePreview.Limitations, "\n"), "正式分模块覆盖导入及从组合包再次选择子模块尚未开放")
			}
			if strings.Contains(scenario, "_invalid_") {
				require.Equal(t, "module_restore_conflicts_found", task.State, "%s: %s", task.Error, reportBytes)
				require.False(t, task.SelectedModuleRestorePreview.OrderHistoryRestoreEligibility.Eligible)
				if !globalPaymentReceipts && !settledActiveOrder {
					require.Positive(t, task.SelectedModuleRestorePreview.OrderHistoryDomainBlockingIssues)
				}
				require.Zero(t, task.SelectedModuleRestorePreview.RestoredRows)
				require.True(t, task.SelectedModuleRestorePreview.DisposableStageRemoved)
				_, err = manager.StartOrderModuleFormalHistoryRestore(task.ID, "synthetic-order-restorer", OrderModuleFormalHistoryRestoreRequest{Confirmation: "RESTORE ORDER HISTORY " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256})
				require.Error(t, err)
				db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, actual, _, err := captureDatabaseInventory(ctx, db)
				require.NoError(t, err)
				require.Equal(t, beforeFingerprint, actual)
				require.NoError(t, db.Close())
				fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
				require.NoError(t, err)
				require.False(t, fenced)
				return
			}
			require.Equal(t, "module_isolated_restore_verified", task.State, "%s: %s", task.Error, string(reportBytes))
			require.True(t, task.SelectedModuleRestorePreview.OrderHistoryRestoreEligibility.Eligible, "%+v", task.SelectedModuleRestorePreview.OrderHistoryRestoreEligibility)
			if settledActiveOrder {
				require.Equal(t, 1, task.SelectedModuleRestorePreview.OrderHistoryRestoreEligibility.SettledActiveOrderOwners)
			}
			if preserveCurrentOrderNotes {
				var preserved int
				for _, table := range task.SelectedModuleRestorePreview.Tables {
					preserved += table.PreservedCurrentOrderNoteRows
				}
				require.Equal(t, 1, preserved)
			}
			restoreRequest := OrderModuleFormalHistoryRestoreRequest{Confirmation: "RESTORE ORDER HISTORY " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256}
			invokeOrderRestoreHTTP := func(action, token string, body any, expected int) {
				t.Helper()
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+task.ID+"/"+action, strings.NewReader(string(encoded)))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, expected, response.Code, response.Body.String())
			}
			for _, permission := range []string{PermissionView, PermissionUpload, PermissionExport} {
				token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-unprivileged", []string{permission})
				require.NoError(t, err)
				invokeOrderRestoreHTTP("order-module-formal-history-restore", token, restoreRequest, http.StatusForbidden)
				invokeOrderRestoreHTTP("order-module-interrupted-restore-recovery", token, map[string]string{"confirmation": "invalid"}, http.StatusForbidden)
			}
			restoreToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-order-restorer", []string{PermissionRestore})
			require.NoError(t, err)
			stale := restoreRequest
			stale.ExpectedPreviewSHA256 = strings.Repeat("0", 64)
			invokeOrderRestoreHTTP("order-module-formal-history-restore", restoreToken, stale, http.StatusConflict)
			if checkpointScenario == "authenticated_restore" {
				invokeOrderRestoreHTTP("order-module-formal-history-restore", restoreToken, restoreRequest, http.StatusOK)
			} else {
				require.NoError(t, manager.startTaskOperation(task, "synthetic_order_history_restore", func(ctx context.Context, task *Task) error {
					return manager.restoreOrderHistoryWithAtomicReceiptAndRecoverableFiles(ctx, task, "synthetic-order-restorer", func(point string) error {
						if point == checkpointScenario {
							return errors.New("synthetic interrupted order restore")
						}
						return nil
					})
				}))
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			committed := checkpointScenario == "authenticated_restore" || checkpointScenario == "rows_committed_before_acknowledgement" || checkpointScenario == "verified_before_fence_removal" || checkpointScenario == "fence_removed_before_acknowledgement"
			if checkpointScenario != "authenticated_restore" {
				require.Equal(t, "failed", task.State, task.Error)
				require.Equal(t, "synthetic interrupted order restore", task.Error)
				require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
				_, err = manager.VerifyUploadedPackage(task.ID)
				require.Error(t, err)
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, manager.ExpireWorkCopies(ctx))
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				require.False(t, task.WorkCopyExpired, "interrupted restore materials cannot expire")
				if !committed {
					db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
					require.NoError(t, err)
					_, actual, _, err := captureDatabaseInventory(ctx, db)
					require.NoError(t, err)
					require.Equal(t, beforeFingerprint, actual, "rows, audit and sequence changes roll back together")
					require.NoError(t, db.Close())
				}
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				task, err = manager.GetTask(task.ID)
				require.NoError(t, err)
				if checkpointScenario == "fence_removed_before_acknowledgement" {
					require.Equal(t, "completed", task.State)
				} else {
					require.Equal(t, "interrupted", task.State)
					invokeOrderRestoreHTTP("order-module-interrupted-restore-recovery", restoreToken, map[string]string{"confirmation": "RECOVER ORDER RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID}, http.StatusOK)
					task = waitForIndependentRecoveryTask(t, manager, task.ID)
				}
			}
			if committed {
				require.Equal(t, "completed", task.State, task.Error)
				expectedRows := 6
				if settledActiveOrder {
					expectedRows++
					if activeOrderStatus == "shipped" {
						expectedRows++
					}
				}
				if missingActiveProcessedEvent {
					expectedRows++
				}
				if completedExecutionHistory {
					expectedRows += 8
				}
				if globalPaymentReceipts {
					expectedRows += 3
				}
				if preserveCurrentOrderNotes {
					expectedRows -= 2
				}
				require.Equal(t, expectedRows, task.OrderModuleRestore.RestoredRows)
			} else {
				require.Equal(t, "rolled_back", task.State, task.Error)
				require.Zero(t, task.OrderModuleRestore.RestoredRows)
			}
			require.NoError(t, validateOrderHistoryRestoreProgress(task))
			require.Equal(t, "released", task.OrderModuleRestore.Phase)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			staged, err := filepath.Glob(filepath.Join(cfg.StorageRoots["uploads_private"], ".local-backup-order-history-*"))
			require.NoError(t, err)
			require.Empty(t, staged)
			db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer db.Close()
			require.Equal(t, beforeConstraintCatalog, readOrderModuleNormalConstraintTestCatalogIdentity(t, db), "order rehearsal and formal restore must preserve constraint and trigger object identities")
			var orders, audits, events, stock, reward int
			var chat, receipt string
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM orders WHERE id=900001`).Scan(&orders))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM audit_logs WHERE resource='order_module_history_restore'`).Scan(&audits))
			require.Equal(t, map[bool]int{true: 1, false: 0}[committed], orders)
			require.Equal(t, orders, audits, "restore audit exists exactly when order rows committed")
			require.NoError(t, db.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900001`).Scan(&stock))
			require.Equal(t, 19, stock)
			if configuredInventoryOrder {
				var masterStock, translationStock int
				require.NoError(t, db.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900011`).Scan(&masterStock))
				require.NoError(t, db.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900010`).Scan(&translationStock))
				require.Equal(t, 11, masterStock)
				require.Zero(t, translationStock)
			}
			require.NoError(t, db.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&chat))
			require.Equal(t, "Latest retained customer message", chat)
			require.NoError(t, db.QueryRowContext(ctx, `SELECT points FROM loyalty_transactions WHERE source='referral'`).Scan(&reward))
			require.Equal(t, 5, reward)
			require.NoError(t, db.QueryRowContext(ctx, `SELECT response_body FROM payment_operation_idempotencies WHERE idempotency_key='order-restore-payment-key'`).Scan(&receipt))
			require.Equal(t, "private completed payment receipt", receipt)
			if globalPaymentReceipts {
				var restoredReceipts int
				require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM payment_operation_idempotencies WHERE id BETWEEN 900010 AND 900012`).Scan(&restoredReceipts))
				require.Equal(t, map[bool]int{true: 3, false: 0}[committed], restoredReceipts)
				if committed {
					verifyOrderModuleRestoredGlobalPaymentReceiptsReplayWithoutEnteringPaymentHandler(t, db)
				}
			}
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE event_key='order-restore-completed' AND status='processed'`).Scan(&events))
			if completedExecutionHistory && !committed {
				require.Zero(t, events)
			} else {
				require.Equal(t, 1, events)
			}
			if (completedExecutionHistory || settledActiveOrder) && committed {
				var pending, executions, notifications int
				require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE status<>'processed'`).Scan(&pending))
				require.Zero(t, pending, "restored history cannot enter the runnable queue")
				require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM payment_refund_executions WHERE status='succeeded' AND provider_refund_id='synthetic-provider-refund'`).Scan(&executions))
				require.Equal(t, map[bool]int{true: 1, false: 0}[completedExecutionHistory], executions)
				require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM email_delivery_records WHERE status='sent' AND sent_at IS NOT NULL`).Scan(&notifications))
				require.Equal(t, map[bool]int{true: 1, false: 0}[completedExecutionHistory], notifications)
				_, beforeConsumerFingerprint, _, err := captureDatabaseInventory(ctx, db)
				require.NoError(t, err)
				orm, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
				require.NoError(t, err)
				claimed, err := repository.NewOutboxRepository(orm).ClaimReadyEvents(time.Now().Add(24*time.Hour), "synthetic-recovery-no-replay-worker", 10, time.Minute)
				require.NoError(t, err)
				require.Empty(t, claimed, "real consumer must not claim restored processed history")
				if settledActiveOrder {
					orders := repository.NewOrderRepository(orm)
					restored, err := orders.FindByID(900001)
					require.NoError(t, err)
					require.Equal(t, activeOrderStatus, restored.Status)
					require.Equal(t, "paid", restored.PaymentStatus)
					if deliveredActiveOrder {
						require.Equal(t, "delivered", restored.ShippingStatus)
						require.NotNil(t, restored.DeliveredAt)
					}
					require.Len(t, restored.Items, 1)
					candidates, err := orders.FindPaymentExpirationCandidates(time.Now().Add(24*time.Hour), 100)
					require.NoError(t, err)
					for _, candidate := range candidates {
						require.NotEqual(t, uint(900001), candidate.ID, "restored paid order cannot release inventory through unpaid expiration")
					}
				}
				_, afterConsumerFingerprint, _, err := captureDatabaseInventory(ctx, db)
				require.NoError(t, err)
				require.Equal(t, beforeConsumerFingerprint, afterConsumerFingerprint)
			}
			if preserveCurrentOrderNotes {
				var customerNote, adminNote string
				require.NoError(t, db.QueryRowContext(ctx, `SELECT customer_note,admin_note FROM orders WHERE id=900001`).Scan(&customerNote, &adminNote))
				require.Equal(t, "New customer note after backup", customerNote)
				require.Equal(t, "New operator note after backup", adminNote)
			}
			if committed {
				data, err := os.ReadFile(attachment)
				require.NoError(t, err)
				require.Equal(t, "original", string(data))
				var nextOrderID int64
				if preserveCurrentOrderNotes {
					var called bool
					require.NoError(t, db.QueryRowContext(ctx, `SELECT last_value,is_called FROM orders_id_seq`).Scan(&nextOrderID, &called))
					require.Equal(t, int64(1), nextOrderID, "orders sequence is untouched when no order row is inserted")
					require.False(t, called)
				} else {
					require.NoError(t, db.QueryRowContext(ctx, `SELECT nextval('orders_id_seq')`).Scan(&nextOrderID))
					require.Greater(t, nextOrderID, int64(900002))
				}
			} else {
				_, actual, _, err := captureDatabaseInventory(ctx, db)
				require.NoError(t, err)
				require.Equal(t, beforeFingerprint, actual)
				_, err = os.Stat(attachment)
				require.ErrorIs(t, err, os.ErrNotExist)
			}
			_, err = manager.StartInterruptedOrderHistoryRestoreRecovery(task.ID, "synthetic-order-restorer", "RECOVER ORDER RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID)
			require.Error(t, err, "released tasks cannot replay or roll back new business")
		})
	}
}
