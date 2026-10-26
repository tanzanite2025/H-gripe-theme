package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createCustomerServiceLegacyCreationFixtureWithActualTransactionalIgnoreAudits(t *testing.T, payload, originalError string) *customerServiceDisposableCandidateTestFixture {
	t.Helper()
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	if payload == "" {
		payload = `{"status":"open","assigned_to":3}`
	}
	if originalError == "" {
		originalError = legacyCustomerServiceCreationIncompleteEnvelopeError
	}
	for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
		_, err := database.ExecContext(t.Context(), `ALTER TABLE outbox_events ADD COLUMN max_attempts int NOT NULL DEFAULT 10;
ALTER TABLE outbox_events ADD COLUMN available_at timestamptz NOT NULL DEFAULT '2026-10-01Z';
ALTER TABLE outbox_events ADD COLUMN last_attempt_at timestamptz;
ALTER TABLE outbox_events ADD COLUMN last_error text;
ALTER TABLE outbox_events ADD COLUMN created_at timestamptz NOT NULL DEFAULT '2026-10-01Z';
ALTER TABLE outbox_events ADD COLUMN updated_at timestamptz NOT NULL DEFAULT '2026-10-01Z';
ALTER TABLE outbox_events ADD COLUMN synthetic_preserved_extension text NOT NULL DEFAULT 'synthetic-private-extension';`)
		require.NoError(t, err)
		_, err = database.ExecContext(t.Context(), `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,attempts,max_attempts,last_error,last_attempt_at)
VALUES(100,'customer_service.conversation.created:1','customer_service.realtime','customer_service_conversation','1',$1::jsonb,'dead_letter',10,10,$2,'2026-10-01Z')`, payload, originalError)
		require.NoError(t, err)
		orm, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		require.NoError(t, orm.AutoMigrate(&audit.AuditLog{}))
		result, err := repository.NewOutboxRepository(orm).ReconcileFailedEventWithTransactionalAudit(100, "ignore", "synthetic-private-reviewed-creation-notification", time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), &audit.AuditLog{UserID: 3, Method: "POST", Path: "/api/admin/ops/outbox/failures/100/ignore"})
		require.NoError(t, err)
		require.NotZero(t, result.AuditID)
	}
	return fixture
}

func changeCustomerServiceLegacyCreationAuditUsingCanonicalWriterJSON(t *testing.T, database *sql.DB, change func(*outbox.FailureDispositionAuditSnapshot, *outbox.FailureDispositionAuditReceipt)) {
	t.Helper()
	var oldValue, newValue string
	require.NoError(t, database.QueryRowContext(t.Context(), `SELECT old_value,new_value FROM audit_logs WHERE resource_id=100`).Scan(&oldValue, &newValue))
	var previous outbox.FailureDispositionAuditSnapshot
	var receipt outbox.FailureDispositionAuditReceipt
	require.NoError(t, json.Unmarshal([]byte(oldValue), &previous))
	require.NoError(t, json.Unmarshal([]byte(newValue), &receipt))
	change(&previous, &receipt)
	oldBytes, err := json.Marshal(previous)
	require.NoError(t, err)
	newBytes, err := json.Marshal(receipt)
	require.NoError(t, err)
	_, err = database.ExecContext(t.Context(), `UPDATE audit_logs SET old_value=$1,new_value=$2 WHERE resource_id=100`, string(oldBytes), string(newBytes))
	require.NoError(t, err)
}

func inspectCustomerServiceLegacyCreationPreservationWithoutWriting(t *testing.T, fixture *customerServiceDisposableCandidateTestFixture, databases []*sql.DB) *ModuleRecoveryRowInspection {
	t.Helper()
	checks := newSelectedCustomerServiceModuleDomainChecks()
	var graphs []*moduleRecoveryRowDependencyGraph
	for index, database := range databases {
		tx, err := database.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		require.NoError(t, err)
		defer tx.Rollback()
		observation, err := inspectModuleRecoveryDatabaseObservation(t.Context(), tx)
		require.NoError(t, err)
		side := "target"
		if len(databases) == 2 && index == 0 {
			side = "package"
		}
		graph := newModuleRecoveryRowDependencyGraph(tx, observation, side, checks, defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, graph.selectCustomerServiceConversationScopeRows(t.Context()))
		graphs = append(graphs, graph)
	}
	require.NoError(t, fixture.Configuration.inspectCustomerServiceDurableEventDependencies(t.Context(), graphs, checks))
	for _, graph := range graphs {
		_, fingerprint, _, err := captureDatabaseTransactionInventory(t.Context(), graph.Transaction)
		require.NoError(t, err)
		require.Equal(t, graph.Observation.Fingerprint, fingerprint, "Event, audit and all unrelated rows must remain unchanged")
	}
	encoded, err := json.Marshal(checks)
	require.NoError(t, err)
	for _, privateValue := range []string{"synthetic-private-reviewed-creation-notification", "synthetic-private-extension", "customer_service.conversation.created:1", "private retained synthetic message"} {
		require.NotContains(t, string(encoded), privateValue)
	}
	return checks
}

func TestPostgreSQLCustomerServiceAuditedLegacyCreationPreservationRejectsMissingForgedOrChangedEvidence(t *testing.T) {
	for _, scenario := range []struct {
		name, payload, originalError, mutation, issue string
		changeReceipt                                 func(*outbox.FailureDispositionAuditSnapshot, *outbox.FailureDispositionAuditReceipt)
		accepted                                      bool
	}{
		{name: "actual_atomic_ignore_with_assigned_operator_preserves_legacy_row", accepted: true},
		{name: "actual_atomic_ignore_with_unassigned_payload_preserves_legacy_row", payload: `{"status":"open","assigned_to":null}`, accepted: true},
		{name: "missing_audit_table_blocks", mutation: `DROP TABLE audit_logs`, issue: "audited_legacy_creation_preservation_schema_missing"},
		{name: "missing_audit_column_blocks", mutation: `ALTER TABLE audit_logs DROP COLUMN user_id`, issue: "audited_legacy_creation_preservation_schema_missing"},
		{name: "missing_event_evidence_column_blocks", mutation: `ALTER TABLE outbox_events DROP COLUMN last_error`, issue: "audited_legacy_creation_preservation_schema_missing"},
		{name: "missing_audit_blocks", mutation: `DELETE FROM audit_logs`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "failed_audit_is_not_disposition_proof", mutation: `UPDATE audit_logs SET status='failed'`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "unattributed_audit_blocks", mutation: `UPDATE audit_logs SET user_id=0`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "wrong_resource_blocks", mutation: `UPDATE audit_logs SET resource_id=101`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "wrong_http_action_blocks", mutation: `UPDATE audit_logs SET path='/api/admin/ops/outbox/failures/100/retry'`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "wrong_http_method_blocks", mutation: `UPDATE audit_logs SET method='GET'`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "duplicate_successful_ignore_is_ambiguous", mutation: `INSERT INTO audit_logs(user_id,action,resource,resource_id,method,path,status,old_value,new_value,created_at) SELECT user_id,action,resource,resource_id,method,path,status,old_value,new_value,created_at FROM audit_logs`, issue: "audited_legacy_creation_preservation_receipt_missing_or_ambiguous"},
		{name: "old_best_effort_audit_is_not_a_receipt", mutation: `UPDATE audit_logs SET old_value='',new_value='{"operation":"ignore","note":"old best effort"}'`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "oversized_audit_body_is_bounded_and_rejected", mutation: `UPDATE audit_logs SET new_value=repeat('synthetic-private-review',4000)`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "duplicate_json_fields_are_rejected", mutation: `UPDATE audit_logs SET new_value='{"operation":"retry",'||substring(new_value from 2)`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "audit_predates_its_disposition", mutation: `UPDATE audit_logs SET created_at='2026-10-01Z'`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "unknown_audit_contract_blocks", changeReceipt: func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			receipt.Contract = "unknown.contract"
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "retry_receipt_cannot_authorize_preserved_ignore", changeReceipt: func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			receipt.Operation = "retry"
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "sqlite_hash_is_not_a_postgresql_receipt", changeReceipt: func(previous *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			previous.RowFingerprintFormat, receipt.Event.RowFingerprintFormat = "domain_event_json_sha256_v1", "domain_event_json_sha256_v1"
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "previous_failed_status_does_not_inherit_dead_letter_exception", changeReceipt: func(previous *outbox.FailureDispositionAuditSnapshot, _ *outbox.FailureDispositionAuditReceipt) {
			previous.Status = outbox.EventStatusFailed
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "unexhausted_attempts_do_not_inherit_exception", changeReceipt: func(previous *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			previous.Attempts, receipt.Event.Attempts = 1, 1
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "changed_audit_identity_blocks", changeReceipt: func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			receipt.Event.AggregateID = "2"
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "forged_original_row_hash_blocks", changeReceipt: func(previous *outbox.FailureDispositionAuditSnapshot, _ *outbox.FailureDispositionAuditReceipt) {
			previous.RowSHA256 = strings.Repeat("a", 64)
		}, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "forged_final_row_hash_blocks", changeReceipt: func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			receipt.Event.RowSHA256 = strings.Repeat("a", 64)
		}, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "payload_changed_after_ignore_blocks", mutation: `UPDATE outbox_events SET payload='{"status":"open","assigned_to":null}' WHERE id=100`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "original_error_changed_after_ignore_blocks", mutation: `UPDATE outbox_events SET last_error='synthetic-private-new-error' WHERE id=100`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "attempts_changed_after_ignore_blocks", mutation: `UPDATE outbox_events SET attempts=11 WHERE id=100`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "unmodeled_column_changed_after_ignore_blocks", mutation: `UPDATE outbox_events SET synthetic_preserved_extension='changed-private-extension' WHERE id=100`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "lock_owner_blocks", mutation: `UPDATE outbox_events SET locked_by='synthetic-worker' WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "lock_timestamp_blocks", mutation: `UPDATE outbox_events SET locked_at=now() WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "uncertain_execution_blocks", mutation: `UPDATE outbox_events SET uncertain_at=now() WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "reconciliation_schedule_blocks", mutation: `UPDATE outbox_events SET reconcile_after=now() WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "processed_timestamp_blocks", mutation: `UPDATE outbox_events SET processed_at=now() WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "soft_deletion_does_not_replace_audit_preservation", mutation: `UPDATE outbox_events SET deleted_at=now() WHERE id=100`, issue: "audited_legacy_creation_preservation_execution_state_conflict"},
		{name: "missing_conversation_blocks", mutation: `UPDATE outbox_events SET aggregate_id='99',event_key='customer_service.conversation.created:99' WHERE id=100`, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "arbitrary_payload_members_are_not_historical_creation_shape", payload: `{"status":"open","assigned_to":3,"private_body":"synthetic-private-arbitrary-content"}`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "other_creation_status_does_not_match_known_defect", payload: `{"status":"closed","assigned_to":3}`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "string_assignee_does_not_match_known_defect", payload: `{"status":"open","assigned_to":"3"}`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "zero_assignee_does_not_match_known_defect", payload: `{"status":"open","assigned_to":0}`, issue: "audited_legacy_creation_preservation_row_or_payload_mismatch"},
		{name: "missing_reason_blocks", changeReceipt: func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
			receipt.Note = ""
		}, issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "other_original_error_does_not_match_known_defect", originalError: "synthetic-private-different-failure", issue: "audited_legacy_creation_preservation_receipt_invalid"},
		{name: "pending_status_cannot_inherit_exception", mutation: `UPDATE outbox_events SET status='pending' WHERE id=100`, issue: "realtime_event_conversation_identity_conflict"},
		{name: "dead_letter_status_cannot_inherit_exception", mutation: `UPDATE outbox_events SET status='dead_letter' WHERE id=100`, issue: "realtime_event_conversation_identity_conflict"},
		{name: "versioned_notification_cannot_inherit_exception", mutation: `UPDATE outbox_events SET event_key='customer_service.conversation.status.changed:1:1' WHERE id=100`, issue: "realtime_event_conversation_identity_conflict"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceLegacyCreationFixtureWithActualTransactionalIgnoreAudits(t, scenario.payload, scenario.originalError)
			if scenario.mutation != "" {
				_, err := fixture.Target.ExecContext(t.Context(), scenario.mutation)
				require.NoError(t, err)
			}
			if scenario.changeReceipt != nil {
				changeCustomerServiceLegacyCreationAuditUsingCanonicalWriterJSON(t, fixture.Target, scenario.changeReceipt)
			}
			checks := inspectCustomerServiceLegacyCreationPreservationWithoutWriting(t, fixture, []*sql.DB{fixture.Target})
			if scenario.accepted {
				require.True(t, checks.Complete)
				require.Zero(t, checks.BlockingIssues, "%v", checks.IssueCounts)
				require.Equal(t, 1, checks.IssueCounts["audited_legacy_creation_notification_preserved_without_replay"])
			} else {
				require.Positive(t, checks.BlockingIssues, "%v", checks.IssueCounts)
				require.Positive(t, checks.IssueCounts[scenario.issue], "%v", checks.IssueCounts)
				require.Zero(t, checks.IssueCounts["audited_legacy_creation_notification_preserved_without_replay"])
			}
		})
	}
}

func TestPostgreSQLCustomerServiceAuditedLegacyCreationPreservationReconstructsOriginalRowAndEnforcesAuditResourceBudget(t *testing.T) {
	fixture := createCustomerServiceLegacyCreationFixtureWithActualTransactionalIgnoreAudits(t, "", "")
	_, err := fixture.Target.ExecContext(t.Context(), `UPDATE outbox_events SET synthetic_preserved_extension='changed-private-extension' WHERE id=100`)
	require.NoError(t, err)
	var changedRowHash string
	require.NoError(t, fixture.Target.QueryRowContext(t.Context(), `SELECT encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex') FROM outbox_events record WHERE id=100`).Scan(&changedRowHash))
	changeCustomerServiceLegacyCreationAuditUsingCanonicalWriterJSON(t, fixture.Target, func(_ *outbox.FailureDispositionAuditSnapshot, receipt *outbox.FailureDispositionAuditReceipt) {
		receipt.Event.RowSHA256 = changedRowHash
	})
	checks := inspectCustomerServiceLegacyCreationPreservationWithoutWriting(t, fixture, []*sql.DB{fixture.Target})
	require.Positive(t, checks.IssueCounts["audited_legacy_creation_preservation_row_or_payload_mismatch"], "A new final hash cannot hide a changed original field")
	tx, err := fixture.Source.BeginTx(t.Context(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	observation, err := inspectModuleRecoveryDatabaseObservation(t.Context(), tx)
	require.NoError(t, err)
	checks = newSelectedCustomerServiceModuleDomainChecks()
	graph := newModuleRecoveryRowDependencyGraph(tx, observation, "target", checks, defaultModuleRecoveryRowInspectionLimits())
	rows, err := graph.readBoundedRelationshipRows(t.Context(), "outbox_events", "record.id=100", nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	graph.ProjectionBytes = graph.Limits.ProjectionBytes - 1
	handled, err := inspectAuditedIgnoredLegacyCustomerServiceCreationNotificationPreservation(t.Context(), graph, rows[0])
	require.NoError(t, err)
	require.True(t, handled)
	require.True(t, graph.Exhausted)
	require.False(t, checks.Complete)
	require.Positive(t, checks.IssueCounts["audited_legacy_creation_preservation_resource_limit"])
}

func TestPostgreSQLCustomerServiceAuditedLegacyCreationPreservationKeepsSourceCurrentComparisonsAndDomainBoundaries(t *testing.T) {
	fixture := createCustomerServiceLegacyCreationFixtureWithActualTransactionalIgnoreAudits(t, "", "")
	checks := inspectCustomerServiceLegacyCreationPreservationWithoutWriting(t, fixture, []*sql.DB{fixture.Source, fixture.Target})
	require.Zero(t, checks.BlockingIssues, "%v", checks.IssueCounts)
	require.Equal(t, 2, checks.IssueCounts["audited_legacy_creation_notification_preserved_without_replay"])
	_, err := fixture.Source.ExecContext(t.Context(), `UPDATE outbox_events SET status='dead_letter',updated_at='2026-10-01Z' WHERE id=100;DELETE FROM audit_logs`)
	require.NoError(t, err)
	checks = inspectCustomerServiceLegacyCreationPreservationWithoutWriting(t, fixture, []*sql.DB{fixture.Source, fixture.Target})
	require.Positive(t, checks.IssueCounts["durable_event_current_state_differs"])
	require.Positive(t, checks.IssueCounts["durable_event_replay_state_conflict"])
	for _, eventType := range []string{"order.paid", "payment.refund_execution_requested", "storage.object_cleanup_requested", "referral.registered"} {
		t.Run(eventType, func(t *testing.T) {
			event := &moduleRecoveryDependencyRow{Values: map[string]json.RawMessage{}}
			for key, value := range map[string]string{"id": "100", "event_type": eventType, "aggregate_type": outbox.AggregateTypeCustomerServiceConversation, "aggregate_id": "1", "status": "ignored", "event_key": "customer_service.conversation.created:1"} {
				event.Values[key] = json.RawMessage(strconv.Quote(value))
			}
			handled, err := inspectAuditedIgnoredLegacyCustomerServiceCreationNotificationPreservation(context.Background(), nil, event)
			require.NoError(t, err)
			require.False(t, handled, "A historical realtime notification exception must never authorize other domains")
		})
	}
}
