package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const orderDurableEventCompletedSyntheticRow = `
INSERT INTO outbox_events(id,status,payload,event_key,event_type,aggregate_type,aggregate_id,processed_at)
VALUES(10,'processed','{"order_id":1,"order_number":"private-synthetic-order-one","schema_version":1,"idempotency_key":"private-order-completion-key","recipient_email":"private-event-customer@example.invalid","private_body":"private order notification body"}','private-order-completion-event','order.completed','order','1','2026-10-01');
INSERT INTO loyalty_transactions VALUES(20,'referral_referee',1,1,50,'private registration gift independent from any order');
`

const orderDurableEventRefundCommandSyntheticUpdate = `UPDATE outbox_events SET event_type='payment.refund_execution_requested',aggregate_type='payment',aggregate_id='1',payload='{"refund_id":1,"attempt":1,"provider":"synthetic-provider","idempotency_key":"private-refund-execution-key"}' WHERE id=10;`

func TestPostgreSQLOrderDurableEventsPreserveExecutionStateAndRegistrationRewardsWithoutMutations(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires an explicitly disposable PostgreSQL 15 server")
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
		_, err = database.ExecContext(ctx, orderTransactionRowInspectionSyntheticSchema+"CREATE TABLE shipment_records(id bigserial PRIMARY KEY,order_id bigint);")
		require.NoError(t, err)
	}
	resetSyntheticOrderEventDatabases := func() {
		for _, database := range databases {
			rows, err := database.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename<>'schema_migrations' ORDER BY tablename`)
			require.NoError(t, err)
			var tables []string
			for rows.Next() {
				var table string
				require.NoError(t, rows.Scan(&table))
				tables = append(tables, pq.QuoteIdentifier(table))
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			_, err = database.ExecContext(ctx, "TRUNCATE "+strings.Join(tables, ",")+" RESTART IDENTITY CASCADE;"+orderTransactionRowInspectionSyntheticRows+orderDurableEventCompletedSyntheticRow)
			require.NoError(t, err)
		}
	}
	inspectSyntheticOrderDurableEventSnapshots := func(limits moduleRecoveryRowInspectionLimits) *ModuleRecoveryRowInspection {
		report := &ModuleRecoveryRowInspection{Complete: true, IssueSampleLimit: limits.IssueSamples, IssueCounts: map[string]int{}, Issues: []ModuleRecoveryRowIssue{}}
		var graphs []*moduleRecoveryRowDependencyGraph
		var before []string
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			before = append(before, observation.Fingerprint)
			side := "package"
			if index == 1 {
				side = "target"
			}
			graphs = append(graphs, newModuleRecoveryRowDependencyGraph(transaction, observation, side, report, limits))
		}
		require.NoError(t, cfg.inspectOrderTransactionDurableEventDependencies(ctx, graphs, report))
		require.NotNil(t, report.DurableEvents)
		require.Equal(t, "preserve_existing_event_state_no_replay", report.DurableEvents.StatePolicy)
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, value := range []string{"private-synthetic-order-one", "private-order-completion-event", "private-order-completion-key", "private-refund-execution-key", "private-event-customer@example.invalid", "private order notification body", "private registration gift independent from any order"} {
			require.NotContains(t, string(encoded), value)
		}
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.NoError(t, transaction.Rollback())
			require.Equal(t, before[index], fingerprint, "business facts, registration gifts and sequences must remain unchanged")
			var points int
			require.NoError(t, database.QueryRowContext(ctx, "SELECT points FROM loyalty_transactions WHERE source='referral_referee'").Scan(&points))
			require.Equal(t, 50, points)
		}
		return report
	}
	for _, scenario := range []struct {
		name, statement, code string
		reviewOnly            bool
	}{
		{"matching_completed_event_and_independent_registration_gift", "SELECT 1", "", false},
		{"event_schema_missing_is_incomplete", "ALTER TABLE outbox_events DROP COLUMN available_at", "order_event_inspection_schema_missing", false},
		{"out_of_range_order_id_is_reported_before_sql_cast", `UPDATE outbox_events SET payload=payload||'{"order_id":18446744073709551615}'::jsonb WHERE id=10`, "order_event_order_identity_invalid", false},
		{"processed_state_cannot_return_to_pending", "UPDATE outbox_events SET status='pending',processed_at=NULL WHERE id=10", "durable_event_replay_state_conflict", false},
		{"ignored_state_must_be_preserved", "UPDATE outbox_events SET status='ignored' WHERE id=10", "durable_event_replay_state_conflict", false},
		{"attempt_count_must_be_preserved", "UPDATE outbox_events SET attempts=2 WHERE id=10", "durable_event_replay_state_conflict", false},
		{"retry_schedule_must_be_preserved", "UPDATE outbox_events SET available_at='2026-11-01' WHERE id=10", "durable_event_replay_state_conflict", false},
		{"max_attempts_must_be_preserved", "UPDATE outbox_events SET max_attempts=20 WHERE id=10", "durable_event_replay_state_conflict", false},
		{"execution_lease_requires_reconciliation", "UPDATE outbox_events SET locked_at='2026-10-02',locked_by='private-worker' WHERE id=10", "durable_event_execution_lease_present", false},
		{"unknown_external_result", "UPDATE outbox_events SET status='unknown' WHERE id=10", "durable_event_outcome_unknown", false},
		{"uncertainty_marker_survives_status_change", "UPDATE outbox_events SET uncertain_at='2026-10-02' WHERE id=10", "durable_event_outcome_unknown", false},
		{"consumed_event_must_not_be_reactivated", "UPDATE outbox_events SET deleted_at='2026-10-02' WHERE id=10", "durable_event_replay_state_conflict", false},
		{"target_missing_event_not_reinserted", "DELETE FROM outbox_events WHERE id=10", "source_durable_event_missing_from_target", false},
		{"same_key_moved_to_another_primary_key", "UPDATE outbox_events SET id=11 WHERE id=10", "durable_event_identifier_moved", false},
		{"primary_key_now_belongs_to_unrelated_event", "DELETE FROM outbox_events WHERE id=10;INSERT INTO outbox_events(id,status,payload,event_key,event_type,aggregate_type,aggregate_id) VALUES(10,'processed','{}','private-other-event','merchant.product_upsert','product','2')", "durable_event_primary_key_reassigned", false},
		{"same_key_private_payload_changed", "UPDATE outbox_events SET payload=payload||'{" + `"private_body":"changed private body"` + "}'::jsonb WHERE id=10", "durable_event_key_identity_conflict", false},
		{"payload_order_disagrees_with_aggregate", "UPDATE outbox_events SET payload=payload||'{" + `"order_id":2` + "}'::jsonb WHERE id=10", "order_event_aggregate_identity_conflict", false},
		{"string_order_identifier_is_not_numeric", "UPDATE outbox_events SET payload=payload||'{" + `"order_id":"1"` + "}'::jsonb WHERE id=10", "order_event_order_identity_invalid", false},
		{"order_number_disagrees_with_current_order", "UPDATE outbox_events SET payload=payload||'{" + `"order_number":"private-other-order"` + "}'::jsonb WHERE id=10", "order_event_order_number_identity_conflict", false},
		{"canonical_schema_version_unregistered", "UPDATE outbox_events SET payload=payload||'{" + `"schema_version":2` + "}'::jsonb WHERE id=10", "order_event_canonical_metadata_invalid", false},
		{"canonical_payload_key_missing", "UPDATE outbox_events SET payload=payload-'idempotency_key' WHERE id=10", "order_event_canonical_metadata_invalid", false},
		{"unknown_order_event_type", "UPDATE outbox_events SET event_type='order.future_command' WHERE id=10", "order_event_type_unregistered", false},
		{"non_object_payload_is_incomplete", "UPDATE outbox_events SET payload='[]' WHERE id=10", "order_event_payload_not_an_object", false},
		{"historic_processed_order_missing_is_review_only", "DELETE FROM orders WHERE id=2;UPDATE outbox_events SET aggregate_id='999',payload=payload||'{" + `"order_id":999` + "}'::jsonb WHERE id=10", "order_event_business_reference_unavailable", true},
		{"pending_order_missing_blocks", "UPDATE outbox_events SET status='pending',aggregate_id='999',payload=payload||'{" + `"order_id":999` + "}'::jsonb WHERE id=10", "order_event_business_reference_unavailable", false},
		{"refund_command_preserves_completed_execution", orderDurableEventRefundCommandSyntheticUpdate, "order_refund_completed_execution_preserved", true},
		{"refund_command_wrong_aggregate_type", orderDurableEventRefundCommandSyntheticUpdate + "UPDATE outbox_events SET aggregate_type='order' WHERE id=10", "order_event_aggregate_identity_conflict", false},
		{"refund_command_key_conflict", orderDurableEventRefundCommandSyntheticUpdate + `UPDATE outbox_events SET payload=payload||'{"idempotency_key":"private-wrong-key"}'::jsonb WHERE id=10`, "order_refund_command_idempotency_conflict", false},
		{"refund_command_provider_conflict", orderDurableEventRefundCommandSyntheticUpdate + `UPDATE outbox_events SET payload=payload||'{"provider":"different-provider"}'::jsonb WHERE id=10`, "order_refund_command_idempotency_conflict", false},
		{"old_refund_attempt_preserves_new_execution", orderDurableEventRefundCommandSyntheticUpdate + "UPDATE payment_refund_executions SET attempt_count=2,idempotency_key='private-new-key'", "order_refund_command_attempt_differs", true},
		{"future_refund_attempt_is_conflict", orderDurableEventRefundCommandSyntheticUpdate + `UPDATE outbox_events SET payload=payload||'{"attempt":2}'::jsonb WHERE id=10`, "order_refund_command_attempt_differs", false},
		{"zero_refund_attempt_is_invalid", orderDurableEventRefundCommandSyntheticUpdate + `UPDATE outbox_events SET payload=payload||'{"attempt":0}'::jsonb WHERE id=10`, "order_refund_command_attempt_invalid", false},
		{"pending_refund_needs_external_reconciliation", orderDurableEventRefundCommandSyntheticUpdate + "UPDATE refunds SET status='pending';UPDATE payment_refund_executions SET status='processing'", "order_refund_external_execution_requires_reconciliation", false},
		{"refund_execution_crosses_order", orderDurableEventRefundCommandSyntheticUpdate + "UPDATE payment_refund_executions SET order_id=2", "order_refund_execution_business_identity_conflict", false},
		{"refund_execution_status_unregistered", orderDurableEventRefundCommandSyntheticUpdate + "UPDATE payment_refund_executions SET status='future-status'", "order_refund_execution_status_unregistered", false},
		{"refund_notification_crosses_order", `UPDATE outbox_events SET event_type='payment.refund_completed',aggregate_type='payment',aggregate_id='1',payload='{"order_id":2,"refund_id":1,"transaction_id":2}' WHERE id=10`, "order_event_reference_crosses_orders", false},
		{"shipping_notification_crosses_order", `INSERT INTO shipment_records VALUES(1,2);UPDATE outbox_events SET event_type='order.shipped',payload=payload||'{"shipment_id":1}'::jsonb WHERE id=10`, "order_event_reference_crosses_orders", false},
		{"after_sales_notification_crosses_order", `UPDATE outbox_events SET event_type='after_sales.status_changed',aggregate_type='after_sales_case',aggregate_id='1',payload=payload||'{"case_id":1,"order_id":2}'::jsonb WHERE id=10`, "order_event_reference_crosses_orders", false},
		{"large_arbitrary_email_body_is_hashed_only", `UPDATE outbox_events SET payload=payload||jsonb_build_object('private_body',repeat('private order notification body',100000)) WHERE id=10`, "durable_event_key_identity_conflict", false},
		{"oversized_identity_projection_is_rejected", `UPDATE outbox_events SET payload=payload||jsonb_build_object('order_number',repeat('private order number',5000)) WHERE id=10`, "row_projection_too_large", false},
		{"new_target_event_is_preserved", `INSERT INTO outbox_events(id,status,payload,event_key,event_type,aggregate_type,aggregate_id) SELECT 11,status,payload,'private-target-new-event',event_type,aggregate_type,aggregate_id FROM outbox_events WHERE id=10`, "target_only_durable_event_preserved", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticOrderEventDatabases()
			_, err := databases[1].ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			if scenario.name == "event_schema_missing_is_incomplete" {
				defer func() {
					_, err := databases[1].ExecContext(ctx, "ALTER TABLE outbox_events ADD COLUMN available_at timestamptz DEFAULT '2026-10-01'")
					require.NoError(t, err)
				}()
			}
			report := inspectSyntheticOrderDurableEventSnapshots(defaultModuleRecoveryRowInspectionLimits())
			if scenario.code == "" {
				require.True(t, report.Complete)
				require.Zero(t, report.BlockingIssues)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
				if scenario.reviewOnly {
					for _, issue := range report.Issues {
						if issue.Code == scenario.code {
							require.Equal(t, "review", issue.Severity)
						}
					}
				}
			}
		})
	}
	t.Run("shared_resource_limits_and_comparison_sampling", func(t *testing.T) {
		resetSyntheticOrderEventDatabases()
		_, err := databases[1].ExecContext(ctx, `INSERT INTO outbox_events(id,status,payload,event_key,event_type,aggregate_type,aggregate_id) SELECT 11,status,payload,'private-target-new-event',event_type,aggregate_type,aggregate_id FROM outbox_events WHERE id=10`)
		require.NoError(t, err)
		limits := defaultModuleRecoveryRowInspectionLimits()
		limits.IssueSamples = 1
		report := inspectSyntheticOrderDurableEventSnapshots(limits)
		require.True(t, report.DurableEvents.ComparisonsTruncated)
		require.Equal(t, 2, report.DurableEvents.ComparedEventIdentities)
		limits.Rows = 1
		report = inspectSyntheticOrderDurableEventSnapshots(limits)
		require.False(t, report.Complete)
		require.False(t, report.DurableEvents.Complete)
		limits = defaultModuleRecoveryRowInspectionLimits()
		limits.ProjectionBytes = 200
		report = inspectSyntheticOrderDurableEventSnapshots(limits)
		require.False(t, report.Complete)
	})
	if directory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); directory != "" {
		resetSyntheticOrderEventDatabases()
		_, err := databases[1].ExecContext(ctx, "UPDATE outbox_events SET status='pending',processed_at=NULL WHERE id=10")
		require.NoError(t, err)
		report := inspectSyntheticOrderDurableEventSnapshots(defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, writeDurableJSON(filepath.Join(directory, "synthetic-order-durable-event-replay-state-conflicts-and-registration-reward-preservation.json"), report))
	}
}
