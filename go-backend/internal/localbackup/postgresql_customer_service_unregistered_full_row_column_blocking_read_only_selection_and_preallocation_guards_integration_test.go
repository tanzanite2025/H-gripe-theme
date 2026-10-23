package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLCustomerServiceUnknownFullRowColumnsBlockEvenMatchingEmptySchemasAndPreserveCurrentSelections(t *testing.T) {
	for _, scenario := range []struct {
		name, table, definition string
		source, target          bool
	}{
		{"both_nullable_ticket_json", "tickets", "jsonb", true, true},
		{"both_empty_message_html", "ticket_messages", "text DEFAULT ''", true, true},
		{"both_empty_inbox_json", "customer_service_inbox_states", "jsonb DEFAULT '{}'", true, true},
		{"both_nullable_attachment_storage_key", "ticket_message_media_attachments", "text", true, true},
		{"source_only_message_json", "ticket_messages", "jsonb", true, false},
		{"target_only_message_html", "ticket_messages", "text DEFAULT ''", false, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				if index == 0 && !scenario.source || index == 1 && !scenario.target {
					continue
				}
				_, err := database.ExecContext(ctx, `ALTER TABLE `+pq.QuoteIdentifier(scenario.table)+` ADD COLUMN custom_unregistered_reference `+scenario.definition)
				require.NoError(t, err)
			}
			_, err := fixture.Target.ExecContext(context.Background(), `UPDATE ticket_messages SET content='private chosen current body' WHERE id=1`)
			require.NoError(t, err)
			before := []string{}
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				before = append(before, fingerprint)
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, map[string]string{"ticket_messages:1": "preserve_current_row"})
			require.False(t, plan.InspectionComplete)
			require.Positive(t, plan.BlockingIssueCounts["conversation_column_reference_contract_unregistered"])
			require.Len(t, plan.ReferenceChecks.ColumnContracts, 2)
			for index, coverage := range plan.ReferenceChecks.ColumnContracts {
				expected := scenario.source
				if index == 1 {
					expected = scenario.target
				}
				require.Equal(t, expected, coverage.UnregisteredColumns == 1)
			}
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.False(t, selection.SelectionDependenciesSatisfied)
			if scenario.source && scenario.target {
				require.Equal(t, 1, selection.EffectiveRowActionCounts["selected_preserve_current"])
			} else {
				require.Positive(t, selection.EffectiveRowActionCounts["blocked_comparison_incomplete"])
			}
			if scenario.target {
				require.Positive(t, selection.DependencyChecks.BlockingIssueCounts["conversation_column_reference_contract_unregistered"])
			}
			if scenario.source {
				_, _, err = readCustomerServiceCandidateSelectedFullRows(ctx, source, selection)
				require.ErrorContains(t, err, "unregistered reference contract")
			}
			if scenario.target {
				require.ErrorContains(t, assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(ctx, target.Transaction, target.Observation), "unregistered reference contract")
			}
			_, err = fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.Error(t, err)
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before[index], fingerprint)
			}
			var candidates int
			require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
			require.Zero(t, candidates)
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private chosen current body")
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, scenario.name+"-blocked-column-contract-selection.json"), selection))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, scenario.name+"-blocked-column-contract-plan.json"), plan))
			}
		})
	}
}

func TestPostgreSQLCustomerServiceUnknownColumnOrPreviousRegistryPlanCannotAllocatePersistentCandidate(t *testing.T) {
	for _, scenario := range []string{"matching_unregistered_columns", "previous_registry_plan"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			cfg := fixture.Configuration
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			if scenario == "matching_unregistered_columns" {
				for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
					_, err := database.ExecContext(context.Background(), `ALTER TABLE ticket_messages ADD COLUMN custom_unregistered_reference jsonb`)
					require.NoError(t, err)
				}
			}
			manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
			if scenario == "previous_registry_plan" {
				source, target, plan, review, _ := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
				currentRequest := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256}
				legacy := *plan
				legacy.Binding.RegistryVersion = 4
				legacy.ReferenceChecks = nil
				legacy.SHA256 = ""
				var err error
				legacy.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(&legacy)
				require.NoError(t, err)
				_, err = validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, nil, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: legacy.SHA256}, time.Now())
				require.ErrorContains(t, err, "plan changed")
				_, err = fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, &legacy, review, CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: legacy.SHA256}, fixture.Manifest, fixture.VerifiedDirectory)
				require.ErrorContains(t, err, "snapshot binding changed")
				_, err = validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, nil, currentRequest, time.Now())
				require.NoError(t, err)
				source.Transaction.Rollback()
				target.Transaction.Rollback()
				request.PlanSHA256 = legacy.SHA256
			}
			before, err := cfg.mediaFingerprint(ctx)
			require.NoError(t, err)
			_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
			require.Error(t, err)
			record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Zero(t, record.Identity.DatabaseOID)
			require.Nil(t, record.MediaUnion)
			var candidates int
			require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
			require.Zero(t, candidates)
			after, err := cfg.mediaFingerprint(ctx)
			require.NoError(t, err)
			require.Equal(t, before, after)
			require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
			require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}
