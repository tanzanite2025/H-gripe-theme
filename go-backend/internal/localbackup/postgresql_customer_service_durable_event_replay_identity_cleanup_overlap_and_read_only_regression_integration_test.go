package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const customerServiceDurableEventSyntheticRealtimeRow = `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES(10,'private-synthetic-event-key:10','customer_service.realtime','customer_service_conversation','1','{"type":"conversation.message.created","event_id":"private-synthetic-event-key:10","ticket_id":1,"conversation_id":"synthetic-conversation-one","payload":{"message_id":1,"private_body":"private realtime body never in a recovery report","customer_email":"private-customer@example.invalid"}}','processed','2026-10-01');`

func TestPostgreSQLCustomerServiceDurableEventInspectionProtectsIdempotencyAndCleanupObjectsWithoutMutations(t *testing.T) {
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
		_, err := control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name))
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
		})
		database, err := cfg.connectDatabase(ctx, name)
		require.NoError(t, err)
		t.Cleanup(func() { database.Close() })
		databases = append(databases, database)
		_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema)
		require.NoError(t, err)
	}
	resetSyntheticEventAndConversationDatabases := func() {
		for _, database := range databases {
			_, err := database.ExecContext(ctx, customerServiceRowInspectionSyntheticRows+customerServiceDurableEventSyntheticRealtimeRow)
			require.NoError(t, err)
		}
	}
	inspectSyntheticDurableEventSnapshots := func(limits moduleRecoveryRowInspectionLimits) *ModuleRecoveryRowInspection {
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
			graph := newModuleRecoveryRowDependencyGraph(transaction, observation, side, report, limits)
			require.NoError(t, graph.selectCustomerServiceConversationScopeRows(ctx))
			require.NoError(t, graph.expandTransitiveForeignKeyAndAccountDependencies(ctx))
			require.NoError(t, cfg.includeCustomerServiceLegacyAttachmentAssets(ctx, graph))
			require.NoError(t, graph.expandTransitiveForeignKeyAndAccountDependencies(ctx))
			graphs = append(graphs, graph)
		}
		require.NoError(t, cfg.inspectCustomerServiceDurableEventDependencies(ctx, graphs, report))
		require.NotNil(t, report.DurableEvents)
		require.Equal(t, "preserve_existing_event_state_no_replay", report.DurableEvents.StatePolicy)
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, privateValue := range []string{"private realtime body never in a recovery report", "private-customer@example.invalid", "private-synthetic-event-key:10", "synthetic-conversation-one", "legacy-attachment.txt", "customer-service/avatars/current-private-name.png"} {
			require.NotContains(t, string(encoded), privateValue)
		}
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.NoError(t, transaction.Rollback())
			require.Equal(t, before[index], fingerprint, "data and sequence state must not change")
		}
		return report
	}
	type durableEventRegressionScenario struct {
		name            string
		side            int
		statement, code string
		reviewOnly      bool
	}
	for _, scenario := range []durableEventRegressionScenario{
		{"matching_processed_realtime_event", 1, `SELECT 1`, "", false},
		{"processed_target_must_not_become_pending", 0, `UPDATE outbox_events SET status='pending',processed_at=NULL WHERE id=10`, "durable_event_replay_state_conflict", false},
		{"ignored_target_must_not_be_reenabled", 1, `UPDATE outbox_events SET status='ignored' WHERE id=10`, "durable_event_replay_state_conflict", false},
		{"changed_attempt_counter", 1, `UPDATE outbox_events SET attempts=5 WHERE id=10`, "durable_event_replay_state_conflict", false},
		{"processing_or_reconciliation_lease", 1, `UPDATE outbox_events SET locked_by='synthetic-worker',locked_at=now() WHERE id=10`, "durable_event_execution_lease_present", false},
		{"unknown_external_result", 1, `UPDATE outbox_events SET status='unknown' WHERE id=10`, "durable_event_outcome_unknown", false},
		{"payload_ticket_does_not_match_aggregate", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{ticket_id}','2') WHERE id=10`, "realtime_event_conversation_identity_conflict", false},
		{"quoted_ticket_identifier_is_not_a_numeric_contract", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{ticket_id}','"1"') WHERE id=10`, "realtime_event_conversation_identity_conflict", false},
		{"public_conversation_identity_changed", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{conversation_id}','"other-public-conversation"') WHERE id=10`, "realtime_event_public_conversation_identity_conflict", false},
		{"payload_event_key_conflict", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{event_id}','"other-event-id"') WHERE id=10`, "realtime_event_idempotency_identity_conflict", false},
		{"message_reference_crosses_other_ticket_category", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{payload,message_id}','2') WHERE id=10`, "realtime_event_message_outside_conversation", false},
		{"future_realtime_payload_type", 0, `UPDATE outbox_events SET payload=jsonb_set(payload,'{type}','"future.unsupported"') WHERE id=10`, "realtime_event_payload_type_unregistered", false},
		{"processed_orphan_realtime_is_historical_review", -1, `UPDATE outbox_events SET aggregate_id='99',payload=jsonb_set(jsonb_set(payload,'{ticket_id}','99'),'{type}','"conversation.created"') WHERE id=10`, "realtime_event_conversation_unavailable", true},
		{"target_event_missing_cannot_be_reinserted", 1, `DELETE FROM outbox_events WHERE id=10`, "source_durable_event_missing_from_target", false},
		{"same_key_now_has_different_event_identifier", 1, `UPDATE outbox_events SET id=11 WHERE id=10`, "durable_event_identifier_moved", false},
		{"primary_key_now_owned_by_an_order_event", 1, `UPDATE outbox_events SET event_key='unrelated-new-order-event',event_type='order.paid',aggregate_type='order',aggregate_id='1' WHERE id=10`, "durable_event_primary_key_reassigned", false},
		{"same_key_has_changed_private_payload_content", 1, `UPDATE outbox_events SET payload=jsonb_set(payload,'{payload,private_body}','"changed private body"') WHERE id=10`, "durable_event_key_identity_conflict", false},
		{"target_new_event_preserves_its_state", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) SELECT 11,'synthetic-new-event',event_type,aggregate_type,aggregate_id,jsonb_set(payload,'{event_id}','"synthetic-new-event"'),'processed' FROM outbox_events WHERE id=10`, "target_only_durable_event_preserved", true},
		{"purged_ticket_cleanup_still_hits_backup_file", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-retention-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":["/uploads/legacy-attachment.txt"]}','pending')`, "durable_cleanup_overlaps_retained_recovery_object", false},
		{"valid_purged_ticket_empty_cleanup_is_not_a_missing_dependency", -1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-retention-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":[]}','processed')`, "", false},
		{"retention_payload_wrong_ticket", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-retention-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":98,"attachment_references":[]}','pending')`, "retention_cleanup_conversation_identity_conflict", false},
		{"retention_references_wrong_shape", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-retention-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":{"unexpected":"shape"}}','pending')`, "durable_cleanup_reference_list_invalid", false},
		{"retention_external_object_requires_review", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-retention-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":["https://foreign.invalid/private.png"]}','pending')`, "durable_cleanup_reference_unresolved", false},
		{"processed_original_cleanup_still_participates_in_reconciliation", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-media-cleanup','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["customer-service/attachments/originals/synthetic.txt"]}','processed')`, "durable_cleanup_overlaps_retained_recovery_object", false},
		{"unrelated_order_cleanup_stays_outside_customer_service_report", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-order-cleanup','storage.object_cleanup_requested','order','55','{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["orders/unrelated.pdf"]}','pending')`, "", false},
		{"order_cleanup_key_overlaps_customer_service_shared_file", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-overlap-cleanup','storage.object_cleanup_requested','order','55','{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["legacy-attachment.txt"]}','pending')`, "durable_cleanup_overlaps_retained_recovery_object", false},
		{"normalization_cannot_hide_cleanup_key_overlap", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-overlap-cleanup','storage.object_cleanup_requested','order','55','{"resource_type":"other","resource_id":"55","object_keys":[" /legacy-attachment.txt/ "]}','pending')`, "durable_cleanup_object_key_noncanonical", false},
		{"cleanup_path_escape_is_rejected", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-unsafe-cleanup','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["../other-business-file"]}','pending')`, "durable_cleanup_object_key_unsafe", false},
		{"avatar_cleanup_protects_preserved_current_profile", 1, `INSERT INTO customer_service_agent_profiles VALUES(1,'/uploads/customer-service/avatars/current-private-name.png'); INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-avatar-cleanup','customer_service.avatar_cleanup','customer_service_agent_profile','1','{"url":"/uploads/customer-service/avatars/current-private-name.png"}','processed')`, "durable_cleanup_overlaps_retained_recovery_object", false},
		{"generic_order_object_cleanup_protects_current_avatar", 1, `INSERT INTO customer_service_agent_profiles VALUES(1,'/uploads/customer-service/avatars/current-private-name.png'); INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-shared-avatar-cleanup','storage.object_cleanup_requested','order','55','{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["customer-service/avatars/current-private-name.png"]}','processed')`, "durable_cleanup_overlaps_retained_recovery_object", false},
		{"foreign_avatar_reference_is_incomplete", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-avatar-cleanup','customer_service.avatar_cleanup','customer_service_agent_profile','1','{"url":"https://foreign.invalid/avatar.png"}','processed')`, "avatar_cleanup_reference_unresolved", false},
		{"consumed_cleanup_receipt_keeps_its_fence", -1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,deleted_at) VALUES(11,'synthetic-consumed-cleanup','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["customer-service/attachments/originals/synthetic.txt"]}','processed','2026-10-01')`, "suppressed_cleanup_state_must_be_preserved", true},
		{"ignored_cleanup_is_not_reactivated", -1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-ignored-cleanup','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["customer-service/attachments/originals/synthetic.txt"]}','ignored')`, "suppressed_cleanup_state_must_be_preserved", true},
		{"payload_scalar_cannot_be_interpreted_as_an_empty_scope", 1, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-malformed-cleanup','storage.object_cleanup_requested','order','55','"legacy malformed text"','pending')`, "durable_event_payload_not_an_object", false},
		{"oversized_arbitrary_payload_body_is_not_transferred", -1, `UPDATE outbox_events SET payload=jsonb_set(payload,'{payload,private_body}',to_jsonb(repeat('private-body',100000))) WHERE id=10`, "", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticEventAndConversationDatabases()
			for index, database := range databases {
				if scenario.side < 0 || index == scenario.side {
					_, err := database.ExecContext(ctx, scenario.statement)
					require.NoError(t, err)
				}
			}
			report := inspectSyntheticDurableEventSnapshots(defaultModuleRecoveryRowInspectionLimits())
			if scenario.code != "" {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
			} else {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			}
			if scenario.reviewOnly {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			}
			if scenario.name == "matching_processed_realtime_event" {
				require.True(t, report.DurableEvents.Complete)
				require.Equal(t, 1, report.DurableEvents.ComparedEventIdentities)
				require.Equal(t, "same_record_preserved", report.DurableEvents.Comparisons[0].Result)
			}
		})
	}
	for _, eventSide := range []int{0, 1} {
		scenarioName := "package_cleanup_protects_target_only_avatar_owner"
		if eventSide == 1 {
			scenarioName = "target_cleanup_protects_package_only_avatar_owner"
		}
		t.Run(scenarioName, func(t *testing.T) {
			resetSyntheticEventAndConversationDatabases()
			_, err := databases[1-eventSide].ExecContext(ctx, `INSERT INTO customer_service_agent_profiles VALUES(1,'/uploads/customer-service/avatars/current-private-name.png')`)
			require.NoError(t, err)
			_, err = databases[eventSide].ExecContext(ctx, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(11,'synthetic-avatar-cleanup','customer_service.avatar_cleanup','customer_service_agent_profile','1','{"url":"/uploads/customer-service/avatars/current-private-name.png"}','processed')`)
			require.NoError(t, err)
			report := inspectSyntheticDurableEventSnapshots(defaultModuleRecoveryRowInspectionLimits())
			require.Positive(t, report.IssueCounts["durable_cleanup_overlaps_retained_recovery_object"])
		})
	}
	t.Run("event_resource_and_sample_limits_are_explicit", func(t *testing.T) {
		resetSyntheticEventAndConversationDatabases()
		for _, database := range databases {
			_, err := database.ExecContext(ctx, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) SELECT 11,'synthetic-second-event',event_type,aggregate_type,aggregate_id,jsonb_set(payload,'{event_id}','"synthetic-second-event"'),'processed' FROM outbox_events WHERE id=10`)
			require.NoError(t, err)
		}
		limits := defaultModuleRecoveryRowInspectionLimits()
		limits.IssueSamples = 1
		report := inspectSyntheticDurableEventSnapshots(limits)
		require.Equal(t, 2, report.DurableEvents.ComparedEventIdentities)
		require.Len(t, report.DurableEvents.Comparisons, 1)
		require.True(t, report.DurableEvents.ComparisonsTruncated)
		limits.Rows = 16
		report = inspectSyntheticDurableEventSnapshots(limits)
		require.False(t, report.DurableEvents.Complete)
		require.Positive(t, report.IssueCounts["row_inspection_resource_limit"])
	})
	if validationDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); validationDirectory != "" {
		resetSyntheticEventAndConversationDatabases()
		_, err := databases[1].ExecContext(ctx, `UPDATE outbox_events SET status='ignored' WHERE id=10; INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,locked_at,locked_by) VALUES(11,'synthetic-cleanup-with-lease','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["customer-service/attachments/originals/synthetic.txt"]}','processed',now(),'synthetic-reconciliation-worker')`)
		require.NoError(t, err)
		report := inspectSyntheticDurableEventSnapshots(defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, writeDurableJSON(filepath.Join(validationDirectory, "synthetic-customer-service-durable-event-state-and-cleanup-conflicts-from-postgresql-integration.json"), report))
	}
}
