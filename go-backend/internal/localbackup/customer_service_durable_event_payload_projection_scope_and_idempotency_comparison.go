package localbackup

import (
	"context"
	"sort"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/outbox"
)

// Extract only registered routing identities and object references. In
// particular neither realtime message bodies nor arbitrary payload members
// leave PostgreSQL. The caller still applies its 64 KiB projection limit.
func customerServiceDurableEventPayloadProjectionEntries() []string {
	payload := `(to_jsonb(record)->'payload')`
	expressions := map[string]string{
		"_payload_sha256":                `encode(sha256(convert_to((` + payload + `)::text,'UTF8')),'hex')`,
		"_payload_shape":                 `jsonb_typeof(` + payload + `)`,
		"_payload_ticket_id":             payload + `->'ticket_id'`,
		"_payload_conversation_id":       payload + `->'conversation_id'`,
		"_payload_event_id":              payload + `->'event_id'`,
		"_payload_type":                  payload + `->'type'`,
		"_payload_message_id":            payload + `->'payload'->'message_id'`,
		"_payload_last_read_message_id":  payload + `->'payload'->'last_read_message_id'`,
		"_payload_status_version":        payload + `->'payload'->'status_version'`,
		"_payload_assignment_version":    payload + `->'payload'->'assignment_version'`,
		"_payload_read_by_user_id":       payload + `->'payload'->'read_by_user_id'`,
		"_payload_assigned_to":           payload + `->'payload'->'assigned_to'`,
		"_payload_recipient_user_id":     payload + `->'payload'->'recipient_user_id'`,
		"_payload_resource_type":         payload + `->'resource_type'`,
		"_payload_resource_id":           payload + `->'resource_id'`,
		"_payload_object_keys":           payload + `->'object_keys'`,
		"_payload_attachment_references": payload + `->'attachment_references'`,
		"_payload_avatar_url":            payload + `->'url'`,
	}
	var names []string
	for name := range expressions {
		names = append(names, name)
	}
	sort.Strings(names)
	var entries []string
	for _, name := range names {
		entries = append(entries, pq.QuoteLiteral(name), expressions[name])
	}
	return entries
}

type customerServiceDurableEventScope struct {
	Tickets     map[string]bool
	MediaAssets map[string]bool
	ExactPaths  map[string]bool
	ObjectKeys  map[string]bool
	AvatarPaths map[string]bool
}

func (cfg Configuration) collectCustomerServiceDurableEventRecoveryObjectScope(graphs []*moduleRecoveryRowDependencyGraph) *customerServiceDurableEventScope {
	scope := &customerServiceDurableEventScope{Tickets: map[string]bool{}, MediaAssets: map[string]bool{}, ExactPaths: map[string]bool{}, ObjectKeys: map[string]bool{}, AvatarPaths: map[string]bool{}}
	for _, graph := range graphs {
		for identity := range graph.RetainedStructuredAssetIDs {
			scope.MediaAssets[identity] = true
		}
		for _, reference := range graph.RetainedStructuredFiles {
			scope.ExactPaths[reference.packagePath()] = true
			scope.ObjectKeys[reference.Key] = true
		}
		for _, ticket := range graph.Rows["tickets"] {
			if ticket.referenceColumnText("category") == "customer_service" {
				scope.Tickets[ticket.referenceColumnText("id")] = true
			}
		}
		for _, reference := range graph.LegacyFiles {
			scope.ExactPaths[reference.packagePath()] = true
			scope.ObjectKeys[reference.Key] = true
		}
		for _, asset := range graph.Rows["media_assets"] {
			scope.MediaAssets[asset.referenceColumnText("id")] = true
			root, key := cfg.customerServiceMediaAssetLocalIdentity(asset)
			if key != "" {
				scope.ExactPaths["media/"+root+"/"+key] = true
				scope.ObjectKeys[key] = true
			}
		}
		for _, derivative := range graph.Rows["media_asset_derivatives"] {
			root, key := "uploads", derivative.referenceColumnText("storage_key")
			parent := findCustomerServiceDependencyRowByID(graph, "media_assets", derivative.Values["media_asset_id"])
			if parent != nil && parent.referenceColumnText("visibility") == "private" {
				root = "uploads_private"
			}
			if key == "" {
				root, key = cfg.resolveLocalMediaReference(derivative.referenceColumnText("url"))
			}
			if safePackagePath("media/" + root + "/" + key) {
				scope.ExactPaths["media/"+root+"/"+key] = true
				scope.ObjectKeys[key] = true
			}
		}
		for _, session := range graph.Rows["media_upload_sessions"] {
			key := session.referenceColumnText("object_key")
			if safePackagePath("media/uploads_private/" + key) {
				scope.ExactPaths["media/uploads_private/"+key] = true
				scope.ObjectKeys[key] = true
			}
		}
	}
	return scope
}

func (cfg Configuration) inspectCustomerServiceDurableEventDependencies(ctx context.Context, graphs []*moduleRecoveryRowDependencyGraph, report *ModuleRecoveryRowInspection) error {
	inspection := &ModuleRecoveryDurableEventInspection{Version: 1, ComparisonSampleLimit: report.IssueSampleLimit, SourceStatuses: map[string]int{}, TargetStatuses: map[string]int{}, Comparisons: []ModuleRecoveryDurableEventComparison{}, StatePolicy: "preserve_existing_event_state_no_replay"}
	report.DurableEvents = inspection
	if len(graphs) == 1 && graphs[0].SelectedCustomerServiceRows != nil {
		inspection.CleanupDispositions = &CustomerServiceRecoveryCleanupEventDispositionInspection{Version: 1, Policy: customerServiceCleanupDispositionPolicy, SampleLimit: report.IssueSampleLimit, entries: []CustomerServiceRecoveryCleanupEventDisposition{}}
	}
	scope := cfg.collectCustomerServiceDurableEventRecoveryObjectScope(graphs)
	var snapshots []*moduleRecoveryDurableEventSnapshot
	for _, graph := range graphs {
		snapshot := &moduleRecoveryDurableEventSnapshot{Graph: graph, Events: map[string]*moduleRecoveryDependencyRow{}, Related: map[string]bool{}}
		snapshots = append(snapshots, snapshot)
		valid := true
		for _, column := range strings.Fields("id event_key event_type aggregate_type aggregate_id payload status attempts locked_at locked_by processed_at deleted_at") {
			if !graph.Observation.Columns["outbox_events"][column] {
				valid = false
			}
		}
		if !valid {
			graph.markIncomplete("durable_event_inspection_schema_missing", "outbox_events", "缺少事件身份、载荷或处理状态字段，无法完成幂等和清理对象检查。")
			continue
		}
		predicate := `record.event_type=ANY($1::text[]) OR record.aggregate_type=ANY($2::text[])`
		if graph.SelectedCustomerServiceRows != nil {
			// Event keys remain globally reserved even if another event type has
			// occupied the customer-service namespace, or the row is soft deleted.
			predicate += ` OR record.event_key LIKE 'customer\_service.conversation.status.changed:%' ESCAPE '\' OR record.event_key LIKE 'customer\_service.conversation.assigned:%' ESCAPE '\' OR record.event_key LIKE 'customer\_service.messages.read:%' ESCAPE '\' OR record.event_key LIKE 'customer\_service.conversation.inbox\_state.changed:%' ESCAPE '\'`
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, "outbox_events", predicate, []any{pq.Array([]string{outbox.EventTypeCustomerServiceRealtime, outbox.EventTypeCustomerServiceRetentionCleanup, outbox.EventTypeCustomerServiceAvatarCleanup, outbox.EventTypeObjectStorageCleanup}), pq.Array([]string{outbox.AggregateTypeCustomerServiceConversation, outbox.AggregateTypeCustomerServiceAgentProfile})}, graph.Limits.Rows)
		if err != nil {
			return err
		}
		for _, event := range rows {
			snapshot.Events[event.Identity] = event
		}
		if graph.Side == "package" {
			inspection.SourceScannedEvents = len(rows)
		} else {
			inspection.TargetScannedEvents = len(rows)
		}
	}
	// A receipt can exist on only one side while the current owner exists on
	// the other. All cleanup types can capture a shared avatar object key.
	hasCleanupEvents := false
	for _, snapshot := range snapshots {
		for _, event := range snapshot.Events {
			switch event.referenceColumnText("event_type") {
			case outbox.EventTypeCustomerServiceAvatarCleanup, outbox.EventTypeCustomerServiceRetentionCleanup, outbox.EventTypeObjectStorageCleanup:
				hasCleanupEvents = true
			}
		}
	}
	if hasCleanupEvents {
		for _, snapshot := range snapshots {
			if err := cfg.includePreservedCustomerServiceAvatarReferencePaths(ctx, snapshot, scope); err != nil {
				return err
			}
		}
	}
	for _, snapshot := range snapshots {
		versionIdentities := buildCustomerServiceEffectiveRealtimeVersionIdentityIndex(snapshot.Graph)
		var identities []string
		for identity := range snapshot.Events {
			identities = append(identities, identity)
		}
		sort.Strings(identities)
		for _, identity := range identities {
			event := snapshot.Events[identity]
			related, err := cfg.inspectCustomerServiceDurableEventPayloadReferences(ctx, snapshot.Graph, event, scope)
			if err != nil {
				return err
			}
			if snapshot.Graph.SelectedCustomerServiceRows != nil {
				related = inspectCustomerServicePreservedRealtimeEventVersionReservation(snapshot.Graph, versionIdentities, event) || related
			}
			if related {
				snapshot.Related[identity] = true
				status := moduleRecoveryDurableEventPublicStatus(event)
				if snapshot.Graph.Side == "package" {
					inspection.SourceRelatedEvents++
					inspection.SourceStatuses[status]++
				} else {
					inspection.TargetRelatedEvents++
					inspection.TargetStatuses[status]++
				}
			}
		}
	}
	if len(snapshots) == 2 {
		if err := compareModuleRecoveryDurableEventSnapshots(ctx, snapshots[0], snapshots[1], inspection); err != nil {
			return err
		}
	}
	inspection.Complete = report.Complete
	if inspection.CleanupDispositions != nil {
		if err := finalizeCustomerServiceRecoveryCleanupEventDispositions(inspection.CleanupDispositions, report.Complete && !graphs[0].Exhausted); err != nil {
			return err
		}
	}
	return nil
}

func (cfg Configuration) includePreservedCustomerServiceAvatarReferencePaths(ctx context.Context, snapshot *moduleRecoveryDurableEventSnapshot, scope *customerServiceDurableEventScope) error {
	graph := snapshot.Graph
	if !graph.Observation.Columns["customer_service_agent_profiles"]["avatar"] {
		graph.markIncomplete("avatar_cleanup_owner_schema_missing", "customer_service_agent_profiles", "存在对象清理事件，但缺少客服头像归属字段，无法排除清理命中保留头像。")
		return nil
	}
	// Match the persisted AgentProfile and migration columns. Do not infer a
	// second field from the URL in an avatar-cleanup event payload.
	profiles, err := graph.readBoundedRelationshipRows(ctx, "customer_service_agent_profiles", `record.avatar IS NOT NULL AND btrim(record.avatar)<>''`, nil, graph.Limits.Rows)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		root, key := cfg.resolveLocalMediaReference(profile.referenceColumnText("avatar"))
		if key == "" {
			graph.markIncomplete("avatar_owner_reference_unresolved", "customer_service_agent_profiles", "头像引用无法映射到本地对象，无法证明清理任务与保留配置互不影响。")
			continue
		}
		scope.AvatarPaths["media/"+root+"/"+key] = true
		scope.ObjectKeys[key] = true
	}
	return nil
}
