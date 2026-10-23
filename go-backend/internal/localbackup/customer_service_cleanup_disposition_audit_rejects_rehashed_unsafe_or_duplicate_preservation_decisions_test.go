package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceCleanupDispositionAuditKeepsLegacyCandidateDigestReadableForSafeDisposal(t *testing.T) {
	manager, task, record := createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t)
	record.Binding = &CustomerServiceRecoveryPlanSnapshotBinding{RegistryVersion: 4}
	record.Validation = &customerServiceDisposableCandidateRehearsalResult{Version: 1, State: "rehearsed_not_switchable", PlanSHA256: record.Request.PlanSHA256, SelectionSHA256: strings.Repeat("b", 64)}
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	directory, err := manager.taskDirectory(task.ID)
	require.NoError(t, err)
	legacyBytes, err := os.ReadFile(filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename))
	require.NoError(t, err)
	require.NotContains(t, string(legacyBytes), "cleanup_disposition_entries_sha256")
	require.NotContains(t, string(legacyBytes), "cleanup_disposition_audit_sha256")
	loaded, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, record.SHA256, loaded.SHA256)
	require.ErrorContains(t, manager.verifyCustomerServiceCandidateBoundCleanupDispositionAudit(loaded), "binding is missing")
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
	loaded, err = manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, "discarded_not_switchable", loaded.State)
}

func TestCustomerServiceCleanupDispositionAuditRejectsUnsafeOrDuplicateDecisionsEvenWithFreshDigests(t *testing.T) {
	baseline := CustomerServiceRecoveryCleanupEventDisposition{EventID: "51", EventType: outbox.EventTypeObjectStorageCleanup, EventKeySHA256: strings.Repeat("a", 64), PreservedRowSHA256: strings.Repeat("b", 64), OriginalStatus: "processed", FutureExecution: "original_receipt_reconciliation_may_execute", Action: "preserve_current_receipt_no_replay"}
	require.NoError(t, validateCustomerServiceRecoveryPreservedCleanupDispositionEntries([]CustomerServiceRecoveryCleanupEventDisposition{baseline}))
	for _, scenario := range []string{"duplicate_receipt", "duplicate_key", "noncanonical_event_id", "invalid_row_digest", "unregistered_event_type", "unknown_outcome", "lease_present", "active_overlap", "suppressed_invalid_status", "invented_execution_class", "destructive_action", "out_of_order_keys", "over_limit"} {
		t.Run(scenario, func(t *testing.T) {
			entry := baseline
			entries := []CustomerServiceRecoveryCleanupEventDisposition{entry}
			switch scenario {
			case "duplicate_receipt":
				entry.EventKeySHA256 = strings.Repeat("c", 64)
				entries = append(entries, entry)
			case "duplicate_key":
				entry.EventID = "52"
				entries = append(entries, entry)
			case "noncanonical_event_id":
				entries[0].EventID = "051"
			case "invalid_row_digest":
				entries[0].PreservedRowSHA256 = "private-invalid-digest"
			case "unregistered_event_type":
				entries[0].EventType = "order.paid"
			case "unknown_outcome":
				entries[0].OriginalStatus = "unknown"
			case "lease_present":
				entries[0].ExecutionLeasePresent = true
				entries[0].FutureExecution = "in_flight_or_leased"
			case "active_overlap":
				entries[0].OverlapsRequiredObject = true
			case "suppressed_invalid_status":
				entries[0].SoftDeleted, entries[0].OriginalStatus, entries[0].FutureExecution, entries[0].Action = true, "synthetic_unknown_status", "suppressed_original_receipt", "preserve_suppressed_receipt_no_replay"
			case "invented_execution_class":
				entries[0].FutureExecution = "never_runs_again"
			case "destructive_action":
				entries[0].Action = "delete_original_receipt"
			case "out_of_order_keys":
				entry.EventID, entry.EventKeySHA256 = "52", strings.Repeat("9", 64)
				entries = append(entries, entry)
			case "over_limit":
				entries = make([]CustomerServiceRecoveryCleanupEventDisposition, 20001)
			}
			_, err := customerServiceRecoveryPlanCanonicalJSONSHA256(entries)
			require.NoError(t, err)
			require.Error(t, validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(entries))
		})
	}
}
