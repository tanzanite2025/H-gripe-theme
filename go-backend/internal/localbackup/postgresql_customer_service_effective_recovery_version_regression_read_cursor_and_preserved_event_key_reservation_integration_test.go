package localbackup

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func insertCustomerServiceVersionedRealtimeReservationForRecoveryTest(t *testing.T, fixture *customerServiceDisposableCandidateTestFixture, eventKey, payloadType, eventStatus string, payload map[string]any, softDeleted, unrelatedRouting bool) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"ticket_id": 1, "conversation_id": "synthetic-conversation-one", "type": payloadType, "event_id": eventKey, "payload": payload})
	require.NoError(t, err)
	eventType, aggregateType := "customer_service.realtime", "customer_service_conversation"
	if unrelatedRouting {
		eventType, aggregateType = "order.paid", "order"
	}
	_, err = fixture.Target.ExecContext(context.Background(), `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,deleted_at) VALUES(100,$1,$2,$3,'1',$4,$5,CASE WHEN $6 THEN now() ELSE NULL END)`, eventKey, eventType, aggregateType, string(body), eventStatus, softDeleted)
	require.NoError(t, err)
}

func TestPostgreSQLCustomerServiceEffectiveRecoveryRejectsVersionRegressionAndReservedEventIdentityReuseWithoutWrites(t *testing.T) {
	for _, scenario := range []struct {
		name, sourceSQL, targetSQL, eventKey, payloadType, status, expectedIssue string
		payload                                                                  map[string]any
		choices                                                                  map[string]string
		softDeleted, unrelatedRouting, satisfied                                 bool
	}{
		{name: "unchanged_versions_are_valid", satisfied: true},
		{name: "status_version_rewind_blocks_selected_package_ticket", targetSQL: `UPDATE tickets SET status_version=6 WHERE id=1`, expectedIssue: "realtime_version_counter_would_rewind"},
		{name: "preserve_current_ticket_resolves_status_version_rewind", targetSQL: `UPDATE tickets SET status_version=6 WHERE id=1`, choices: map[string]string{"tickets:1": "preserve_current_row"}, satisfied: true},
		{name: "assignment_version_rewind_blocks_selected_inbox", targetSQL: `UPDATE customer_service_inbox_states SET assignment_version=6 WHERE id=1`, expectedIssue: "realtime_version_counter_would_rewind"},
		{name: "preserve_current_inbox_resolves_assignment_version_rewind", targetSQL: `UPDATE customer_service_inbox_states SET assignment_version=6 WHERE id=1`, choices: map[string]string{"customer_service_inbox_states:1": "preserve_current_row"}, satisfied: true},
		{name: "read_cursor_rewind_in_same_assignment_epoch_blocks", sourceSQL: `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1`, expectedIssue: "realtime_read_event_identity_would_be_reused"},
		{name: "preserve_current_inbox_resolves_read_cursor_rewind", sourceSQL: `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1`, choices: map[string]string{"customer_service_inbox_states:1": "preserve_current_row"}, satisfied: true},
		{name: "reactivating_read_cursor_with_unread_state_blocks", sourceSQL: `UPDATE customer_service_inbox_states SET unread_count=2 WHERE id=1`, expectedIssue: "realtime_read_event_identity_would_be_reused"},
		{name: "higher_assignment_epoch_allows_reset_cursor", sourceSQL: `UPDATE customer_service_inbox_states SET assignment_version=2,last_read_message_id=0,unread_count=2 WHERE id=1`, satisfied: true},
		{name: "invalid_zero_status_version_blocks", sourceSQL: `UPDATE tickets SET status_version=0 WHERE id=1`, expectedIssue: "realtime_version_identity_invalid_or_exhausted"},
		{name: "exhausted_status_version_blocks", sourceSQL: `UPDATE tickets SET status_version=2147483647 WHERE id=1`, expectedIssue: "realtime_version_identity_invalid_or_exhausted"},
		{name: "exhausted_assignment_version_blocks", sourceSQL: `UPDATE customer_service_inbox_states SET assignment_version=2147483647 WHERE id=1`, expectedIssue: "realtime_version_identity_invalid_or_exhausted"},
		{name: "processed_status_receipt_ahead_blocks", eventKey: "customer_service.conversation.status.changed:1:4", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 4}, expectedIssue: "realtime_preserved_event_version_ahead_of_effective_row"},
		{name: "ignored_status_receipt_still_reserves_key", eventKey: "customer_service.conversation.status.changed:1:4", payloadType: "conversation.status.changed", status: "ignored", payload: map[string]any{"status_version": 4}, expectedIssue: "realtime_preserved_event_version_ahead_of_effective_row"},
		{name: "soft_deleted_status_receipt_still_reserves_key", eventKey: "customer_service.conversation.status.changed:1:4", payloadType: "conversation.status.changed", softDeleted: true, payload: map[string]any{"status_version": 4}, expectedIssue: "realtime_preserved_event_version_ahead_of_effective_row"},
		{name: "matching_status_version_is_valid", eventKey: "customer_service.conversation.status.changed:1:1", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 1}, satisfied: true},
		{name: "newer_selected_version_exceeds_reserved_version", sourceSQL: `UPDATE tickets SET status_version=7 WHERE id=1`, eventKey: "customer_service.conversation.status.changed:1:4", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 4}, satisfied: true},
		{name: "event_key_and_payload_versions_disagree", eventKey: "customer_service.conversation.status.changed:1:4", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 3}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "string_payload_version_is_not_numeric", eventKey: "customer_service.conversation.status.changed:1:1", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": "1"}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "version_beyond_postgresql_integer_range_blocks", eventKey: "customer_service.conversation.status.changed:1:2147483648", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": int64(2147483648)}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "truncated_reserved_event_key_is_rejected", eventKey: "customer_service.conversation.assigned:1", payloadType: "conversation.assigned", payload: map[string]any{"assignment_version": 1, "assigned_to": 3}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "noncanonical_reserved_version_is_rejected", eventKey: "customer_service.conversation.status.changed:1:01", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 1}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "opaque_historical_version_event_needs_contract", eventKey: "synthetic-private-opaque-historical-version-event", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 1}, expectedIssue: "realtime_version_event_key_contract_unrecognized"},
		{name: "different_event_type_cannot_occupy_reserved_key", eventKey: "customer_service.conversation.status.changed:1:1", payloadType: "conversation.status.changed", unrelatedRouting: true, payload: map[string]any{"status_version": 1}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "assignment_receipt_ahead_blocks", eventKey: "customer_service.conversation.assigned:1:3:4", payloadType: "conversation.assigned", payload: map[string]any{"assignment_version": 4, "assigned_to": 3}, expectedIssue: "realtime_preserved_event_version_ahead_of_effective_row"},
		{name: "assignment_recipient_disagrees_with_reserved_key", eventKey: "customer_service.conversation.assigned:1:3:1", payloadType: "conversation.assigned", payload: map[string]any{"assignment_version": 1, "assigned_to": 2}, expectedIssue: "realtime_version_event_key_payload_conflict"},
		{name: "missing_recipient_state_cannot_restart_at_first_epoch", sourceSQL: `DELETE FROM customer_service_inbox_states WHERE id=1`, targetSQL: `DELETE FROM customer_service_inbox_states WHERE id=1`, eventKey: "customer_service.conversation.assigned:1:3:1", payloadType: "conversation.assigned", payload: map[string]any{"assignment_version": 1, "assigned_to": 3}, expectedIssue: "realtime_version_event_owner_missing"},
		{name: "preserved_read_receipt_ahead_of_cursor_blocks", sourceSQL: `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1`, targetSQL: `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1`, eventKey: "customer_service.messages.read:1:3:1:1", payloadType: "conversation.messages.read", payload: map[string]any{"assignment_version": 1, "read_by_user_id": 3, "last_read_message_id": 1}, expectedIssue: "realtime_preserved_read_cursor_ahead_of_effective_row"},
		{name: "matching_read_receipt_is_valid", eventKey: "customer_service.messages.read:1:3:1:1", payloadType: "conversation.messages.read", payload: map[string]any{"assignment_version": 1, "read_by_user_id": 3, "last_read_message_id": 1}, satisfied: true},
		{name: "matching_archive_receipt_is_valid", eventKey: "customer_service.conversation.inbox_state.changed:1:3:1:archived:2026-10-21T01:02:03Z", payloadType: "conversation.inbox_state.changed", payload: map[string]any{"assignment_version": 1, "recipient_user_id": 3}, satisfied: true},
		{name: "reserved_conversation_and_payload_disagree", eventKey: "customer_service.conversation.status.changed:2:8", payloadType: "conversation.status.changed", payload: map[string]any{"status_version": 8}, expectedIssue: "realtime_version_event_key_payload_conflict"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			for _, change := range []struct {
				statement string
				source    bool
			}{{scenario.sourceSQL, true}, {scenario.targetSQL, false}} {
				if change.statement == "" {
					continue
				}
				database := fixture.Target
				if change.source {
					database = fixture.Source
				}
				_, err := database.ExecContext(ctx, change.statement)
				require.NoError(t, err)
			}
			if scenario.eventKey != "" {
				status := scenario.status
				if status == "" {
					status = "processed"
				}
				insertCustomerServiceVersionedRealtimeReservationForRecoveryTest(t, fixture, scenario.eventKey, scenario.payloadType, status, scenario.payload, scenario.softDeleted, scenario.unrelatedRouting)
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, scenario.choices)
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.Equal(t, scenario.satisfied, selection.SelectionDependenciesSatisfied, "%+v / original %+v", selection.DependencyChecks.IssueCounts, selection.RepairRequiredIssueCounts)
			if scenario.expectedIssue != "" {
				require.Positive(t, selection.DependencyChecks.IssueCounts[scenario.expectedIssue], "%+v", selection.DependencyChecks.IssueCounts)
			}
			require.False(t, selection.ExecutionAvailable)
			require.False(t, selection.CandidateConstructionAvailable)
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			for _, privateValue := range []string{scenario.eventKey, "private retained synthetic message", "synthetic-conversation-one", "private synthetic customer name"} {
				if privateValue != "" {
					require.NotContains(t, string(encoded), privateValue)
				}
			}
			for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
				_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, graph.Transaction)
				require.NoError(t, err)
				require.Equal(t, graph.Observation.Fingerprint, fingerprint)
			}
			require.NotContains(t, strings.Join(plan.UnimplementedSafeguards, ","), "candidate_capacity_validation_and_independent_durable_journal")
			t.Logf("effective version reservation: accepted=%t issues=%v", scenario.satisfied, selection.DependencyChecks.IssueCounts)
		})
	}
}

func TestPostgreSQLCustomerServicePersistentCandidateRejectsRewoundRealtimeVersionBeforeAllocation(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	_, err := fixture.Target.ExecContext(context.Background(), `UPDATE tickets SET status_version=9 WHERE id=1`)
	require.NoError(t, err)
	cfg := fixture.Configuration
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
	_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
	require.ErrorContains(t, err, "repair blockers")
	var candidates int
	require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
	require.Zero(t, candidates)
	var version int
	require.NoError(t, fixture.Target.QueryRowContext(context.Background(), `SELECT status_version FROM tickets WHERE id=1`).Scan(&version))
	require.Equal(t, 9, version)
	t.Logf("rewound version refused before candidate allocation; current version remains %d", version)
}
