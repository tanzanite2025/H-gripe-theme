package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/domain/outbox"
)

const legacyCustomerServiceCreationIncompleteEnvelopeError = "customer-service realtime outbox event is incomplete"

// This exception preserves an already suppressed row; it never repairs its
// payload, imports an audit, changes its state or authorizes a dispatcher replay.
// Both snapshots must pass independently, and their ordinary comparison still
// rejects a backup taken before the manual disposition.
func inspectAuditedIgnoredLegacyCustomerServiceCreationNotificationPreservation(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow) (bool, error) {
	aggregateID := event.referenceColumnText("aggregate_id")
	if event.referenceColumnText("event_type") != outbox.EventTypeCustomerServiceRealtime ||
		event.referenceColumnText("aggregate_type") != outbox.AggregateTypeCustomerServiceConversation ||
		!canonicalModuleRecoveryDurableEventPositiveID(aggregateID) ||
		event.referenceColumnText("event_key") != "customer_service.conversation.created:"+aggregateID ||
		moduleRecoveryDurableEventPublicStatus(event) != outbox.EventStatusIgnored ||
		event.referenceColumnText("_payload_type") != "" || event.referenceColumnText("_payload_event_id") != "" || event.referenceColumnText("_payload_ticket_id") != "" {
		return false, nil
	}
	reject := func(code string, incomplete bool) (bool, error) {
		recordModuleRecoveryDurableEventIssue(graph, event, code, "blocking", "历史创建通知缺少可核对的同事务忽略凭证、完整原行或安全保留状态；保持阻断，未重放或改写事件。", incomplete)
		return true, nil
	}
	for table, columns := range map[string]string{
		"outbox_events": "id event_key event_type aggregate_type aggregate_id payload status attempts max_attempts created_at updated_at last_error locked_at locked_by processed_at uncertain_at reconcile_after deleted_at",
		"audit_logs":    "id user_id action resource resource_id method path status old_value new_value created_at",
	} {
		for _, column := range strings.Fields(columns) {
			if !graph.Observation.Columns[table][column] {
				return reject("audited_legacy_creation_preservation_schema_missing", true)
			}
		}
	}
	if graph.Transaction == nil || graph.Exhausted {
		return reject("audited_legacy_creation_preservation_snapshot_unavailable", true)
	}
	for _, column := range strings.Fields("locked_at locked_by processed_at uncertain_at reconcile_after deleted_at") {
		if event.referenceColumnText(column) != "" {
			return reject("audited_legacy_creation_preservation_execution_state_conflict", false)
		}
	}
	// These private JSON bodies contain only the audit contract, not an outbox
	// payload. Bound them before transfer; none of their operator text, event
	// keys, account identities or hashes is copied into the public inspection.
	rows, err := graph.Transaction.QueryContext(ctx, `SELECT id::text,COALESCE(user_id,0),created_at,
CASE WHEN octet_length(old_value)<=65536 THEN old_value ELSE '' END,
CASE WHEN octet_length(new_value)<=65536 THEN new_value ELSE '' END
FROM public.audit_logs WHERE resource='outbox_event' AND resource_id=$1 AND action='reconcile'
AND status='success' AND method='POST' AND path=$2 ORDER BY id LIMIT 2`,
		event.referenceColumnText("id"), "/api/admin/ops/outbox/failures/"+event.referenceColumnText("id")+"/ignore")
	if err != nil {
		return true, err
	}
	type privateDispositionAudit struct {
		ID, Before, After string
		ActorID           int64
		CreatedAt         sql.NullTime
	}
	var audits []privateDispositionAudit
	for rows.Next() {
		var record privateDispositionAudit
		var before, after sql.NullString
		if err := rows.Scan(&record.ID, &record.ActorID, &record.CreatedAt, &before, &after); err != nil {
			_ = rows.Close()
			return true, err
		}
		record.Before, record.After = before.String, after.String
		bytes := len(record.Before) + len(record.After)
		if graph.ProjectionBytes+bytes > graph.Limits.ProjectionBytes || len(graph.ObservedRows) >= graph.Limits.Rows {
			_ = rows.Close()
			graph.Exhausted = true
			return reject("audited_legacy_creation_preservation_resource_limit", true)
		}
		graph.ProjectionBytes += bytes
		graph.ObservedRows["audit_logs:"+record.ID] = true
		audits = append(audits, record)
	}
	readErr, closeErr := rows.Err(), rows.Close()
	if readErr != nil {
		return true, readErr
	}
	if closeErr != nil {
		return true, closeErr
	}
	if len(audits) != 1 || audits[0].ActorID <= 0 || !audits[0].CreatedAt.Valid {
		return reject("audited_legacy_creation_preservation_receipt_missing_or_ambiguous", false)
	}
	var previous outbox.FailureDispositionAuditSnapshot
	var receipt outbox.FailureDispositionAuditReceipt
	if !decodeCanonicalCustomerServiceCreationDispositionAudit(audits[0].Before, &previous) ||
		!decodeCanonicalCustomerServiceCreationDispositionAudit(audits[0].After, &receipt) ||
		!customerServiceLegacyCreationDispositionSnapshotsMatchPreservedIgnoredEvent(previous, receipt, event, audits[0].CreatedAt.Time) {
		return reject("audited_legacy_creation_preservation_receipt_invalid", false)
	}
	// This is the writer's exact PostgreSQL/UTC hash contract. Reconstruct the
	// pre-ignore row by undoing only status and updated_at; every other column,
	// including unmodeled extensions and the original error, must be unchanged.
	if _, err := graph.Transaction.ExecContext(ctx, `SET LOCAL timezone='UTC'`); err != nil {
		return true, err
	}
	var knownPayloadShape, preservedState bool
	var currentRowSHA256, previousRowSHA256, payloadSHA256, errorSHA256 string
	var errorBytes int
	err = graph.Transaction.QueryRowContext(ctx, `SELECT
jsonb_typeof(record.payload)='object' AND record.payload ?& ARRAY['status','assigned_to']
AND record.payload-ARRAY['status','assigned_to']='{}'::jsonb AND record.payload->'status'='"open"'::jsonb
AND (record.payload->'assigned_to'='null'::jsonb OR (jsonb_typeof(record.payload->'assigned_to')='number'
AND record.payload->>'assigned_to' ~ '^[1-9][0-9]{0,18}$'
AND (length(record.payload->>'assigned_to')<19 OR record.payload->>'assigned_to'<='9223372036854775807'))),
encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex'),
encode(sha256(convert_to(jsonb_set(jsonb_set(to_jsonb(record),'{status}','"dead_letter"'::jsonb),'{updated_at}',to_jsonb($2::timestamptz))::text,'UTF8')),'hex'),
encode(sha256(convert_to(record.payload::text,'UTF8')),'hex'),
encode(sha256(convert_to(COALESCE(record.last_error,''),'UTF8')),'hex'),octet_length(COALESCE(record.last_error,'')),
record.status='ignored' AND record.deleted_at IS NULL AND record.locked_at IS NULL AND COALESCE(record.locked_by,'')=''
AND record.processed_at IS NULL AND record.uncertain_at IS NULL AND record.reconcile_after IS NULL
AND record.attempts=$4 AND record.max_attempts=$5 AND record.updated_at=$3 AND record.created_at<=$2
FROM public.outbox_events record WHERE record.id=$1`, event.referenceColumnText("id"), previous.UpdatedAt, receipt.Event.UpdatedAt, previous.Attempts, previous.MaxAttempts).
		Scan(&knownPayloadShape, &currentRowSHA256, &previousRowSHA256, &payloadSHA256, &errorSHA256, &errorBytes, &preservedState)
	if err == sql.ErrNoRows {
		return reject("audited_legacy_creation_preservation_row_missing", false)
	}
	if err != nil {
		return true, err
	}
	if !knownPayloadShape || !preservedState || currentRowSHA256 != event.Digest || currentRowSHA256 != receipt.Event.RowSHA256 ||
		previousRowSHA256 != previous.RowSHA256 || payloadSHA256 != previous.PayloadSHA256 || errorSHA256 != previous.LastErrorSHA256 || errorBytes != previous.LastErrorBytes {
		return reject("audited_legacy_creation_preservation_row_or_payload_mismatch", false)
	}
	var owner *moduleRecoveryDependencyRow
	for _, candidate := range graph.Rows["tickets"] {
		if candidate.referenceColumnText("id") == aggregateID {
			owner = candidate
			break
		}
	}
	if owner == nil || owner.referenceColumnText("category") != "customer_service" {
		return reject("audited_legacy_creation_preservation_conversation_missing", false)
	}
	recordModuleRecoveryDurableEventIssue(graph, event, "audited_legacy_creation_notification_preserved_without_replay", "review", "已核对历史创建通知的同事务忽略审计及完整原行；保留事件、失败证据和幂等键，不补造通知或重新投递。", false)
	return true, nil
}

func decodeCanonicalCustomerServiceCreationDispositionAudit(data string, result any) bool {
	if data == "" || len(data) > 65536 || decodeStrictJSON([]byte(data), result) != nil {
		return false
	}
	// The v1 writer uses json.Marshal on these structs. Require that exact
	// representation to reject duplicate fields and alternate JSON interpretations.
	canonical, err := json.Marshal(result)
	return err == nil && string(canonical) == data
}

func customerServiceLegacyCreationDispositionSnapshotsMatchPreservedIgnoredEvent(previous outbox.FailureDispositionAuditSnapshot, receipt outbox.FailureDispositionAuditReceipt, event *moduleRecoveryDependencyRow, auditedAt time.Time) bool {
	after := receipt.Event
	if receipt.Contract != outbox.FailureDispositionAuditContract || receipt.Operation != "ignore" ||
		receipt.Note != strings.TrimSpace(receipt.Note) || utf8.RuneCountInString(receipt.Note) < 3 || utf8.RuneCountInString(receipt.Note) > 2000 ||
		previous.Status != outbox.EventStatusDeadLetter || after.Status != outbox.EventStatusIgnored ||
		previous.MaxAttempts <= 0 || previous.Attempts < previous.MaxAttempts ||
		previous.UpdatedAt.IsZero() || after.UpdatedAt.Before(previous.UpdatedAt) || auditedAt.Before(after.UpdatedAt) ||
		previous.RowFingerprintFormat != "postgresql_complete_row_jsonb_sha256_v1" || after.RowFingerprintFormat != previous.RowFingerprintFormat ||
		strconv.FormatUint(uint64(after.EventID), 10) != event.referenceColumnText("id") ||
		after.EventKey != event.referenceColumnText("event_key") || after.EventType != event.referenceColumnText("event_type") ||
		after.AggregateType != event.referenceColumnText("aggregate_type") || after.AggregateID != event.referenceColumnText("aggregate_id") ||
		after.PayloadSHA256 != event.referenceColumnText("_payload_sha256") {
		return false
	}
	errorDigest := sha256.Sum256([]byte(legacyCustomerServiceCreationIncompleteEnvelopeError))
	if previous.LastErrorSHA256 != hex.EncodeToString(errorDigest[:]) || previous.LastErrorBytes != len(legacyCustomerServiceCreationIncompleteEnvelopeError) {
		return false
	}
	for _, digest := range []string{previous.RowSHA256, after.RowSHA256, after.PayloadSHA256} {
		if len(digest) != 64 || strings.ToLower(digest) != digest {
			return false
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return false
		}
	}
	previous.Status, previous.UpdatedAt, previous.RowSHA256 = after.Status, after.UpdatedAt, after.RowSHA256
	// All preserved identity, payload/error evidence and attempt fields must agree.
	return previous == after
}
