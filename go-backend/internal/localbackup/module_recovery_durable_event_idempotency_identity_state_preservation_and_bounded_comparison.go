package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/outbox"
)

type ModuleRecoveryDurableEventInspection struct {
	Version                 int                                                       `json:"version"`
	Complete                bool                                                      `json:"complete"`
	SourceScannedEvents     int                                                       `json:"source_scanned_events"`
	TargetScannedEvents     int                                                       `json:"target_scanned_events"`
	SourceRelatedEvents     int                                                       `json:"source_related_events"`
	TargetRelatedEvents     int                                                       `json:"target_related_events"`
	SourceStatuses          map[string]int                                            `json:"source_statuses"`
	TargetStatuses          map[string]int                                            `json:"target_statuses"`
	ComparedEventIdentities int                                                       `json:"compared_event_identities"`
	ComparisonSampleLimit   int                                                       `json:"comparison_sample_limit"`
	ComparisonsTruncated    bool                                                      `json:"comparisons_truncated"`
	Comparisons             []ModuleRecoveryDurableEventComparison                    `json:"comparisons"`
	StatePolicy             string                                                    `json:"state_policy"`
	CleanupDispositions     *CustomerServiceRecoveryCleanupEventDispositionInspection `json:"cleanup_dispositions,omitempty"`
}

type ModuleRecoveryDurableEventComparison struct {
	EventKeySHA256 string `json:"event_key_sha256"`
	SourceEventID  string `json:"source_event_id,omitempty"`
	TargetEventID  string `json:"target_event_id,omitempty"`
	SourceStatus   string `json:"source_status,omitempty"`
	TargetStatus   string `json:"target_status,omitempty"`
	Result         string `json:"result"`
}

type moduleRecoveryDurableEventSnapshot struct {
	Graph   *moduleRecoveryRowDependencyGraph
	Events  map[string]*moduleRecoveryDependencyRow
	Related map[string]bool
}

func canonicalModuleRecoveryDurableEventPositiveID(value string) bool {
	number, err := strconv.ParseInt(value, 10, 64)
	return err == nil && number > 0 && strconv.FormatInt(number, 10) == value
}

func moduleRecoveryDurableEventPublicStatus(row *moduleRecoveryDependencyRow) string {
	if row == nil {
		return ""
	}
	switch status := row.referenceColumnText("status"); status {
	case outbox.EventStatusPending, outbox.EventStatusProcessing, outbox.EventStatusProcessed, outbox.EventStatusFailed, outbox.EventStatusUnknown, outbox.EventStatusDeadLetter, outbox.EventStatusIgnored:
		return status
	default:
		return "invalid"
	}
}

func loadModuleRecoveryDurableEventCounterparts(ctx context.Context, snapshot *moduleRecoveryDurableEventSnapshot, events []*moduleRecoveryDependencyRow) error {
	graph := snapshot.Graph
	if !graph.Observation.Columns["outbox_events"]["event_key"] {
		return nil
	}
	for _, column := range []string{"id", "event_key"} {
		for start := 0; start < len(events) && !graph.Exhausted; start += 200 {
			var keys []map[string]json.RawMessage
			for _, event := range events[start:min(start+200, len(events))] {
				keys = append(keys, map[string]json.RawMessage{column: event.Values[column]})
			}
			matches, err := graph.readRowsMatchingReferenceKeys(ctx, "outbox_events", []string{column}, keys, graph.Limits.Rows)
			if err != nil {
				return err
			}
			for _, event := range matches {
				snapshot.Events[event.Identity] = event
			}
		}
	}
	return nil
}

func compareModuleRecoveryDurableEventSnapshots(ctx context.Context, source, target *moduleRecoveryDurableEventSnapshot, inspection *ModuleRecoveryDurableEventInspection) error {
	for _, pair := range [][2]*moduleRecoveryDurableEventSnapshot{{source, target}, {target, source}} {
		var related []*moduleRecoveryDependencyRow
		for identity := range pair[0].Related {
			related = append(related, pair[0].Events[identity])
		}
		sort.Slice(related, func(i, j int) bool { return related[i].Identity < related[j].Identity })
		if err := loadModuleRecoveryDurableEventCounterparts(ctx, pair[1], related); err != nil {
			return err
		}
	}
	if source.Graph.Exhausted || target.Graph.Exhausted {
		return nil
	}
	sourceByKey, targetByKey := map[string]*moduleRecoveryDependencyRow{}, map[string]*moduleRecoveryDependencyRow{}
	for _, pair := range []struct {
		snapshot *moduleRecoveryDurableEventSnapshot
		byKey    map[string]*moduleRecoveryDependencyRow
	}{{source, sourceByKey}, {target, targetByKey}} {
		for _, event := range pair.snapshot.Events {
			key := event.referenceColumnText("event_key")
			if previous := pair.byKey[key]; previous != nil && previous.Identity != event.Identity {
				pair.snapshot.Graph.Report.recordIssue("durable_event_key_not_unique", "blocking", pair.snapshot.Graph.Side, "outbox_events", event.referenceColumnText("id"), "同一幂等键对应多个事件行，不能自动选择或重放。")
			} else {
				pair.byKey[key] = event
			}
		}
	}
	keySet := map[string]bool{}
	for _, snapshot := range []*moduleRecoveryDurableEventSnapshot{source, target} {
		for identity := range snapshot.Related {
			keySet[snapshot.Events[identity].referenceColumnText("event_key")] = true
		}
	}
	var keys []string
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		left, right := sourceByKey[key], targetByKey[key]
		digest := sha256.Sum256([]byte(key))
		comparison := ModuleRecoveryDurableEventComparison{EventKeySHA256: hex.EncodeToString(digest[:]), SourceStatus: moduleRecoveryDurableEventPublicStatus(left), TargetStatus: moduleRecoveryDurableEventPublicStatus(right)}
		if left != nil {
			comparison.SourceEventID = left.referenceColumnText("id")
		}
		if right != nil {
			comparison.TargetEventID = right.referenceColumnText("id")
		}
		switch {
		case left == nil:
			comparison.Result = "target_only_preserved"
			target.Graph.Report.recordIssue("target_only_durable_event_preserved", "review", "target", "outbox_events", comparison.TargetEventID, "目标新增事件继续保留，不因较旧的模块恢复点而删除或重建。")
		case right == nil:
			comparison.Result = "source_only_requires_reconciliation"
			source.Graph.Report.recordIssue("source_durable_event_missing_from_target", "blocking", "target", "outbox_events", comparison.SourceEventID, "包内事件在目标中缺失，处理结果未知，不能补入后自动消费。")
		case !moduleRecoveryDurableEventRoutingIdentityMatches(left, right):
			comparison.Result = "idempotency_identity_conflict"
			source.Graph.Report.recordIssue("durable_event_key_identity_conflict", "blocking", "target", "outbox_events", comparison.TargetEventID, "同一幂等键对应不同事件类型、聚合身份或载荷，禁止覆盖。")
		case left.Identity != right.Identity:
			comparison.Result = "event_identifier_moved"
			source.Graph.Report.recordIssue("durable_event_identifier_moved", "blocking", "target", "outbox_events", comparison.TargetEventID, "同一幂等事件在两侧主键不同，需要保留现有身份并核对引用。")
		case left.Digest != right.Digest:
			comparison.Result = "current_state_must_be_preserved"
			source.Graph.Report.recordIssue("durable_event_current_state_differs", "blocking", "target", "outbox_events", comparison.TargetEventID, "事件当前内容或处理状态不同，须保留目标状态、尝试次数和幂等记录。")
		default:
			comparison.Result = "same_record_preserved"
		}
		if left != nil && right != nil && moduleRecoveryDurableEventExecutionStateDiffers(left, right) {
			source.Graph.Report.recordIssue("durable_event_replay_state_conflict", "blocking", "target", "outbox_events", comparison.TargetEventID, "处理、忽略、租约或已消费状态与恢复点不同，不能回退为待执行任务。")
		}
		if left != nil {
			if collision := target.Events[left.Identity]; collision != nil && collision.referenceColumnText("event_key") != key {
				source.Graph.Report.recordIssue("durable_event_primary_key_reassigned", "blocking", "target", "outbox_events", left.referenceColumnText("id"), "包内事件主键已由其他业务事件占用，不能按主键覆盖。")
			}
		}
		inspection.ComparedEventIdentities++
		if len(inspection.Comparisons) < inspection.ComparisonSampleLimit {
			inspection.Comparisons = append(inspection.Comparisons, comparison)
		} else {
			inspection.ComparisonsTruncated = true
		}
	}
	return nil
}

func moduleRecoveryDurableEventRoutingIdentityMatches(source, target *moduleRecoveryDependencyRow) bool {
	for _, column := range strings.Fields("event_type aggregate_type aggregate_id") {
		if source.referenceColumnText(column) != target.referenceColumnText(column) {
			return false
		}
	}
	// PostgreSQL computes this from the entire payload without returning it.
	return source.referenceColumnText("_payload_sha256") == target.referenceColumnText("_payload_sha256")
}

func moduleRecoveryDurableEventExecutionStateDiffers(source, target *moduleRecoveryDependencyRow) bool {
	for _, column := range strings.Fields("status attempts max_attempts available_at last_attempt_at locked_at locked_by processed_at uncertain_at reconcile_after deleted_at") {
		if source.referenceColumnText(column) != target.referenceColumnText(column) {
			return true
		}
	}
	return false
}

func recordModuleRecoveryDurableEventIssue(graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, code, severity, message string, incomplete bool) {
	if incomplete {
		graph.Report.Complete = false
	}
	graph.Report.recordIssue(code, severity, graph.Side, "outbox_events", event.referenceColumnText("id"), message)
}

func moduleRecoveryDurableEventPayloadPositiveID(event *moduleRecoveryDependencyRow, column string) string {
	raw := event.Values[column]
	if len(raw) == 0 || raw[0] == '"' {
		return ""
	}
	value := event.referenceColumnText(column)
	if !canonicalModuleRecoveryDurableEventPositiveID(value) {
		return ""
	}
	return value
}
