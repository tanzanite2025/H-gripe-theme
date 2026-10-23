package localbackup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"

	"commerce-platform/internal/domain/outbox"
)

const customerServiceCleanupDispositionPolicy = "preserve_current_cleanup_receipts_and_fences_no_import_no_replay_block_unsafe_restore"
const customerServiceCleanupDispositionMaximumAuditBytes = 32 << 20

type CustomerServiceRecoveryCleanupEventDisposition struct {
	EventID                string `json:"event_id"`
	EventType              string `json:"event_type"`
	EventKeySHA256         string `json:"event_key_sha256"`
	PreservedRowSHA256     string `json:"preserved_row_sha256"`
	OriginalStatus         string `json:"original_status"`
	SoftDeleted            bool   `json:"soft_deleted"`
	ExecutionLeasePresent  bool   `json:"execution_lease_present"`
	OverlapsRequiredObject bool   `json:"overlaps_required_object"`
	FutureExecution        string `json:"future_execution"`
	Action                 string `json:"action"`
}

type CustomerServiceRecoveryCleanupEventDispositionInspection struct {
	Version          int                                              `json:"version"`
	Complete         bool                                             `json:"complete"`
	Policy           string                                           `json:"policy"`
	TargetReceipts   int                                              `json:"target_receipts"`
	ActionCounts     map[string]int                                   `json:"action_counts"`
	EntriesSHA256    string                                           `json:"entries_sha256"`
	SampleLimit      int                                              `json:"sample_limit"`
	SamplesTruncated bool                                             `json:"samples_truncated"`
	Samples          []CustomerServiceRecoveryCleanupEventDisposition `json:"samples"`
	entries          []CustomerServiceRecoveryCleanupEventDisposition
}

type customerServiceRecoveryCleanupEventDispositionAudit struct {
	Version                   int                                              `json:"version"`
	TaskID                    string                                           `json:"task_id"`
	AttemptID                 string                                           `json:"attempt_id"`
	PlanSHA256                string                                           `json:"plan_sha256"`
	SelectionSHA256           string                                           `json:"selection_sha256"`
	TargetDatabaseFingerprint string                                           `json:"target_database_fingerprint"`
	Policy                    string                                           `json:"policy"`
	EntriesSHA256             string                                           `json:"entries_sha256"`
	Entries                   []CustomerServiceRecoveryCleanupEventDisposition `json:"entries"`
	SHA256                    string                                           `json:"sha256"`
}

func isCustomerServiceRecoveryAuditedCleanupEvent(eventType string) bool {
	return eventType == outbox.EventTypeCustomerServiceRetentionCleanup || eventType == outbox.EventTypeCustomerServiceAvatarCleanup || eventType == outbox.EventTypeObjectStorageCleanup
}

// These names mirror ObjectStorageCleanupOutboxHandler.Handle. New handlers
// require a recovery contract before an effective selection may retain them.
func customerServiceRecoveryObjectCleanupResourceContractRegistered(resourceType string) bool {
	switch resourceType {
	case "media_asset", "site_logo", "site_favicon", "home_visual_tile", "ugc_showcase", "ugc_showcase_failed_upload",
		"unregistered_media_upload", "unregistered_media_derivative", "retired_media_derivative", "customer_service_unregistered_upload",
		"expired_business_draft_media_upload", "order_evidence_failed_upload", "private_order_business_attachment_failed_upload",
		outbox.SuggestionAndPayPalAttachmentCleanupResourceType:
		return true
	}
	return false
}

func customerServiceRecoveryCleanupReceiptFutureExecution(eventType, status string, deleted, leased bool) string {
	switch status {
	case outbox.EventStatusPending, outbox.EventStatusProcessing, outbox.EventStatusProcessed, outbox.EventStatusFailed, outbox.EventStatusUnknown, outbox.EventStatusDeadLetter, outbox.EventStatusIgnored:
	default:
		return "unsupported_execution_state"
	}
	if deleted || status == outbox.EventStatusIgnored {
		return "suppressed_original_receipt"
	}
	if leased || status == outbox.EventStatusProcessing {
		return "in_flight_or_leased"
	}
	if status == outbox.EventStatusPending || status == outbox.EventStatusFailed {
		return "dispatcher_may_execute"
	}
	if eventType != outbox.EventTypeCustomerServiceRetentionCleanup && (status == outbox.EventStatusProcessed || status == outbox.EventStatusUnknown || status == outbox.EventStatusDeadLetter) {
		return "original_receipt_reconciliation_may_execute"
	}
	if status == outbox.EventStatusProcessed || status == outbox.EventStatusDeadLetter {
		return "retained_history_no_automatic_dispatch"
	}
	return "unsupported_execution_state"
}

func recordCustomerServiceRecoveryCleanupEventDisposition(graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, overlaps bool, previousBlockingIssues int) {
	inspection := graph.Report.DurableEvents.CleanupDispositions
	keyDigest := sha256.Sum256([]byte(event.referenceColumnText("event_key")))
	entry := CustomerServiceRecoveryCleanupEventDisposition{
		EventID: event.referenceColumnText("id"), EventType: event.referenceColumnText("event_type"),
		EventKeySHA256: hex.EncodeToString(keyDigest[:]), PreservedRowSHA256: event.Digest,
		OriginalStatus: moduleRecoveryDurableEventPublicStatus(event), SoftDeleted: event.referenceColumnText("deleted_at") != "",
		ExecutionLeasePresent:  event.referenceColumnText("locked_at") != "" || event.referenceColumnText("locked_by") != "",
		OverlapsRequiredObject: overlaps, Action: "preserve_current_receipt_no_replay",
	}
	entry.FutureExecution = customerServiceRecoveryCleanupReceiptFutureExecution(entry.EventType, entry.OriginalStatus, entry.SoftDeleted, entry.ExecutionLeasePresent)
	if entry.FutureExecution == "suppressed_original_receipt" {
		entry.Action = "preserve_suppressed_receipt_no_replay"
	}
	if graph.Report.BlockingIssues > previousBlockingIssues {
		entry.Action = "block_restore_requires_cleanup_reconciliation"
	}
	inspection.entries = append(inspection.entries, entry)
}

func finalizeCustomerServiceRecoveryCleanupEventDispositions(inspection *CustomerServiceRecoveryCleanupEventDispositionInspection, complete bool) error {
	inspection.Complete = complete
	sort.Slice(inspection.entries, func(i, j int) bool {
		return inspection.entries[i].EventKeySHA256 < inspection.entries[j].EventKeySHA256
	})
	inspection.TargetReceipts = len(inspection.entries)
	inspection.ActionCounts = map[string]int{}
	for _, entry := range inspection.entries {
		inspection.ActionCounts[entry.Action]++
	}
	inspection.Samples = append([]CustomerServiceRecoveryCleanupEventDisposition{}, inspection.entries[:min(len(inspection.entries), inspection.SampleLimit)]...)
	inspection.SamplesTruncated = len(inspection.entries) > len(inspection.Samples)
	var err error
	inspection.EntriesSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(inspection.entries)
	return err
}

func customerServiceRecoveryCleanupDispositionAuditFilename(attemptID string) string {
	return "customer-service-cleanup-event-preservation-disposition-audit-" + attemptID + ".json"
}

func isCustomerServiceCleanupDispositionCanonicalSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}

func validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(entries []CustomerServiceRecoveryCleanupEventDisposition) error {
	if len(entries) > 20000 {
		return errors.New("cleanup disposition audit exceeds its receipt limit")
	}
	previousKey := ""
	seenIDs := map[string]bool{}
	for _, entry := range entries {
		if !canonicalModuleRecoveryDurableEventPositiveID(entry.EventID) || seenIDs[entry.EventID] || !isCustomerServiceRecoveryAuditedCleanupEvent(entry.EventType) || !isCustomerServiceCleanupDispositionCanonicalSHA256(entry.EventKeySHA256) || !isCustomerServiceCleanupDispositionCanonicalSHA256(entry.PreservedRowSHA256) || entry.EventKeySHA256 <= previousKey {
			return errors.New("cleanup disposition audit receipt identity or ordering is invalid")
		}
		seenIDs[entry.EventID], previousKey = true, entry.EventKeySHA256
		expectedExecution := customerServiceRecoveryCleanupReceiptFutureExecution(entry.EventType, entry.OriginalStatus, entry.SoftDeleted, entry.ExecutionLeasePresent)
		expectedAction := "preserve_current_receipt_no_replay"
		if expectedExecution == "suppressed_original_receipt" {
			expectedAction = "preserve_suppressed_receipt_no_replay"
		}
		if entry.FutureExecution != expectedExecution || entry.Action != expectedAction || entry.OriginalStatus == outbox.EventStatusUnknown || entry.OriginalStatus == "invalid" || expectedExecution == "unsupported_execution_state" || expectedExecution == "in_flight_or_leased" || entry.OverlapsRequiredObject && expectedExecution != "suppressed_original_receipt" {
			return errors.New("cleanup disposition audit contains an unsafe preservation decision")
		}
	}
	return nil
}

func (manager *Manager) persistCustomerServiceCandidateBoundCleanupDispositionAudit(record *customerServicePersistentCandidateLifecycle, selection *CustomerServiceEffectiveSelectionInspection) error {
	if selection == nil || selection.DependencyChecks == nil || selection.DependencyChecks.DurableEvents == nil || record.Binding == nil || !safeTaskID.MatchString(record.AttemptID) || selection.PlanSHA256 != record.Request.PlanSHA256 || selection.TaskID != record.TaskID || selection.Binding.TargetDatabaseFingerprint != record.Binding.TargetDatabaseFingerprint {
		return errors.New("cleanup disposition audit requires the stopped candidate selection binding")
	}
	inspection := selection.DependencyChecks.DurableEvents.CleanupDispositions
	if inspection == nil || !inspection.Complete || inspection.Policy != customerServiceCleanupDispositionPolicy || inspection.TargetReceipts != len(inspection.entries) {
		return errors.New("complete cleanup disposition audit is required before candidate allocation")
	}
	if err := validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(inspection.entries); err != nil {
		return err
	}
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(inspection.entries)
	if err != nil || digest != inspection.EntriesSHA256 {
		return errors.New("cleanup disposition entries changed before persistence")
	}
	audit := &customerServiceRecoveryCleanupEventDispositionAudit{Version: 1, TaskID: record.TaskID, AttemptID: record.AttemptID, PlanSHA256: selection.PlanSHA256, SelectionSHA256: selection.SHA256, TargetDatabaseFingerprint: selection.Binding.TargetDatabaseFingerprint, Policy: customerServiceCleanupDispositionPolicy, EntriesSHA256: digest, Entries: inspection.entries}
	audit.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(audit)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(audit, "", "  ")
	if err != nil || len(encoded) > customerServiceCleanupDispositionMaximumAuditBytes {
		return errors.New("cleanup disposition audit exceeds its byte limit")
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(encoded)) * 2); err != nil {
		return err
	}
	directory, err := manager.taskDirectory(record.TaskID)
	if err != nil {
		return err
	}
	if err := writeDurableFile(filepath.Join(directory, customerServiceRecoveryCleanupDispositionAuditFilename(record.AttemptID)), encoded); err != nil {
		return err
	}
	record.CleanupDispositionAuditSHA256 = audit.SHA256
	return manager.saveCustomerServicePersistentCandidateLifecycle(record)
}

func (manager *Manager) verifyCustomerServiceCandidateBoundCleanupDispositionAudit(record *customerServicePersistentCandidateLifecycle) error {
	if record.Binding == nil || record.Validation == nil || record.CleanupDispositionAuditSHA256 == "" || !safeTaskID.MatchString(record.AttemptID) {
		return errors.New("candidate cleanup disposition audit binding is missing")
	}
	directory, err := manager.taskDirectory(record.TaskID)
	if err != nil {
		return err
	}
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(directory, customerServiceRecoveryCleanupDispositionAuditFilename(record.AttemptID)), customerServiceCleanupDispositionMaximumAuditBytes)
	if err != nil {
		return errors.New("candidate cleanup disposition audit is missing or unsafe")
	}
	var audit customerServiceRecoveryCleanupEventDispositionAudit
	if decodeStrictJSON(data, &audit) != nil {
		return errors.New("candidate cleanup disposition audit is invalid")
	}
	copyOfAudit := audit
	copyOfAudit.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfAudit)
	entriesDigest, entriesErr := customerServiceRecoveryPlanCanonicalJSONSHA256(audit.Entries)
	if err != nil || entriesErr != nil || digest != audit.SHA256 || digest != record.CleanupDispositionAuditSHA256 || audit.Version != 1 || audit.TaskID != record.TaskID || audit.AttemptID != record.AttemptID || audit.PlanSHA256 != record.Request.PlanSHA256 || audit.SelectionSHA256 != record.Validation.SelectionSHA256 || audit.TargetDatabaseFingerprint != record.Binding.TargetDatabaseFingerprint || audit.Policy != customerServiceCleanupDispositionPolicy || audit.EntriesSHA256 != entriesDigest || audit.EntriesSHA256 != record.Validation.CleanupDispositionEntriesSHA256 {
		return errors.New("candidate cleanup disposition audit snapshot binding changed")
	}
	return validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(audit.Entries)
}
