package localbackup

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLCompleteMigrationSchemaBacksUpAndRestoresWithoutChangingTheSource(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL 15 server and the complete migration directory")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = connectionURL
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	cfg.MaxFiles = 10000
	cfg.MaxExpandedBytes = 256 << 20
	cfg.MaxWorkspaceBytes = 1 << 30
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	stageName := "local_backup_stage_" + newTaskID()[:16]
	defer func() {
		for _, name := range []string{cfg.DatabaseName, stageName} {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
		}
	}()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
	require.NoError(t, err)
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
	sourceCloseError, databaseCloseError := migrator.Close()
	require.NoError(t, sourceCloseError)
	require.NoError(t, databaseCloseError)
	migratedDatabase, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	require.NoError(t, migratedDatabase.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
	_, err = migratedDatabase.ExecContext(ctx, `INSERT INTO product_view_count_flush_receipts(batch_id,product_id,delta,product_increment_applied) VALUES('22222222-2222-4222-8222-222222222222',999999,9,false)`)
	require.NoError(t, err)
	require.NoError(t, migratedDatabase.Close())
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	backup, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Contains(t, []string{"ready_to_download", "generated_with_gaps"}, backup.State, backup.Error)
	require.Greater(t, len(backup.Manifest.Tables), 100, "exercise the complete application schema")
	// Seeded public-media URLs may have no local file fixture. Such a package is
	// never applied; this test separately verifies full-schema dump/restore and
	// all registered reference queries against the actual migration columns.
	packagePath, err := manager.taskPackagePath(backup.ID)
	require.NoError(t, err)
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	require.NoError(t, err)
	extracted := filepath.Join(t.TempDir(), "verified-full-schema")
	manifest, err := ExtractAndVerifyEncryptedPackage(packagePath, extracted, passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles, cfg.checkWorkspaceCapacity)
	require.NoError(t, err)
	require.NoError(t, cfg.restoreIsolatedDatabase(ctx, extracted, stageName, manifest))
	stage, err := cfg.connectDatabase(ctx, stageName)
	require.NoError(t, err)
	defer stage.Close()
	var restoredViewReceiptCount int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT count(*) FROM product_view_count_flush_receipts WHERE batch_id='22222222-2222-4222-8222-222222222222' AND product_id=999999 AND delta=9 AND NOT product_increment_applied`).Scan(&restoredViewReceiptCount))
	require.Equal(t, 1, restoredViewReceiptCount, "whole-storefront encrypted recovery must retain no-op replay receipts even for absent products")
	require.NoError(t, findBusinessRelationshipGaps(ctx, stage, &manifest))
	// Exercise the row queries on actual migrated columns, including retained
	// conversations, ownership constraints, upload receipts and derivatives.
	_, err = stage.ExecContext(ctx, `DO $$ DECLARE synthetic_user bigint; synthetic_ticket bigint; synthetic_message bigint; synthetic_asset bigint; BEGIN
INSERT INTO users(email,username,password) VALUES('synthetic-row-check@example.invalid','synthetic-row-check','synthetic-unused-password') RETURNING id INTO synthetic_user;
INSERT INTO tickets(ticket_number,user_id,customer_user_id,conversation_id,visitor_session_hash,subject,category,deleted_at) VALUES('synthetic-row-check',synthetic_user,synthetic_user,'synthetic-row-check',repeat('b',64),'synthetic retained conversation','customer_service',now()) RETURNING id INTO synthetic_ticket;
INSERT INTO ticket_messages(ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(synthetic_ticket,synthetic_user,'synthetic private retained body','[]','media_asset_bindings') RETURNING id INTO synthetic_message;
INSERT INTO visitor_profiles(customer_service_visitor_hash,user_id) VALUES(repeat('b',64),synthetic_user);
INSERT INTO media_assets(filename,url,storage_key,size,visibility,library_partition,upload_origin,classification_method,customer_service_media_purpose,uploader_id,uploader_actor_type) VALUES('synthetic-row-check.txt','/uploads-private/customer-service/attachments/originals/synthetic-row-check.txt','customer-service/attachments/originals/synthetic-row-check.txt',8,'private','customer_service','customer_service_conversation_upload','recorded_upload_context','conversation_attachment',synthetic_user,'customer') RETURNING id INTO synthetic_asset;
INSERT INTO media_asset_derivatives(media_asset_id,preset,url,storage_key,mime_type,size) VALUES(synthetic_asset,'synthetic','/uploads-private/customer-service/attachments/derivatives/synthetic-row-check.txt','customer-service/attachments/derivatives/synthetic-row-check.txt','text/plain',8);
INSERT INTO media_upload_sessions(id,business_context_type,business_context_id,upload_actor_type,upload_actor_account_id,upload_actor_visitor_session_hash,object_key,media_asset_id,status,expires_at) VALUES('synthetic-full-schema-upload','customer_service',synthetic_ticket,'customer',synthetic_user,'','customer-service/attachments/originals/synthetic-row-check.txt',synthetic_asset,'bound',now()+interval '1 day');
INSERT INTO ticket_message_media_attachments(ticket_message_id,media_asset_id,attachment_index,upload_session_id) VALUES(synthetic_message,synthetic_asset,0,'synthetic-full-schema-upload');
INSERT INTO customer_service_inbox_states(ticket_id,recipient_user_id,last_read_message_id) VALUES(synthetic_ticket,synthetic_user,synthetic_message);
INSERT INTO customer_service_agent_profiles(agent_id,user_id,name,avatar,status,deleted_at) VALUES('synthetic-retained-avatar-owner',synthetic_user,'private retained avatar profile','/uploads/customer-service/avatars/private-retained-profile.png','inactive',now());
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status) SELECT 'synthetic-full-schema-avatar-cleanup','customer_service.avatar_cleanup','customer_service_agent_profile',id::text,jsonb_build_object('url',avatar),'ignored' FROM customer_service_agent_profiles WHERE agent_id='synthetic-retained-avatar-owner';
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('synthetic-full-schema-realtime-event','customer_service.realtime','customer_service_conversation',synthetic_ticket::text,jsonb_build_object('event_id','synthetic-full-schema-realtime-event','type','conversation.message.created','ticket_id',synthetic_ticket,'conversation_id','synthetic-row-check','payload',jsonb_build_object('message_id',synthetic_message)),'processed',now());
END $$`)
	require.NoError(t, err)
	_, err = stage.ExecContext(ctx, `DO $$ DECLARE synthetic_user bigint; synthetic_product bigint; synthetic_variant bigint; synthetic_order bigint; synthetic_transaction bigint; BEGIN
SELECT id INTO synthetic_user FROM users WHERE username='synthetic-row-check';
INSERT INTO products(name,slug,price_minor,currency) VALUES('synthetic historical product','synthetic-order-row-product',100,'USD') RETURNING id INTO synthetic_product;
INSERT INTO product_variants(product_id,sku,price_minor,currency,stock) VALUES(synthetic_product,'synthetic-order-row-variant',100,'USD',20) RETURNING id INTO synthetic_variant;
INSERT INTO orders(order_number,user_id,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor,deleted_at) VALUES('synthetic-full-schema-order',synthetic_user,'USD','USD',300,300,300,0,0,0,0,now()) RETURNING id INTO synthetic_order;
INSERT INTO order_items(order_id,product_id,variant_id,product_name,quantity,currency,price_minor,subtotal_minor,total_minor,tax_amount_minor,discount_minor) VALUES(synthetic_order,synthetic_product,synthetic_variant,'synthetic retained order line',3,'USD',100,300,300,0,0);
INSERT INTO transactions(order_id,payment_method,amount_minor,currency,status) VALUES(synthetic_order,'synthetic-provider',300,'USD','completed') RETURNING id INTO synthetic_transaction;
INSERT INTO refunds(order_id,transaction_id,amount_minor,requested_amount_minor,discount_clawback_amount_minor,currency,status) VALUES(synthetic_order,synthetic_transaction,100,100,0,'USD','completed');
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('synthetic-full-schema-order-completed','order.completed','order',synthetic_order::text,jsonb_build_object('order_id',synthetic_order,'order_number','synthetic-full-schema-order','schema_version',1,'idempotency_key','synthetic-full-schema-order-completion'),'processed',now());
INSERT INTO email_delivery_records(outbox_event_id,event_key,event_type,template_code,recipient_email,reference_type,reference_number,status,first_attempt_at,last_attempt_at,sent_at)
SELECT id,'notification:synthetic-full-schema-order-completion:order_completed','order.completed','order_completed','synthetic-row-check@example.invalid','order','synthetic-full-schema-order','sent',now(),now(),now() FROM outbox_events WHERE event_key='synthetic-full-schema-order-completed';
INSERT INTO payment_operation_idempotencies(user_id,scope,idempotency_key,request_hash,status,status_code,response_body) VALUES(synthetic_user,'synthetic-capture','synthetic-global-payment-key',repeat('a',64),'completed',200,'synthetic private payment response');
INSERT INTO refund_idempotencies(admin_user_id,scope,idempotency_key,request_hash,refund_id) SELECT synthetic_user,'synthetic-refund','synthetic-global-refund-key',repeat('b',64),id FROM refunds WHERE order_id=synthetic_order;
END $$`)
	require.NoError(t, err)
	for _, key := range []string{"customer-service/attachments/originals/synthetic-row-check.txt", "customer-service/attachments/derivatives/synthetic-row-check.txt"} {
		packagePath := "media/uploads_private/" + key
		for _, filename := range []string{filepath.Join(extracted, filepath.FromSlash(packagePath)), filepath.Join(cfg.StorageRoots["uploads_private"], filepath.FromSlash(key))} {
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
		}
		identity, err := fileIdentity(filepath.Join(extracted, filepath.FromSlash(packagePath)), packagePath)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, identity)
	}
	observationTransaction, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, err)
	defer observationTransaction.Rollback()
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, observationTransaction)
	require.NoError(t, err)
	require.Equal(t, len(observation.Tables), len(registeredStorefrontModuleRecoveryTableNames()), "registry must not retain removed migration tables")
	var conversationSourceGraph, conversationTargetGraph *moduleRecoveryRowDependencyGraph
	conversationRows, err := cfg.inspectCustomerServiceConversationRows(ctx, observationTransaction, observationTransaction, observation, observation, &manifest, extracted, defaultModuleRecoveryRowInspectionLimits(), func(source, target *moduleRecoveryRowDependencyGraph) {
		conversationSourceGraph, conversationTargetGraph = source, target
	})
	require.NoError(t, err)
	require.True(t, conversationRows.Complete)
	require.Zero(t, conversationRows.BlockingIssues, "%+v", conversationRows.IssueCounts)
	require.Len(t, conversationRows.ConversationReferenceChecks.ColumnContracts, 2)
	for _, coverage := range conversationRows.ConversationReferenceChecks.ColumnContracts {
		require.True(t, coverage.Complete, "%+v", coverage)
		require.Zero(t, coverage.UnregisteredColumns)
		t.Logf("complete migration candidate column contracts: side=%s observed=%d registered=%d foreign_key=%d digest=%s", coverage.Side, coverage.ObservedColumns, coverage.RegisteredColumns, coverage.ForeignKeyColumns, coverage.ColumnContractsSHA256)
	}
	require.Len(t, conversationRows.ConversationReferenceChecks.RetainedMediaOwnerColumnContracts, 2)
	for _, coverage := range conversationRows.ConversationReferenceChecks.RetainedMediaOwnerColumnContracts {
		require.True(t, coverage.Complete, "%+v", coverage)
		require.Zero(t, coverage.UnregisteredColumns)
		t.Logf("complete migration retained avatar/auto-reply column contracts: side=%s observed=%d registered=%d digest=%s", coverage.Side, coverage.ObservedColumns, coverage.RegisteredColumns, coverage.ColumnContractsSHA256)
	}
	require.Equal(t, 2, conversationRows.Files.PackageVerifiedFiles)
	require.NotNil(t, conversationRows.ConversationRecoveryPlan)
	require.Len(t, conversationRows.ConversationRecoveryPlan.TargetPreservation, len(observation.Tables))
	require.False(t, conversationRows.ConversationRecoveryPlan.ExecutionAvailable)
	require.Empty(t, conversationRows.ConversationRecoveryPlan.BlockingIssueCounts)
	require.Zero(t, conversationRows.ConversationRecoveryPlan.RowActionCounts["propose_replace_after_review"])
	require.NotNil(t, conversationRows.DurableEvents)
	require.True(t, conversationRows.DurableEvents.Complete)
	require.Equal(t, 2, conversationRows.DurableEvents.SourceRelatedEvents)
	require.Equal(t, 2, conversationRows.DurableEvents.TargetRelatedEvents)
	require.Equal(t, 2, conversationRows.DurableEvents.ComparedEventIdentities)
	require.Zero(t, conversationRows.IssueCounts["avatar_cleanup_owner_schema_missing"])
	require.Positive(t, conversationRows.IssueCounts["suppressed_cleanup_state_must_be_preserved"])
	require.Equal(t, "same_record_preserved", conversationRows.DurableEvents.Comparisons[0].Result)
	t.Logf("full schema conversation inventory: %d selected rows, %d verified media files", conversationRows.SourceSelectedRows, conversationRows.Files.PackageVerifiedFiles)
	selectionPlan := conversationRows.ConversationRecoveryPlan
	selectionPlan.Binding = CustomerServiceRecoveryPlanSnapshotBinding{TaskID: newTaskID(), SiteID: cfg.SiteID, Environment: cfg.Environment, PackageID: newTaskID(), WorkCopyExpiresAt: time.Now().UTC().Add(time.Hour)}
	selectionPlan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(selectionPlan)
	require.NoError(t, err)
	effectiveSelection, err := cfg.inspectCustomerServiceEffectiveSelectedRows(ctx, conversationSourceGraph, conversationTargetGraph, selectionPlan, nil, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: selectionPlan.SHA256}, &manifest, extracted)
	require.NoError(t, err)
	require.True(t, effectiveSelection.SelectionDependenciesSatisfied, "%+v", effectiveSelection.DependencyChecks.IssueCounts)
	require.Len(t, effectiveSelection.TargetPreservation, len(observation.Tables))
	t.Logf("full schema effective selection: %d retained tables, %d required files; candidate construction remains disabled", len(effectiveSelection.TargetPreservation), len(effectiveSelection.RequiredFiles))
	verifyCustomerServiceDisposableCandidateRehearsalWithCompleteMigrationFixture(t, cfg, stageName, manifest.Files, extracted)
	orderRows, err := cfg.inspectOrderTransactionRows(ctx, observationTransaction, observationTransaction, observation, observation, &manifest, extracted, defaultModuleRecoveryRowInspectionLimits())
	require.NoError(t, err)
	require.True(t, orderRows.Complete, "%+v", orderRows.IssueCounts)
	require.Zero(t, orderRows.BlockingIssues, "%+v", orderRows.IssueCounts)
	require.GreaterOrEqual(t, orderRows.SourceSelectedRows, 7)
	require.NotNil(t, orderRows.DurableEvents)
	require.True(t, orderRows.DurableEvents.Complete)
	require.Equal(t, 1, orderRows.DurableEvents.SourceRelatedEvents)
	require.Equal(t, 1, orderRows.DurableEvents.ComparedEventIdentities)
	require.NotNil(t, orderRows.SideEffectReceipts)
	require.True(t, orderRows.SideEffectReceipts.Complete)
	require.Equal(t, 3, orderRows.SideEffectReceipts.ComparedIdentities)
	for _, table := range orderRows.SideEffectReceipts.Tables {
		require.Equal(t, 1, table.Matching, table.Table)
	}
	t.Logf("full schema order inventory: %d selected rows; shared customer, product and SKU stock preserved", orderRows.SourceSelectedRows)
	for module, contract := range registeredOrderAndCustomerServiceRecoveryDependencyContracts() {
		t.Run("complete_schema_module_dependency_registry_"+module, func(t *testing.T) {
			report := buildModuleRecoveryDependencyReport(contract, &Task{Manifest: &manifest}, observation, observation)
			sortModuleRecoveryDependencyReport(report)
			for _, blocker := range report.Blockers {
				switch blocker.Code {
				case "unregistered_table", "source_dependency_table_missing", "referenced_table_missing", "registered_reference_column_missing":
					t.Errorf("%s: %s: %s", blocker.Code, blocker.Table, blocker.Message)
				}
			}
			require.False(t, report.RestoreAvailable)
			t.Logf("registered %d actual tables and %d module dependency references", len(observation.Tables), len(report.References))
		})
	}
	t.Logf("verified %d application tables at migration %d; %d seeded media/reference gaps remain explicit", len(manifest.Tables), manifest.MigrationVersion, len(manifest.Gaps))
	_, _, err = cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	verifyCustomerServiceRetainedAvatarCleanupAgainstActualMigrationColumns(t, cfg, stage, &manifest, extracted)
	verifyCustomerServiceRetiredDerivativeCleanupIdentitiesAgainstActualMigrationSchema(t, cfg, stage, &manifest, extracted)
	verifyCustomerServicePrivateUploadActorReadabilityAgainstActualMigrationSchema(t, cfg, stage, &manifest, extracted)
	verifyCompleteMigrationCanonicalArticleCoverMediaGapsWithoutSchemaChanges(t, cfg, stage)
}
