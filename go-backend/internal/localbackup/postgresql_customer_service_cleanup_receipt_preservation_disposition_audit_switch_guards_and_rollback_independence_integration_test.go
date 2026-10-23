package localbackup

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLCustomerServiceCleanupDispositionAuditsEveryCurrentReceiptWithoutChangingEventState(t *testing.T) {
	for _, scenario := range []struct {
		name, eventType, status, payload, execution, issue string
		deleted, leased, overlap, blocked                  bool
	}{
		{name: "pending_disjoint_object_receipt_can_dispatch", status: "pending", execution: "dispatcher_may_execute"},
		{name: "failed_disjoint_object_receipt_can_retry", status: "failed", execution: "dispatcher_may_execute"},
		{name: "processed_object_receipt_can_reconcile_again", status: "processed", execution: "original_receipt_reconciliation_may_execute"},
		{name: "dead_letter_object_receipt_can_reconcile_again", status: "dead_letter", execution: "original_receipt_reconciliation_may_execute"},
		{name: "unknown_object_outcome_still_blocks", status: "unknown", execution: "original_receipt_reconciliation_may_execute", blocked: true, issue: "durable_event_outcome_unknown"},
		{name: "processing_object_receipt_blocks", status: "processing", execution: "in_flight_or_leased", blocked: true, issue: "durable_event_execution_lease_present"},
		{name: "processed_receipt_with_reconciliation_lease_blocks", status: "processed", leased: true, execution: "in_flight_or_leased", blocked: true, issue: "durable_event_execution_lease_present"},
		{name: "ignored_overlap_preserves_original_suppression", status: "ignored", overlap: true, execution: "suppressed_original_receipt"},
		{name: "soft_deleted_pending_overlap_stays_suppressed", status: "pending", deleted: true, overlap: true, execution: "suppressed_original_receipt"},
		{name: "processed_overlap_still_blocks", status: "processed", overlap: true, execution: "original_receipt_reconciliation_may_execute", blocked: true, issue: "durable_cleanup_overlaps_retained_recovery_object"},
		{name: "pending_overlap_blocks", status: "pending", overlap: true, execution: "dispatcher_may_execute", blocked: true, issue: "durable_cleanup_overlaps_retained_recovery_object"},
		{name: "dead_letter_overlap_blocks", status: "dead_letter", overlap: true, execution: "original_receipt_reconciliation_may_execute", blocked: true, issue: "durable_cleanup_overlaps_retained_recovery_object"},
		{name: "unregistered_disjoint_resource_is_not_declared_safe", status: "pending", payload: `{"resource_type":"synthetic_unknown_private_resource","resource_id":"55","object_keys":["unrelated-private-object.txt"]}`, execution: "dispatcher_may_execute", blocked: true, issue: "cleanup_disposition_resource_contract_unregistered"},
		{name: "malformed_reference_list_has_blocked_disposition", status: "pending", payload: `{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":{}}`, execution: "dispatcher_may_execute", blocked: true, issue: "durable_cleanup_reference_list_invalid"},
		{name: "non_object_payload_has_blocked_disposition", status: "pending", payload: `[]`, execution: "dispatcher_may_execute", blocked: true, issue: "durable_event_payload_not_an_object"},
		{name: "invalid_suppressed_status_still_blocks", status: "synthetic_unknown_status", deleted: true, execution: "unsupported_execution_state", blocked: true, issue: "durable_event_status_unregistered"},
		{name: "retention_pending_for_absent_conversation_can_dispatch", eventType: outbox.EventTypeCustomerServiceRetentionCleanup, status: "pending", execution: "dispatcher_may_execute"},
		{name: "retention_processed_history_does_not_periodically_dispatch", eventType: outbox.EventTypeCustomerServiceRetentionCleanup, status: "processed", execution: "retained_history_no_automatic_dispatch"},
		{name: "retention_dead_letter_history_stays_retained", eventType: outbox.EventTypeCustomerServiceRetentionCleanup, status: "dead_letter", execution: "retained_history_no_automatic_dispatch"},
		{name: "retention_same_conversation_blocks_even_with_no_files", eventType: outbox.EventTypeCustomerServiceRetentionCleanup, status: "pending", payload: `{"ticket_id":1,"attachment_references":[]}`, overlap: true, execution: "dispatcher_may_execute", blocked: true, issue: "durable_cleanup_overlaps_retained_recovery_object"},
		{name: "avatar_processed_disjoint_receipt_can_reconcile", eventType: outbox.EventTypeCustomerServiceAvatarCleanup, status: "processed", execution: "original_receipt_reconciliation_may_execute"},
		{name: "avatar_ignored_disjoint_receipt_stays_suppressed", eventType: outbox.EventTypeCustomerServiceAvatarCleanup, status: "ignored", execution: "suppressed_original_receipt"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			eventType, aggregateType, aggregateID, payload := scenario.eventType, "order", "55", scenario.payload
			if eventType == "" {
				eventType = outbox.EventTypeObjectStorageCleanup
			}
			if payload == "" {
				payload = `{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["unrelated-private-object.txt"]}`
				if scenario.overlap {
					payload = `{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["legacy-attachment.txt"]}`
				}
				if eventType == outbox.EventTypeCustomerServiceRetentionCleanup {
					payload = `{"ticket_id":99,"attachment_references":[]}`
				}
				if eventType == outbox.EventTypeCustomerServiceAvatarCleanup {
					payload = `{"url":"/uploads/customer-service/avatars/unrelated-private-avatar.png"}`
				}
			}
			if eventType == outbox.EventTypeCustomerServiceRetentionCleanup {
				aggregateType, aggregateID = outbox.AggregateTypeCustomerServiceConversation, "99"
				if scenario.overlap {
					aggregateID = "1"
				}
			}
			if eventType == outbox.EventTypeCustomerServiceAvatarCleanup {
				aggregateType, aggregateID = outbox.AggregateTypeCustomerServiceAgentProfile, "99"
			}
			_, err := fixture.Target.ExecContext(context.Background(), `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,locked_at,locked_by,deleted_at) VALUES(51,'synthetic-private-cleanup-receipt-key',$1,$2,$3,$4::jsonb,$5,CASE WHEN $6 THEN clock_timestamp() END,CASE WHEN $6 THEN 'synthetic-private-lease' ELSE '' END,CASE WHEN $7 THEN clock_timestamp() END)`, eventType, aggregateType, aggregateID, payload, scenario.status, scenario.leased, scenario.deleted)
			require.NoError(t, err)
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(context.Background(), source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.Equal(t, !scenario.blocked, selection.SelectionDependenciesSatisfied, "%v / original %v", selection.DependencyChecks.IssueCounts, selection.RepairRequiredIssueCounts)
			inspection := selection.DependencyChecks.DurableEvents.CleanupDispositions
			require.NotNil(t, inspection)
			require.Equal(t, 1, inspection.TargetReceipts)
			require.Len(t, inspection.Samples, 1)
			entry := inspection.Samples[0]
			require.Equal(t, scenario.execution, entry.FutureExecution)
			require.Equal(t, scenario.overlap, entry.OverlapsRequiredObject)
			require.Equal(t, scenario.deleted, entry.SoftDeleted)
			if scenario.blocked {
				require.Equal(t, "block_restore_requires_cleanup_reconciliation", entry.Action)
				require.Positive(t, selection.DependencyChecks.IssueCounts[scenario.issue])
			} else {
				require.NoError(t, validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(inspection.entries))
			}
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			for _, privateValue := range []string{"synthetic-private-cleanup-receipt-key", "synthetic-private-lease", "unrelated-private-object.txt", "unrelated-private-avatar.png", "synthetic_unknown_private_resource", "private retained synthetic message"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
				_, fingerprint, _, err := captureDatabaseTransactionInventory(context.Background(), graph.Transaction)
				require.NoError(t, err)
				require.Equal(t, graph.Observation.Fingerprint, fingerprint)
			}
		})
	}
}

func TestPostgreSQLCustomerServiceCleanupDispositionFullAuditSurvivesDiscardAndGuardsSwitchWithoutPreventingRollback(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	cfg := fixture.Configuration
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	_, err := fixture.Source.ExecContext(context.Background(), `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format,is_internal) VALUES(100,1,'private package message','[]','legacy_urls',true)`)
	require.NoError(t, err)
	_, err = fixture.Target.ExecContext(context.Background(), `UPDATE ticket_messages SET content='private current body' WHERE id=1;
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,attempts) VALUES
(51,'synthetic-private-ignored-overlap','storage.object_cleanup_requested','media_asset','1','{"resource_type":"media_asset","resource_id":"1","object_keys":["legacy-attachment.txt"]}','ignored',17),
(52,'synthetic-private-disjoint-pending','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":[]}','pending',3),
(53,'synthetic-private-avatar-processed','customer_service.avatar_cleanup','customer_service_agent_profile','99','{"url":"/uploads/customer-service/avatars/unrelated-private-avatar.png"}','processed',8);
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) SELECT receipt_id,'synthetic-private-extra-cleanup-'||receipt_id,'storage.object_cleanup_requested','order','55',jsonb_build_object('resource_type','private_order_business_attachment_failed_upload','resource_id','55','object_keys',jsonb_build_array('unrelated-private-object-'||receipt_id)),'pending' FROM generate_series(100,304) receipt_id`)
	require.NoError(t, err)
	var expectedEventFingerprint string
	const eventFingerprintQuery = `SELECT encode(sha256(convert_to(jsonb_agg(to_jsonb(record) ORDER BY id)::text,'UTF8')),'hex') FROM outbox_events record`
	require.NoError(t, fixture.Target.QueryRowContext(context.Background(), eventFingerprintQuery).Scan(&expectedEventFingerprint))
	manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
	selection, err := manager.InspectCustomerServiceRecoveryEffectiveSelection(ctx, task.ID, request)
	require.NoError(t, err)
	require.True(t, selection.SelectionDependenciesSatisfied, "%v / original %v", selection.DependencyChecks.IssueCounts, selection.RepairRequiredIssueCounts)
	require.Equal(t, 208, selection.DependencyChecks.DurableEvents.CleanupDispositions.TargetReceipts)
	require.Len(t, selection.DependencyChecks.DurableEvents.CleanupDispositions.Samples, 200)
	require.True(t, selection.DependencyChecks.DurableEvents.CleanupDispositions.SamplesTruncated)
	record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
	require.NoError(t, err)
	require.NoError(t, manager.verifyCustomerServiceCandidateBoundCleanupDispositionAudit(record))
	directory, err := manager.taskDirectory(task.ID)
	require.NoError(t, err)
	auditFilename := filepath.Join(directory, customerServiceRecoveryCleanupDispositionAuditFilename(record.AttemptID))
	auditBytes, err := os.ReadFile(auditFilename)
	require.NoError(t, err)
	var audit customerServiceRecoveryCleanupEventDispositionAudit
	require.NoError(t, decodeStrictJSON(auditBytes, &audit))
	require.Len(t, audit.Entries, 208)
	require.NotContains(t, string(auditBytes), "synthetic-private-")
	require.Equal(t, record.Validation.CleanupDispositionEntriesSHA256, audit.EntriesSHA256)
	for _, corruption := range []string{"missing", "invalid_json", "changed_digest", "other_task", "other_attempt", "other_plan", "other_selection", "other_target", "changed_policy", "changed_receipt_state"} {
		t.Run(corruption, func(t *testing.T) {
			var corruptedAudit customerServiceRecoveryCleanupEventDispositionAudit
			require.NoError(t, decodeStrictJSON(auditBytes, &corruptedAudit))
			switch corruption {
			case "missing":
				require.NoError(t, os.Remove(auditFilename))
			case "invalid_json":
				require.NoError(t, os.WriteFile(auditFilename, []byte("{"), 0600))
			default:
				switch corruption {
				case "changed_digest":
					corruptedAudit.SHA256 = strings.Repeat("0", 64)
				case "other_task":
					corruptedAudit.TaskID = newTaskID()
				case "other_attempt":
					corruptedAudit.AttemptID = newTaskID()
				case "other_plan":
					corruptedAudit.PlanSHA256 = strings.Repeat("0", 64)
				case "other_selection":
					corruptedAudit.SelectionSHA256 = strings.Repeat("0", 64)
				case "other_target":
					corruptedAudit.TargetDatabaseFingerprint = strings.Repeat("0", 64)
				case "changed_policy":
					corruptedAudit.Policy = "replay_all_cleanup"
				case "changed_receipt_state":
					corruptedAudit.Entries[0].OriginalStatus = "pending"
					corruptedAudit.Entries[0].Action = "delete_original_receipt"
				}
				require.NoError(t, writeDurableJSON(auditFilename, &corruptedAudit))
			}
			_, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
			require.Error(t, err)
			var currentOID int64
			require.NoError(t, readDatabaseOID(ctx, fixture.Control, cfg.DatabaseName, &currentOID))
			require.NotEqual(t, record.Identity.DatabaseOID, currentOID)
			require.NoError(t, os.WriteFile(auditFilename, auditBytes, 0600))
		})
	}
	journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
	require.NoError(t, err)
	require.NoError(t, os.Remove(auditFilename))
	require.Error(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
	require.NoError(t, manager.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID), "missing audit must not prevent safe rollback")
	require.NoError(t, os.WriteFile(auditFilename, auditBytes, 0600))
	journal, err = manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
	require.NoError(t, err)
	require.NoError(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
	_, err = fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`)
	require.NoError(t, err)
	candidate, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	var actualEventFingerprint string
	require.NoError(t, candidate.QueryRowContext(context.Background(), eventFingerprintQuery).Scan(&actualEventFingerprint))
	require.Equal(t, expectedEventFingerprint, actualEventFingerprint, "all current outbox rows, payloads, attempts and fences must stay exact")
	require.NoError(t, candidate.Close())
	require.NoError(t, sealDatabase(ctx, fixture.Control, cfg.DatabaseName))
	require.NoError(t, os.Remove(auditFilename))
	require.NoError(t, manager.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID), "audit loss after a switch must not prevent restoring the original database and media")
	verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t, fixture, manager, task, record)
	require.NoError(t, os.WriteFile(auditFilename, auditBytes, 0600))
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	retainedAuditBytes, err := os.ReadFile(auditFilename)
	require.NoError(t, err)
	require.Equal(t, auditBytes, retainedAuditBytes, "disposal must retain the complete attempt audit")
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	if evidenceDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); evidenceDirectory != "" {
		require.NoError(t, os.MkdirAll(evidenceDirectory, 0700))
		require.NoError(t, os.WriteFile(filepath.Join(evidenceDirectory, "customer-service-cleanup-disposition-full-208-receipt-preservation-audit.json"), auditBytes, 0600))
	}
}

func TestPostgreSQLCustomerServiceUnsafeCleanupDispositionIsRejectedBeforeCandidateAllocation(t *testing.T) {
	for _, resourceType := range []string{"private_order_business_attachment_failed_upload", "synthetic_unknown_private_resource", "package_only_cleanup_receipt"} {
		t.Run(resourceType, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			cfg := fixture.Configuration
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			var err error
			if resourceType == "package_only_cleanup_receipt" {
				_, err = fixture.Source.ExecContext(context.Background(), `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(51,'synthetic-private-package-only-cleanup','customer_service.retention_cleanup_requested','customer_service_conversation','99','{"ticket_id":99,"attachment_references":[]}','pending')`)
			} else {
				_, err = fixture.Target.ExecContext(context.Background(), `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(51,'synthetic-private-unsafe-cleanup','storage.object_cleanup_requested','order','55',jsonb_build_object('resource_type',$1::text,'resource_id','55','object_keys',jsonb_build_array('legacy-attachment.txt')),'processed')`, resourceType)
			}
			require.NoError(t, err)
			manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
			_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
			require.ErrorContains(t, err, "repair blockers")
			var candidates int
			require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE datname ~ '^local_backup_candidate_'`).Scan(&candidates))
			require.Zero(t, candidates)
			record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Empty(t, record.CleanupDispositionAuditSHA256)
			require.Nil(t, record.MediaUnion)
			if resourceType == "package_only_cleanup_receipt" {
				var importedReceipts int
				require.NoError(t, fixture.Target.QueryRowContext(context.Background(), `SELECT count(*) FROM outbox_events WHERE id=51`).Scan(&importedReceipts))
				require.Zero(t, importedReceipts)
			}
		})
	}
}
