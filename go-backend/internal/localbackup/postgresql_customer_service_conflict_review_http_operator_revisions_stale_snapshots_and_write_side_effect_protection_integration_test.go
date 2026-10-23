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

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLCustomerServiceConflictReviewRecordsOperatorChoicesWithoutWaivingIdentityOrWritingBusinessData(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires a new disposable PostgreSQL server")
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
	database, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	defer database.Close()
	_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema+customerServiceRowInspectionSyntheticRows+`
DELETE FROM ticket_message_media_attachments; DELETE FROM media_upload_sessions; DELETE FROM media_asset_derivatives; DELETE FROM media_assets;
UPDATE ticket_messages SET attachments='[]'; CREATE SEQUENCE synthetic_review_freshness_sequence;`)
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
	backup, err := manager.CreateBackup("synthetic-operator", "BACKUP "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	filename, err := manager.taskPackagePath(backup.ID)
	require.NoError(t, err)
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	verified := uploadEncryptedRecoveryFixture(t, manager, data)
	_, err = database.ExecContext(ctx, `UPDATE ticket_messages SET content='private current content to preserve' WHERE id=1; DELETE FROM ticket_messages WHERE id=4; INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments) VALUES(5,1,1,'private target-only message','[]')`)
	require.NoError(t, err)
	report, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	plan := report.ConversationRows.ConversationRecoveryPlan
	require.True(t, report.ConversationRows.Complete, "%+v", report.ConversationRows.IssueCounts)
	require.Equal(t, 2, plan.ConflictDecisionPolicy.RequiredRowDecisions)
	require.True(t, plan.ConflictDecisionPolicy.MappingComplete)
	require.Empty(t, plan.ConflictDecisionPolicy.RepairRequiredIssueCounts)
	require.Equal(t, 1, plan.ConflictDecisionPolicy.FixedPreservationIssueCounts["target_only_conversation_row"])
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	firstToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "first-review-operator", []string{PermissionRestore})
	require.NoError(t, err)
	secondToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "second-review-operator", []string{PermissionRestore})
	require.NoError(t, err)
	path := ConsolePrefix + "/api/tasks/" + verified.ID + "/customer-service-conflict-review"
	callReviewHTTP := func(method, token string, body any) *httptest.ResponseRecorder {
		var encoded []byte
		if body != nil {
			var err error
			encoded, err = json.Marshal(body)
			require.NoError(t, err)
		}
		request := httptest.NewRequest(method, path, strings.NewReader(string(encoded)))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		return response
	}
	callEffectiveSelectionHTTP := func(token string, body any) *httptest.ResponseRecorder {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodPost, ConsolePrefix+"/api/tasks/"+verified.ID+"/customer-service-effective-selection-inspection", strings.NewReader(string(encoded)))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		return response
	}
	decodeReview := func(response *httptest.ResponseRecorder) *CustomerServiceRecoveryConflictReview {
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var value struct {
			Review *CustomerServiceRecoveryConflictReview `json:"review"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &value))
		return value.Review
	}
	var replaceRow, insertRow CustomerServiceRecoveryPlanRowAction
	for _, row := range plan.RowActions {
		if row.Action == "propose_replace_after_review" {
			replaceRow = row
		}
		if row.Action == "propose_insert_after_review" {
			insertRow = row
		}
	}
	require.NotEmpty(t, replaceRow.IdentitySHA256)
	require.NotEmpty(t, insertRow.IdentitySHA256)
	request := CustomerServiceRecoveryConflictReviewRequest{PlanSHA256: plan.SHA256, Confirmation: "REVIEW " + cfg.SiteID + "/" + cfg.Environment + "/" + verified.Manifest.PackageID, Decisions: []CustomerServiceRecoveryContentDecision{{Table: replaceRow.Table, IdentitySHA256: replaceRow.IdentitySHA256, Choice: "preserve_current_row", Reason: "checked the current message"}}}
	before, _, err := cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	t.Run("read_and_save_require_restore_permission", func(t *testing.T) {
		for _, permission := range []string{PermissionView, PermissionExport, PermissionUpload} {
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "unauthorized-review-operator", []string{permission})
			require.NoError(t, err)
			for _, method := range []string{http.MethodGet, http.MethodPost} {
				require.Equal(t, http.StatusForbidden, callReviewHTTP(method, token, request).Code)
			}
			require.Equal(t, http.StatusForbidden, callEffectiveSelectionHTTP(token, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256}).Code)
		}
	})
	t.Run("unknown_override_fields_cannot_be_smuggled", func(t *testing.T) {
		require.Equal(t, http.StatusConflict, callReviewHTTP(http.MethodPost, firstToken, map[string]any{"ignore_all_blockers": true, "plan_sha256": plan.SHA256}).Code)
	})
	require.Nil(t, decodeReview(callReviewHTTP(http.MethodGet, firstToken, nil)))
	first := decodeReview(callReviewHTTP(http.MethodPost, firstToken, request))
	require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(firstToken, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256, ReviewSHA256: first.SHA256}).Code)
	require.Equal(t, 1, first.PendingRowDecisions)
	require.Equal(t, "first-review-operator", first.Actor)
	require.False(t, first.ExecutionAvailable)
	directory, _ := manager.taskDirectory(verified.ID)
	firstRevision, err := os.ReadFile(filepath.Join(directory, "customer-service-conflict-review-revision-"+first.SHA256+".json"))
	require.NoError(t, err)
	manager.Close()
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	require.Equal(t, first.SHA256, decodeReview(callReviewHTTP(http.MethodGet, secondToken, nil)).SHA256)
	t.Run("concurrent_operator_cannot_overwrite_new_revision", func(t *testing.T) {
		require.Equal(t, http.StatusConflict, callReviewHTTP(http.MethodPost, secondToken, request).Code)
	})
	request.ExpectedReviewSHA256 = first.SHA256
	request.Decisions = []CustomerServiceRecoveryContentDecision{{Table: insertRow.Table, IdentitySHA256: insertRow.IdentitySHA256, Choice: "use_package_row", Reason: "checked the missing historical message"}}
	second := decodeReview(callReviewHTTP(http.MethodPost, secondToken, request))
	require.True(t, second.ContentDecisionsComplete)
	require.False(t, second.ExecutionAvailable)
	require.Equal(t, first.SHA256, second.PreviousReviewSHA256)
	selectionRequest := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256, ReviewSHA256: second.SHA256}
	selectionResponse := callEffectiveSelectionHTTP(secondToken, selectionRequest)
	require.Equal(t, http.StatusOK, selectionResponse.Code, selectionResponse.Body.String())
	var selection struct {
		Report CustomerServiceEffectiveSelectionInspection `json:"report"`
	}
	require.NoError(t, json.Unmarshal(selectionResponse.Body.Bytes(), &selection))
	require.True(t, selection.Report.SelectionDependenciesSatisfied, "%+v", selection.Report.DependencyChecks.IssueCounts)
	require.Equal(t, second.SHA256, selection.Report.ReviewSHA256)
	require.False(t, selection.Report.ExecutionAvailable)
	require.False(t, selection.Report.CandidateConstructionAvailable)
	require.Equal(t, 1, selection.Report.EffectiveRowActionCounts["selected_preserve_current"])
	require.Equal(t, 1, selection.Report.EffectiveRowActionCounts["selected_package_insert"])
	storedSelection, err := os.ReadFile(filepath.Join(directory, "customer-service-effective-selection-inspection-v1.json"))
	require.NoError(t, err)
	require.NotEmpty(t, storedSelection)
	t.Run("selection_rejects_stale_review_and_unknown_override_fields", func(t *testing.T) {
		require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(firstToken, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256, ReviewSHA256: first.SHA256}).Code)
		require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(firstToken, map[string]any{"plan_sha256": plan.SHA256, "review_sha256": second.SHA256, "execution_available": true}).Code)
	})
	unchangedRevision, err := os.ReadFile(filepath.Join(directory, "customer-service-conflict-review-revision-"+first.SHA256+".json"))
	require.NoError(t, err)
	require.Equal(t, string(firstRevision), string(unchangedRevision))
	after, _, err := cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
	freshTask, err := manager.GetTask(verified.ID)
	require.NoError(t, err)
	require.Equal(t, "verified", freshTask.State)
	require.Nil(t, freshTask.Journal)
	require.Equal(t, verified.PreviewTargetFingerprint, freshTask.PreviewTargetFingerprint)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.False(t, fenced)
	request.ExpectedReviewSHA256 = second.SHA256
	currentReviewPath := filepath.Join(directory, customerServiceConflictReviewCurrentFilename)
	currentReviewBytes, err := os.ReadFile(currentReviewPath)
	require.NoError(t, err)
	t.Run("workspace_capacity_refuses_new_revision_without_losing_current_choices", func(t *testing.T) {
		previousLimit := manager.Configuration.MaxWorkspaceBytes
		manager.Configuration.MaxWorkspaceBytes = 1
		response := callReviewHTTP(http.MethodPost, secondToken, request)
		manager.Configuration.MaxWorkspaceBytes = previousLimit
		require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
		require.Contains(t, response.Body.String(), "capacity")
		stored, err := os.ReadFile(currentReviewPath)
		require.NoError(t, err)
		require.Equal(t, string(currentReviewBytes), string(stored))
	})
	for _, scenario := range []struct{ name, change, undo string }{
		{"target_content_changed", `UPDATE ticket_messages SET content='changed after content review' WHERE id=1`, `UPDATE ticket_messages SET content='private current content to preserve' WHERE id=1`},
		{"target_sequence_changed", `SELECT nextval('synthetic_review_freshness_sequence')`, `SELECT setval('synthetic_review_freshness_sequence',1,false)`},
		{"target_index_changed", `CREATE INDEX synthetic_review_new_index ON tickets(category)`, `DROP INDEX synthetic_review_new_index`},
		{"target_schema_changed", `ALTER TABLE ticket_messages ADD COLUMN synthetic_review_new_field text`, `ALTER TABLE ticket_messages DROP COLUMN synthetic_review_new_field`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := database.ExecContext(ctx, scenario.change)
			require.NoError(t, err)
			response := callReviewHTTP(http.MethodPost, secondToken, request)
			require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
			require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(secondToken, selectionRequest).Code)
			_, err = database.ExecContext(ctx, scenario.undo)
			require.NoError(t, err)
			stored, err := os.ReadFile(currentReviewPath)
			require.NoError(t, err)
			require.Equal(t, string(currentReviewBytes), string(stored))
		})
	}
	t.Run("target_media_changed", func(t *testing.T) {
		filename := filepath.Join(cfg.StorageRoots["uploads"], "synthetic-new-preserved-media.txt")
		require.NoError(t, os.WriteFile(filename, []byte("new current file"), 0600))
		response := callReviewHTTP(http.MethodPost, secondToken, request)
		require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
		require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(secondToken, selectionRequest).Code)
		require.NoError(t, os.Remove(filename))
	})
	for _, scenario := range []struct{ name, create, drop string }{
		{"candidate_trigger", `CREATE FUNCTION synthetic_review_trigger_function() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'private trigger body must never execute'; END $$; CREATE TRIGGER synthetic_review_trigger BEFORE UPDATE ON ticket_messages FOR EACH ROW EXECUTE FUNCTION synthetic_review_trigger_function()`, `DROP TRIGGER synthetic_review_trigger ON ticket_messages; DROP FUNCTION synthetic_review_trigger_function()`},
		{"disabled_candidate_trigger", `CREATE FUNCTION synthetic_review_trigger_function() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER synthetic_review_trigger BEFORE UPDATE ON ticket_messages FOR EACH ROW EXECUTE FUNCTION synthetic_review_trigger_function(); ALTER TABLE ticket_messages DISABLE TRIGGER synthetic_review_trigger`, `DROP TRIGGER synthetic_review_trigger ON ticket_messages; DROP FUNCTION synthetic_review_trigger_function()`},
		{"candidate_rewrite_rule", `CREATE RULE synthetic_review_write_rule AS ON UPDATE TO ticket_messages DO ALSO INSERT INTO customer_service_retention_policies VALUES(999,123)`, `DROP RULE synthetic_review_write_rule ON ticket_messages`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := database.ExecContext(ctx, scenario.create)
			require.NoError(t, err)
			report, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
			require.NoError(t, err)
			require.False(t, report.ConversationRows.Complete)
			require.Positive(t, report.ConversationRows.IssueCounts["conversation_write_side_effect_contract_unsupported"])
			require.NotEqual(t, plan.Binding.TargetWriteSideEffectFingerprint, report.ConversationRows.ConversationRecoveryPlan.Binding.TargetWriteSideEffectFingerprint)
			encoded, err := json.Marshal(report)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private trigger body")
			response := callReviewHTTP(http.MethodPost, secondToken, request)
			require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
			_, err = database.ExecContext(ctx, scenario.drop)
			require.NoError(t, err)
		})
	}
	t.Run("trigger_function_body_change_invalidates_plan_without_row_changes", func(t *testing.T) {
		_, err := database.ExecContext(ctx, `CREATE FUNCTION synthetic_review_trigger_function() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER synthetic_review_trigger BEFORE UPDATE ON ticket_messages FOR EACH ROW EXECUTE FUNCTION synthetic_review_trigger_function()`)
		require.NoError(t, err)
		firstReport, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
		require.NoError(t, err)
		_, err = database.ExecContext(ctx, `CREATE OR REPLACE FUNCTION synthetic_review_trigger_function() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN OLD; END $$`)
		require.NoError(t, err)
		secondReport, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
		require.NoError(t, err)
		require.Equal(t, firstReport.TargetDatabaseFingerprint, secondReport.TargetDatabaseFingerprint)
		require.NotEqual(t, firstReport.ConversationRows.ConversationRecoveryPlan.SHA256, secondReport.ConversationRows.ConversationRecoveryPlan.SHA256)
		_, err = database.ExecContext(ctx, `DROP TRIGGER synthetic_review_trigger ON ticket_messages; DROP FUNCTION synthetic_review_trigger_function()`)
		require.NoError(t, err)
	})
	t.Run("issue_samples_do_not_limit_decision_mapping", func(t *testing.T) {
		_, err := database.ExecContext(ctx, `INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments) SELECT number,1,1,'private target-only bounded sample','[]' FROM generate_series(100,320) number`)
		require.NoError(t, err)
		report, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
		require.NoError(t, err)
		require.True(t, report.ConversationRows.IssuesTruncated)
		require.True(t, report.ConversationRows.ConversationRecoveryPlan.ConflictDecisionPolicy.MappingComplete)
		require.Equal(t, 222, report.ConversationRows.ConversationRecoveryPlan.ConflictDecisionPolicy.FixedPreservationIssueCounts["target_only_conversation_row"])
		_, err = database.ExecContext(ctx, `DELETE FROM ticket_messages WHERE id>=100`)
		require.NoError(t, err)
	})
	finalReport, err := manager.InspectVerifiedPackageModuleDependencies(ctx, verified.ID, "customer_service")
	require.NoError(t, err)
	if outputDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); outputDirectory != "" {
		require.NoError(t, writeDurableJSON(filepath.Join(outputDirectory, "synthetic-customer-service-effective-selection-http-report.json"), selection.Report))
		require.NoError(t, writeDurableJSON(filepath.Join(outputDirectory, "synthetic-customer-service-conflict-decision-policy-http-report.json"), finalReport))
		require.NoError(t, writeDurableJSON(filepath.Join(outputDirectory, "synthetic-customer-service-snapshot-bound-content-conflict-review.json"), second))
	}
	t.Run("corrupt_current_review_fails_closed", func(t *testing.T) {
		require.NoError(t, os.WriteFile(currentReviewPath, []byte(`{"version":1,"execution_available":true}`), 0600))
		require.Equal(t, http.StatusConflict, callReviewHTTP(http.MethodGet, firstToken, nil).Code)
		require.Equal(t, http.StatusConflict, callReviewHTTP(http.MethodPost, firstToken, request).Code)
		require.Equal(t, http.StatusConflict, callEffectiveSelectionHTTP(firstToken, selectionRequest).Code)
		require.NoError(t, os.WriteFile(currentReviewPath, currentReviewBytes, 0600))
	})
}
