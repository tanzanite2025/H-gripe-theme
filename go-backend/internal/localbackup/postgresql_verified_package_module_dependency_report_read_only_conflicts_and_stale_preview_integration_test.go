package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLModuleDependencyReportDetectsConflictsWithoutChangingDataOrRestorePreview(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires an explicitly disposable PostgreSQL 15 server")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
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
		for _, task := range tasks {
			if task.StageDatabase != "" {
				_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(task.StageDatabase)+` WITH (FORCE)`)
			}
		}
		_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	}()
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
CREATE TABLE schema_migrations(version bigint NOT NULL,dirty boolean NOT NULL);INSERT INTO schema_migrations VALUES(434,false);
CREATE TABLE users(id bigserial PRIMARY KEY,name text);INSERT INTO users(name) VALUES('synthetic-customer');
CREATE TABLE orders(id bigserial PRIMARY KEY,user_id bigint REFERENCES users(id),total_minor bigint NOT NULL);INSERT INTO orders(user_id,total_minor) VALUES(1,12500);
ALTER TABLE orders ADD COLUMN order_number text; UPDATE orders SET order_number='synthetic-report-order';
CREATE TABLE tickets(id bigserial PRIMARY KEY,user_id bigint REFERENCES users(id),customer_user_id bigint REFERENCES users(id),visitor_session_hash text,category text DEFAULT 'customer_service');INSERT INTO tickets(user_id,customer_user_id) VALUES(1,1);
CREATE TABLE ticket_messages(id bigserial PRIMARY KEY,ticket_id bigint REFERENCES tickets(id),user_id bigint REFERENCES users(id),content text,attachments text,deleted_at timestamptz);
ALTER TABLE ticket_messages ADD COLUMN message_type text DEFAULT 'order'; ALTER TABLE ticket_messages ADD COLUMN metadata text DEFAULT '{"order_number":"synthetic-report-order"}';
INSERT INTO ticket_messages(ticket_id,user_id,content,attachments,deleted_at) VALUES(1,1,'retained synthetic message','[]',now());
CREATE TABLE media_assets(id bigserial PRIMARY KEY,storage_key text,visibility text,url text);
CREATE TABLE visitor_profiles(id bigserial PRIMARY KEY,user_id bigint,customer_service_visitor_hash text);
CREATE TABLE outbox_events(id bigserial PRIMARY KEY,status text);INSERT INTO outbox_events(status) VALUES('pending');`)
	require.NoError(t, err)
	db.Close()
	backup, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	packagePath, err := manager.taskPackagePath(backup.ID)
	require.NoError(t, err)
	data, err := os.ReadFile(packagePath)
	require.NoError(t, err)
	verified := uploadEncryptedRecoveryFixture(t, manager, data)
	initialPreview := verified.PreviewTargetFingerprint
	before, _, err := cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	report, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	require.True(t, report.TargetMatchesWholeStorefrontPreview)
	require.False(t, report.RestoreAvailable)
	initialPlan := report.ConversationRows.ConversationRecoveryPlan
	require.Positive(t, report.ConversationRows.ConversationReferenceChecks.SnapshotReferenceChecks)
	require.NotNil(t, initialPlan)
	require.Len(t, initialPlan.SHA256, 64)
	require.NoError(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(initialPlan, initialPlan.Binding, time.Now()))
	repeatedReport, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	require.Equal(t, initialPlan.SHA256, repeatedReport.ConversationRows.ConversationRecoveryPlan.SHA256, "same observed state produces the same draft digest")
	for _, mutation := range []struct {
		name   string
		change func(*CustomerServiceConversationRecoveryPlan)
	}{
		{"row_action", func(plan *CustomerServiceConversationRecoveryPlan) { plan.Scope = "tampered" }},
		{"execution_flag", func(plan *CustomerServiceConversationRecoveryPlan) { plan.ExecutionAvailable = true }},
		{"unknown_version", func(plan *CustomerServiceConversationRecoveryPlan) { plan.Version = 99 }},
	} {
		t.Run("plan_rejects_"+mutation.name, func(t *testing.T) {
			copyOfPlan := *initialPlan
			mutation.change(&copyOfPlan)
			require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(&copyOfPlan, initialPlan.Binding, time.Now()))
		})
	}
	require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(initialPlan, initialPlan.Binding, initialPlan.Binding.WorkCopyExpiresAt))
	for _, mutation := range []struct {
		name   string
		change func(*CustomerServiceRecoveryPlanSnapshotBinding)
	}{
		{"sequence", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.TargetDatabaseFingerprint = "changed sequence"
		}},
		{"media", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.TargetMediaFingerprint = "changed media"
		}},
		{"source", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.SourceDatabaseFingerprint = "changed source"
		}},
		{"column_constraint", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.TargetColumnConstraintFingerprint = "changed constraint"
		}},
		{"index_definition", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.TargetIndexFingerprint = "changed index"
		}},
		{"storage_boundary", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) {
			binding.TargetBoundarySHA256 = "changed mapping"
		}},
		{"manifest", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) { binding.ManifestSHA256 = "changed manifest" }},
		{"task", func(binding *CustomerServiceRecoveryPlanSnapshotBinding) { binding.TaskID = "another task" }},
	} {
		t.Run("plan_rejects_changed_"+mutation.name, func(t *testing.T) {
			current := initialPlan.Binding
			mutation.change(&current)
			require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(initialPlan, current, time.Now()))
		})
	}
	orderAuthenticationKey, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	orderRequest := httptest.NewRequest(http.MethodPost, ConsolePrefix+"/api/tasks/"+verified.ID+"/module-dependency-report", strings.NewReader(`{"module":"orders"}`))
	orderRequest.Header.Set("Authorization", "Bearer "+orderAuthenticationKey)
	orderResponse := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(orderResponse, orderRequest)
	require.Equal(t, http.StatusOK, orderResponse.Code, orderResponse.Body.String())
	var orderResult struct {
		Report ModuleRecoveryDependencyReport `json:"report"`
	}
	require.NoError(t, json.Unmarshal(orderResponse.Body.Bytes(), &orderResult))
	orderReport := &orderResult.Report
	require.NotNil(t, orderReport.OrderRows)
	require.NotNil(t, orderReport.OrderRows.DurableEvents)
	require.NotNil(t, orderReport.OrderRows.SideEffectReceipts)
	require.False(t, orderReport.OrderRows.SideEffectReceipts.Complete)
	require.Equal(t, "preserve_all_target_receipts_no_automatic_import_or_replay", orderReport.OrderRows.SideEffectReceipts.StatePolicy)
	require.False(t, orderReport.OrderRows.DurableEvents.Complete)
	require.False(t, orderReport.OrderRows.Complete, "minimal test schema must not authorize complete order inspection")
	require.Nil(t, orderReport.ConversationRows)
	require.False(t, orderReport.RestoreAvailable)
	if validationDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); validationDirectory != "" {
		require.NoError(t, writeDurableJSON(filepath.Join(validationDirectory, "synthetic-order-module-dependency-report-from-postgresql-http-integration.json"), orderReport))
	}
	after, _, err := cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after, "all table data, sequences and media remain unchanged")
	for _, table := range report.Tables {
		if table.Table == "ticket_messages" {
			require.EqualValues(t, 1, *table.SourceRecords, "retained soft-deleted messages remain in scope")
			require.Equal(t, "table_contents_match", table.Status)
		}
	}
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `SELECT nextval('orders_id_seq')`)
	require.NoError(t, err)
	db.Close()
	report, err = manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	require.False(t, report.TargetMatchesWholeStorefrontPreview, "a sequence-only change must not be lost")
	require.NotEqual(t, initialPlan.SHA256, report.ConversationRows.ConversationRecoveryPlan.SHA256)
	require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(initialPlan, report.ConversationRows.ConversationRecoveryPlan.Binding, time.Now()))
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE users SET name='newer-synthetic-customer' WHERE id=1;CREATE TABLE orders_new_unregistered_dependency(id bigint PRIMARY KEY)`)
	require.NoError(t, err)
	db.Close()
	before, _, err = cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, ConsolePrefix+"/api/tasks/"+verified.ID+"/module-dependency-report", strings.NewReader(`{"module":"customer_service"}`))
	request.Header.Set("Authorization", "Bearer "+key)
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct {
		Report ModuleRecoveryDependencyReport `json:"report"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	report = &result.Report
	require.NotNil(t, report.ConversationRows)
	require.NotNil(t, report.ConversationRows.DurableEvents)
	require.False(t, report.ConversationRows.DurableEvents.Complete, "the deliberately minimal outbox schema must not appear fully inspected")
	require.Equal(t, "preserve_existing_event_state_no_replay", report.ConversationRows.DurableEvents.StatePolicy)
	require.Positive(t, report.ConversationRows.IssueCounts["durable_event_inspection_schema_missing"])
	require.Positive(t, report.ConversationRows.IssueCounts["shared_dependency_row_differs"], "the HTTP report includes the referenced customer row conflict")
	codes := collectModuleDependencyReportBlockerCodes(report)
	require.Contains(t, codes, "shared_dependency_content_differs:users")
	require.Contains(t, codes, "unregistered_table:orders_new_unregistered_dependency")
	require.Contains(t, codes, "target_changed_since_whole_storefront_preview:")
	require.False(t, report.RestoreAvailable)
	after, _, err = cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
	freshTask, err := manager.GetTask(verified.ID)
	require.NoError(t, err)
	require.Equal(t, initialPreview, freshTask.PreviewTargetFingerprint)
	require.Equal(t, "verified", freshTask.State)
	require.Nil(t, freshTask.Journal)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.False(t, fenced)
	directory, _ := manager.taskDirectory(verified.ID)
	persisted, err := os.ReadFile(filepath.Join(directory, "module-dependency-report-customer_service.json"))
	require.NoError(t, err)
	require.NotContains(t, string(persisted), "newer-synthetic-customer", "reports must not copy customer field values")
	require.NotContains(t, string(persisted), "retained synthetic message")
	persistedPlan, err := os.ReadFile(filepath.Join(directory, "customer-service-conversation-recovery-plan-draft-v1.json"))
	require.NoError(t, err)
	var storedPlan CustomerServiceConversationRecoveryPlan
	require.NoError(t, json.Unmarshal(persistedPlan, &storedPlan))
	require.Equal(t, report.ConversationRows.ConversationRecoveryPlan.SHA256, storedPlan.SHA256)
	require.False(t, storedPlan.ExecutionAvailable)
	require.NotEmpty(t, storedPlan.DependencyBlockers)
	require.NoError(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(&storedPlan, storedPlan.Binding, time.Now()))
	require.NotContains(t, string(persistedPlan), "newer-synthetic-customer")
	if validationDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); validationDirectory != "" {
		require.NoError(t, writeDurableJSON(filepath.Join(validationDirectory, "synthetic-customer-service-module-dependency-conflicts-from-postgresql-integration.json"), report))
	}
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	beforeConstraintChange := report.ConversationRows.ConversationRecoveryPlan
	_, err = db.ExecContext(ctx, `ALTER TABLE users ADD CONSTRAINT synthetic_nonempty_customer_name CHECK(length(name)>0)`)
	require.NoError(t, err)
	db.Close()
	report, err = manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	require.Contains(t, collectModuleDependencyReportBlockerCodes(report), "dependency_schema_differs:users")
	require.NotEqual(t, beforeConstraintChange.Binding.TargetColumnConstraintFingerprint, report.ConversationRows.ConversationRecoveryPlan.Binding.TargetColumnConstraintFingerprint)
	require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(beforeConstraintChange, report.ConversationRows.ConversationRecoveryPlan.Binding, time.Now()))
	beforeIndexChange := report.ConversationRows.ConversationRecoveryPlan
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE UNIQUE INDEX synthetic_partial_ticket_user_unique ON tickets(user_id) WHERE category='customer_service'`)
	require.NoError(t, err)
	db.Close()
	report, err = manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	require.Equal(t, beforeIndexChange.Binding.TargetDatabaseFingerprint, report.ConversationRows.ConversationRecoveryPlan.Binding.TargetDatabaseFingerprint, "adding an index alone must not appear as a row change")
	require.NotEqual(t, beforeIndexChange.Binding.TargetIndexFingerprint, report.ConversationRows.ConversationRecoveryPlan.Binding.TargetIndexFingerprint)
	require.Error(t, verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(beforeIndexChange, report.ConversationRows.ConversationRecoveryPlan.Binding, time.Now()))
	require.Positive(t, report.ConversationRows.IssueCounts["conversation_unique_index_contract_unsupported"])

	_, err = control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	require.NoError(t, err)
	report, err = manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "orders")
	require.NoError(t, err)
	require.True(t, report.TargetDatabaseAbsent)
	require.Contains(t, collectModuleDependencyReportBlockerCodes(report), "target_database_absent:")
	stage, err := cfg.connectDatabase(ctx, verified.StageDatabase)
	require.NoError(t, err)
	_, err = stage.ExecContext(ctx, `SELECT nextval('ticket_messages_id_seq')`)
	require.NoError(t, err)
	stage.Close()
	_, err = manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.ErrorContains(t, err, "isolated database changed")
}
