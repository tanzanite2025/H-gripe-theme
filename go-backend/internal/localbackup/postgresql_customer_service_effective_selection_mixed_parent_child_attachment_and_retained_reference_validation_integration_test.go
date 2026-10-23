package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLCustomerServiceEffectiveSelectionValidatesMixedRowsFilesAndRecomputedPreservationWithoutWrites(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires a new disposable PostgreSQL server")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	var databases []*sql.DB
	for range 2 {
		name := "local_backup_test_" + newTaskID()[:12]
		_, err := control.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(name))
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = control.ExecContext(ctx, "DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(name)+" WITH (FORCE)")
		})
		database, err := cfg.connectDatabase(ctx, name)
		require.NoError(t, err)
		t.Cleanup(func() { database.Close() })
		databases = append(databases, database)
		_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema)
		require.NoError(t, err)
	}
	verifiedDirectory := filepath.Join(t.TempDir(), "verified")
	paths := []string{"media/uploads_private/customer-service/attachments/originals/synthetic.txt", "media/uploads_private/customer-service/attachments/derivatives/synthetic.txt", "media/uploads/legacy-attachment.txt"}
	const removeBindings = `DELETE FROM ticket_message_media_attachments; DELETE FROM media_upload_sessions; UPDATE ticket_messages SET attachment_reference_format='legacy_urls' WHERE id=1;`
	for _, scenario := range []struct {
		name, bothSQL, sourceSQL, targetSQL, expectedIssue, missingFile string
		choices                                                         map[string]string
		satisfied                                                       bool
		seedLimit                                                       int
	}{
		{name: "unchanged_selection_needs_no_empty_review", satisfied: true},
		{name: "preserved_current_parent_with_package_message", targetSQL: `UPDATE tickets SET deleted_at='2026-10-02' WHERE id=1; UPDATE ticket_messages SET content='new private current message' WHERE id=1;`, choices: map[string]string{"tickets:1": "preserve_current_row"}, satisfied: true},
		{name: "current_message_format_conflicts_with_package_binding", targetSQL: `DELETE FROM ticket_message_media_attachments; UPDATE ticket_messages SET attachment_reference_format='legacy_urls' WHERE id=1;`, choices: map[string]string{"ticket_messages:1": "preserve_current_row"}, expectedIssue: "attachment_binding_format_conflict"},
		{name: "package_message_and_binding_form_valid_combination", targetSQL: `DELETE FROM ticket_message_media_attachments; UPDATE ticket_messages SET attachment_reference_format='legacy_urls' WHERE id=1;`, satisfied: true},
		{name: "source_only_parent_and_children_resolve_through_overlay", bothSQL: removeBindings, targetSQL: `DELETE FROM customer_service_inbox_states WHERE ticket_id=1; DELETE FROM ticket_messages WHERE ticket_id=1; DELETE FROM tickets WHERE id=1;`, satisfied: true},
		{name: "source_cursor_resolves_source_only_message", bothSQL: removeBindings, targetSQL: `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1; DELETE FROM ticket_messages WHERE id=1;`, satisfied: true},
		{name: "retained_incoming_reference_to_removed_alternate_key_blocks", bothSQL: `CREATE TABLE synthetic_effective_retained_ticket_reference(id bigint PRIMARY KEY,ticket_number text REFERENCES tickets(ticket_number));`, targetSQL: `UPDATE tickets SET ticket_number='private-current-ticket-number' WHERE id=1; INSERT INTO synthetic_effective_retained_ticket_reference VALUES(1,'private-current-ticket-number');`, expectedIssue: "referenced_row_missing"},
		{name: "preserving_parent_keeps_incoming_reference_but_original_hard_issue_survives", bothSQL: `CREATE TABLE synthetic_effective_retained_ticket_reference(id bigint PRIMARY KEY,ticket_number text REFERENCES tickets(ticket_number));`, targetSQL: `UPDATE tickets SET ticket_number='private-current-ticket-number' WHERE id=1; INSERT INTO synthetic_effective_retained_ticket_reference VALUES(1,'private-current-ticket-number');`, choices: map[string]string{"tickets:1": "preserve_current_row"}},
		{name: "natural_key_occupied_by_preserved_other_category_remains_blocked", targetSQL: `UPDATE tickets SET ticket_number='current-other-number' WHERE id=1; INSERT INTO tickets(id,category,ticket_number) VALUES(5,'order','synthetic-ticket-1');`},
		{name: "missing_target_file_has_explicit_package_add_without_waiving_original_issue", missingFile: "media/uploads/legacy-attachment.txt"},
		{name: "effective_union_seed_limit_remains_incomplete", targetSQL: `UPDATE ticket_messages SET content='new private current message' WHERE id=1;`, seedLimit: 1, expectedIssue: "effective_selection_seed_resource_limit"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			for _, database := range databases {
				_, err := database.ExecContext(ctx, `DROP TABLE IF EXISTS synthetic_effective_retained_ticket_reference;`+customerServiceRowInspectionSyntheticRows+scenario.bothSQL)
				require.NoError(t, err)
			}
			manifest := &Manifest{Files: []FileIdentity{}}
			for _, path := range paths {
				parts := strings.SplitN(strings.TrimPrefix(path, "media/"), "/", 2)
				for _, filename := range []string{filepath.Join(verifiedDirectory, filepath.FromSlash(path)), filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))} {
					require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
					require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
				}
				identity, err := fileIdentity(filepath.Join(verifiedDirectory, filepath.FromSlash(path)), path)
				require.NoError(t, err)
				manifest.Files = append(manifest.Files, identity)
			}
			if scenario.missingFile != "" {
				parts := strings.SplitN(strings.TrimPrefix(scenario.missingFile, "media/"), "/", 2)
				require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))))
			}
			for index, statement := range []string{scenario.sourceSQL, scenario.targetSQL} {
				if statement != "" {
					_, err := databases[index].ExecContext(ctx, statement)
					require.NoError(t, err)
				}
			}
			var transactions []*sql.Tx
			var observations []*moduleRecoveryDatabaseObservation
			for _, database := range databases {
				transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
				require.NoError(t, err)
				defer transaction.Rollback()
				observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
				require.NoError(t, err)
				transactions, observations = append(transactions, transaction), append(observations, observation)
			}
			var source, target *moduleRecoveryRowDependencyGraph
			original, err := cfg.inspectCustomerServiceConversationRows(ctx, transactions[0], transactions[1], observations[0], observations[1], manifest, verifiedDirectory, defaultModuleRecoveryRowInspectionLimits(), func(packageGraph, currentGraph *moduleRecoveryRowDependencyGraph) {
				source, target = packageGraph, currentGraph
			})
			require.NoError(t, err)
			plan := original.ConversationRecoveryPlan
			plan.Binding = CustomerServiceRecoveryPlanSnapshotBinding{TaskID: strings.Repeat("1", 32), SiteID: cfg.SiteID, Environment: cfg.Environment, PackageID: strings.Repeat("2", 32), WorkCopyExpiresAt: time.Now().UTC().Add(time.Hour)}
			plan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(plan)
			require.NoError(t, err)
			request := CustomerServiceRecoveryConflictReviewRequest{PlanSHA256: plan.SHA256, Confirmation: "REVIEW " + cfg.SiteID + "/" + cfg.Environment + "/" + plan.Binding.PackageID}
			for _, row := range plan.RowActions {
				if customerServiceRecoveryRowRequiresContentDecision(row) {
					choice := scenario.choices[row.Table+":"+row.RowIdentity]
					if choice == "" {
						choice = "use_package_row"
					}
					request.Decisions = append(request.Decisions, CustomerServiceRecoveryContentDecision{Table: row.Table, IdentitySHA256: row.IdentitySHA256, Choice: choice, Reason: "verified synthetic selection"})
				}
			}
			var review *CustomerServiceRecoveryConflictReview
			selectionRequest := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256}
			if len(request.Decisions) > 0 {
				review, err = buildCustomerServiceRecoveryConflictReview(plan, nil, request, "synthetic-operator", time.Now().UTC())
				require.NoError(t, err)
				selectionRequest.ReviewSHA256 = review.SHA256
			}
			if scenario.seedLimit > 0 {
				target.Limits.Rows = scenario.seedLimit
			}
			result, err := cfg.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, selectionRequest, manifest, verifiedDirectory)
			require.NoError(t, err)
			require.Equal(t, scenario.satisfied, result.SelectionDependenciesSatisfied, "%+v / original %+v", result.DependencyChecks.IssueCounts, result.RepairRequiredIssueCounts)
			if scenario.expectedIssue != "" {
				require.Positive(t, result.DependencyChecks.IssueCounts[scenario.expectedIssue], "%+v", result.DependencyChecks.IssueCounts)
			}
			require.False(t, result.ExecutionAvailable)
			require.False(t, result.CandidateConstructionAvailable)
			require.Len(t, result.TargetPreservation, len(observations[1].Tables))
			if scenario.choices["tickets:1"] == "preserve_current_row" {
				wholeCurrent, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, transactions[1], "tickets", observations[1].Shapes["tickets"], nil)
				require.NoError(t, err)
				for _, preserved := range result.TargetPreservation {
					if preserved.Table == "tickets" {
						require.Equal(t, wholeCurrent, preserved)
					}
				}
			}
			if scenario.missingFile != "" {
				var additions int
				for _, action := range result.RequiredFiles {
					if action.Action == "selected_add_package_file" {
						additions++
					}
				}
				require.Equal(t, 1, additions)
			}
			encoded, err := json.Marshal(result)
			require.NoError(t, err)
			for _, privateValue := range []string{"private retained synthetic message", "private current message", "private-current-ticket-number", "legacy-attachment.txt", "private composite dependency"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			for index, transaction := range transactions {
				_, after, _, err := captureDatabaseTransactionInventory(ctx, transaction)
				require.NoError(t, err)
				require.Equal(t, observations[index].Fingerprint, after)
			}
		})
	}
}
