package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lib/pq"
)

// A draft describes proposed changes, never permission to execute them. Target
// fingerprints are observations made without a write fence; the future builder
// must freeze writers and revalidate every binding before cloning the target.
type CustomerServiceConversationRecoveryPlan struct {
	Version                 int                                                                  `json:"version"`
	State                   string                                                               `json:"state"`
	ExecutionAvailable      bool                                                                 `json:"execution_available"`
	Scope                   string                                                               `json:"scope"`
	Binding                 CustomerServiceRecoveryPlanSnapshotBinding                           `json:"binding"`
	SHA256                  string                                                               `json:"sha256"`
	InspectionComplete      bool                                                                 `json:"inspection_complete"`
	ReferenceChecks         *CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection `json:"reference_checks"`
	BlockingIssueCounts     map[string]int                                                       `json:"blocking_issue_counts"`
	DependencyBlockers      []ModuleRecoveryDependencyBlocker                                    `json:"dependency_blockers"`
	RowActions              []CustomerServiceRecoveryPlanRowAction                               `json:"row_actions"`
	RowActionCounts         map[string]int                                                       `json:"row_action_counts"`
	FileActions             []CustomerServiceRecoveryPlanFileAction                              `json:"file_actions"`
	FileActionCounts        map[string]int                                                       `json:"file_action_counts"`
	TargetPreservation      []CustomerServiceRecoveryPlanTablePreservation                       `json:"target_preservation"`
	TargetMediaPolicy       string                                                               `json:"target_media_policy"`
	DurableStatePolicy      string                                                               `json:"durable_state_policy"`
	UnimplementedSafeguards []string                                                             `json:"unimplemented_safeguards"`
	ConflictDecisionPolicy  *CustomerServiceRecoveryConflictDecisionPolicy                       `json:"conflict_decision_policy,omitempty"`
}

type CustomerServiceRecoveryPlanSnapshotBinding struct {
	TaskID                                    string    `json:"task_id"`
	PackageID                                 string    `json:"package_id"`
	SiteID                                    string    `json:"site_id"`
	Environment                               string    `json:"environment"`
	ApplicationVersion                        string    `json:"application_version"`
	MigrationVersion                          int64     `json:"migration_version"`
	RegistryVersion                           int       `json:"registry_version"`
	RegistryMigrationVersion                  int64     `json:"registry_migration_version"`
	ManifestSHA256                            string    `json:"manifest_sha256"`
	CiphertextSHA256                          string    `json:"ciphertext_sha256"`
	SourceDatabaseFingerprint                 string    `json:"source_database_fingerprint"`
	TargetDatabaseFingerprint                 string    `json:"target_database_fingerprint"`
	SourceColumnConstraintFingerprint         string    `json:"source_column_constraint_fingerprint"`
	TargetColumnConstraintFingerprint         string    `json:"target_column_constraint_fingerprint"`
	SourceIndexFingerprint                    string    `json:"source_index_fingerprint"`
	TargetIndexFingerprint                    string    `json:"target_index_fingerprint"`
	SourceWriteSideEffectFingerprint          string    `json:"source_write_side_effect_fingerprint"`
	TargetWriteSideEffectFingerprint          string    `json:"target_write_side_effect_fingerprint"`
	TargetMediaFingerprint                    string    `json:"target_media_fingerprint"`
	OriginalWholeStorefrontPreviewFingerprint string    `json:"original_whole_storefront_preview_fingerprint"`
	TargetBoundarySHA256                      string    `json:"target_boundary_sha256"`
	WorkCopyExpiresAt                         time.Time `json:"work_copy_expires_at"`
}

type CustomerServiceRecoveryPlanRowAction struct {
	Table                 string         `json:"table"`
	IdentitySHA256        string         `json:"identity_sha256"`
	SourceSHA256          string         `json:"source_sha256,omitempty"`
	TargetSHA256          string         `json:"target_sha256,omitempty"`
	Action                string         `json:"action"`
	RowIdentity           string         `json:"row_identity,omitempty"`
	ReviewableIssueCounts map[string]int `json:"reviewable_issue_counts,omitempty"`
}

type CustomerServiceRecoveryPlanFileAction struct {
	PathSHA256   string `json:"path_sha256"`
	SourceSHA256 string `json:"source_sha256,omitempty"`
	TargetSHA256 string `json:"target_sha256,omitempty"`
	SourceSize   int64  `json:"source_size"`
	TargetSize   int64  `json:"target_size"`
	Action       string `json:"action"`
}

type CustomerServiceRecoveryPlanTablePreservation struct {
	Table                      string `json:"table"`
	RetainedRecords            int64  `json:"retained_records"`
	ProposedReplacementRecords int    `json:"proposed_replacement_records"`
	ContentFingerprint         string `json:"content_fingerprint"`
}

func customerServiceRecoveryPlanCanonicalJSONSHA256(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func customerServiceRecoveryPlanProposedRowReplacementAllowed(table string) bool {
	// Upload receipts, global events, media and configuration remain shared even
	// if the table-level dependency registry describes them as module-owned.
	switch table {
	case "tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments":
		return true
	}
	return false
}

func buildCustomerServiceConversationRecoveryPlanDraft(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph) (*CustomerServiceConversationRecoveryPlan, error) {
	plan := &CustomerServiceConversationRecoveryPlan{
		Version: 1, State: "draft_not_executable", Scope: "all_customer_service_conversations_including_soft_deleted",
		BlockingIssueCounts: source.Report.BlockingIssueCounts, InspectionComplete: source.Report.Complete,
		ReferenceChecks: source.Report.ConversationReferenceChecks,
		RowActions:      []CustomerServiceRecoveryPlanRowAction{}, RowActionCounts: map[string]int{},
		FileActions: []CustomerServiceRecoveryPlanFileAction{}, FileActionCounts: map[string]int{},
		TargetPreservation: []CustomerServiceRecoveryPlanTablePreservation{},
		TargetMediaPolicy:  "clone_current_complete_media_then_add_approved_source_files_no_deletion",
		DurableStatePolicy: "preserve_all_target_events_receipts_configuration_and_unselected_rows_no_replay",
		UnimplementedSafeguards: []string{
			"unregistered_business_snapshot_and_custom_html_reference_contracts",
			"custom_database_write_side_effect_contracts_remain_rejected",
			"conflicting_cleanup_receipt_cancellation_or_replacement_contracts_remain_unsupported",
			"external_projection_reconciliation_before_production_release",
			"production_module_restore_eligibility_and_final_overwrite_confirmation",
			"production_partial_restore_acceptance_reconciliation_and_release",
			"production_scale_capacity_and_remaining_crash_boundary_validation",
		},
	}
	if plan.BlockingIssueCounts == nil {
		plan.BlockingIssueCounts = map[string]int{}
	}
	replacements := map[string][]json.RawMessage{}
	identities := map[string]map[string]bool{}
	for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
		if graph == nil {
			continue
		}
		for table, rows := range graph.Rows {
			if identities[table] == nil {
				identities[table] = map[string]bool{}
			}
			for identity := range rows {
				identities[table][identity] = true
			}
		}
	}
	for table, keys := range identities {
		for identity := range keys {
			original := source.Rows[table][identity]
			var current *moduleRecoveryDependencyRow
			if target != nil {
				current = target.Rows[table][identity]
				if matched := target.ComparedTargetRows[table][identity]; matched != nil {
					current = matched
				}
			}
			keyHash, err := customerServiceRecoveryPlanCanonicalJSONSHA256(json.RawMessage(identity))
			if err != nil {
				return nil, err
			}
			action := CustomerServiceRecoveryPlanRowAction{Table: table, IdentitySHA256: keyHash}
			identityRow, identityObservation := original, source.Observation
			if original == nil && target != nil {
				identityRow, identityObservation = current, target.Observation
			}
			if identityRow != nil {
				action.RowIdentity = moduleRecoveryDependencyRowPublicIdentity(identityRow, identityObservation.PrimaryKeys[table])
			}
			if original != nil {
				action.SourceSHA256 = original.Digest
			}
			if current != nil {
				action.TargetSHA256 = current.Digest
			}
			switch {
			case original == nil:
				action.Action = "preserve_target_only"
			case target == nil || source.Exhausted || target.Exhausted || source.Observation.Shapes[table] != target.Observation.Shapes[table]:
				action.Action = "blocked_comparison_incomplete"
			case source.BlockedCandidateRows[table][identity] || target.BlockedCandidateRows[table][identity]:
				action.Action = "blocked_reference_contract_conflict"
			case current != nil && customerServiceConversationRowOwnershipConflicts(table, original, current):
				action.Action = "blocked_ownership_conflict"
			case current != nil && original.Digest == current.Digest:
				action.Action = "preserve_matching_target"
			case !source.Scoped[table][identity] || !customerServiceRecoveryPlanProposedRowReplacementAllowed(table):
				if current == nil {
					action.Action = "blocked_shared_dependency_missing"
				} else {
					action.Action = "blocked_shared_dependency_differs"
				}
			case current == nil:
				action.Action = "propose_insert_after_review"
			default:
				action.Action = "propose_replace_after_review"
				replacements[table] = append(replacements[table], json.RawMessage(identity))
			}
			if customerServiceRecoveryPlanProposedRowReplacementAllowed(table) && (action.Action == "propose_insert_after_review" || action.Action == "propose_replace_after_review" || action.Action == "preserve_target_only") {
				action.ReviewableIssueCounts = source.CustomerServiceDecisionIssueCounts[table][identity]
			}
			plan.RowActions = append(plan.RowActions, action)
			plan.RowActionCounts[action.Action]++
		}
	}
	sort.Slice(plan.RowActions, func(i, j int) bool {
		a, b := plan.RowActions[i], plan.RowActions[j]
		return a.Table+":"+a.IdentitySHA256 < b.Table+":"+b.IdentitySHA256
	})
	paths := map[string]bool{}
	for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
		if graph != nil {
			for path := range graph.InspectedFiles {
				paths[path] = true
			}
		}
	}
	for path := range paths {
		original, inSource := source.InspectedFiles[path]
		var current moduleRecoveryInspectedFileSnapshot
		if target != nil {
			current = target.InspectedFiles[path]
		}
		hash := sha256.Sum256([]byte(path))
		action := CustomerServiceRecoveryPlanFileAction{PathSHA256: hex.EncodeToString(hash[:]), SourceSHA256: original.Identity.SHA256, TargetSHA256: current.Identity.SHA256, SourceSize: original.Identity.Size, TargetSize: current.Identity.Size}
		switch {
		case !inSource && current.Available:
			action.Action = "preserve_target_only"
		case !inSource || !original.Available || target == nil:
			action.Action = "blocked_file_observation_incomplete"
		case current.Available && original.Identity.SHA256 == current.Identity.SHA256 && original.Identity.Size == current.Identity.Size:
			action.Action = "preserve_matching_target"
		case current.Missing:
			action.Action = "propose_add_source_file_after_review"
		case current.Available || current.ContentConflict:
			action.Action = "blocked_same_path_content_conflict"
		default:
			action.Action = "blocked_file_observation_incomplete"
		}
		plan.FileActions = append(plan.FileActions, action)
		plan.FileActionCounts[action.Action]++
	}
	sort.Slice(plan.FileActions, func(i, j int) bool { return plan.FileActions[i].PathSHA256 < plan.FileActions[j].PathSHA256 })
	if target != nil {
		var tables []string
		for table := range target.Observation.Tables {
			tables = append(tables, table)
		}
		sort.Strings(tables)
		for _, table := range tables {
			keys := replacements[table]
			sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
			preserved, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, target.Transaction, table, target.Observation.Shapes[table], keys)
			if err != nil {
				return nil, err
			}
			plan.TargetPreservation = append(plan.TargetPreservation, preserved)
		}
	}
	plan.ConflictDecisionPolicy = buildCustomerServiceRecoveryConflictDecisionPolicy(plan)
	return plan, nil
}

// The same aggregate format as table inspection is used for consistency
// comparison. It is not a cryptographic authentication of individual records.
func fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx context.Context, transaction *sql.Tx, table, shape string, excludedPrimaryKeys []json.RawMessage) (CustomerServiceRecoveryPlanTablePreservation, error) {
	result := CustomerServiceRecoveryPlanTablePreservation{Table: table, ProposedReplacementRecords: len(excludedPrimaryKeys)}
	if excludedPrimaryKeys == nil {
		excludedPrimaryKeys = []json.RawMessage{}
	}
	encoded, err := json.Marshal(excludedPrimaryKeys)
	if err != nil {
		return result, err
	}
	query := `SELECT count(*),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),1,15))::bit(60)::bigint)::text,'0'),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),17,15))::bit(60)::bigint)::text,'0') FROM public.` + pq.QuoteIdentifier(table) + ` record WHERE NOT EXISTS(SELECT 1 FROM jsonb_array_elements($1::jsonb) excluded(key) WHERE to_jsonb(record) @> excluded.key)`
	var first, second string
	if err := transaction.QueryRowContext(ctx, query, string(encoded)).Scan(&result.RetainedRecords, &first, &second); err != nil {
		return result, err
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s:%s", shape, result.RetainedRecords, first, second)))
	result.ContentFingerprint = hex.EncodeToString(digest[:])
	return result, nil
}

func (cfg Configuration) bindCustomerServiceRecoveryPlanToObservedSnapshots(plan *CustomerServiceConversationRecoveryPlan, task *Task, report *ModuleRecoveryDependencyReport, mediaFingerprint string, sourceObservation, targetObservation *moduleRecoveryDatabaseObservation) error {
	if !time.Now().Before(task.ExpiresAt) {
		return errors.New("work copy expired while observing the recovery plan draft")
	}
	manifestDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(task.Manifest)
	if err != nil {
		return err
	}
	sourceSchemaDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(sourceObservation.Shapes)
	if err != nil {
		return err
	}
	targetSchemaDigest := "absent_database"
	if targetObservation != nil {
		targetSchemaDigest, err = customerServiceRecoveryPlanCanonicalJSONSHA256(targetObservation.Shapes)
		if err != nil {
			return err
		}
	}
	sourceIndexDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(sourceObservation.Indexes)
	if err != nil {
		return err
	}
	targetIndexDigest := "absent_database"
	if targetObservation != nil {
		targetIndexDigest, err = customerServiceRecoveryPlanCanonicalJSONSHA256(targetObservation.Indexes)
		if err != nil {
			return err
		}
	}
	sourceWriteSideEffectDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(sourceObservation.WriteSideEffects)
	if err != nil {
		return err
	}
	targetWriteSideEffectDigest := "absent_database"
	if targetObservation != nil {
		targetWriteSideEffectDigest, err = customerServiceRecoveryPlanCanonicalJSONSHA256(targetObservation.WriteSideEffects)
		if err != nil {
			return err
		}
	}
	origins := append([]string{}, cfg.PublicOrigins...)
	sort.Strings(origins)
	boundaryDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(struct {
		DatabaseName, DatabaseURL, ApplicationDatabaseRole, StageDatabase, Workspace, MaintenanceDirectory string
		StorageRoots                                                                                       map[string]string
		PublicOrigins                                                                                      []string
	}{cfg.DatabaseName, cfg.DatabaseURL, cfg.ApplicationDatabaseRole, task.StageDatabase, cfg.Workspace, cfg.MaintenanceDirectory, cfg.StorageRoots, origins})
	if err != nil {
		return err
	}
	plan.Binding = CustomerServiceRecoveryPlanSnapshotBinding{
		TaskID: task.ID, PackageID: task.Manifest.PackageID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: cfg.ExpectedMigrationVersion,
		RegistryVersion: moduleDependencyRegistryVersion, RegistryMigrationVersion: moduleDependencyRegistryMigrationVersion, ManifestSHA256: manifestDigest, CiphertextSHA256: task.CiphertextSHA256,
		SourceDatabaseFingerprint: report.SourceDatabaseFingerprint, TargetDatabaseFingerprint: report.TargetDatabaseFingerprint, TargetMediaFingerprint: mediaFingerprint,
		SourceColumnConstraintFingerprint: sourceSchemaDigest, TargetColumnConstraintFingerprint: targetSchemaDigest,
		SourceIndexFingerprint: sourceIndexDigest, TargetIndexFingerprint: targetIndexDigest,
		SourceWriteSideEffectFingerprint: sourceWriteSideEffectDigest, TargetWriteSideEffectFingerprint: targetWriteSideEffectDigest,
		OriginalWholeStorefrontPreviewFingerprint: task.PreviewTargetFingerprint, TargetBoundarySHA256: boundaryDigest, WorkCopyExpiresAt: task.ExpiresAt,
	}
	plan.DependencyBlockers = append([]ModuleRecoveryDependencyBlocker{}, report.Blockers...)
	plan.SHA256 = ""
	plan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(plan)
	return err
}

// This checks draft integrity/freshness only. It never authorizes a write and
// must not be used as an execution eligibility check by a future candidate API.
func verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(plan *CustomerServiceConversationRecoveryPlan, current CustomerServiceRecoveryPlanSnapshotBinding, now time.Time) error {
	if plan == nil || plan.Version != 1 || plan.State != "draft_not_executable" || plan.ExecutionAvailable || !now.Before(plan.Binding.WorkCopyExpiresAt) {
		return errors.New("recovery plan draft is invalid or expired")
	}
	copyOfPlan := *plan
	copyOfPlan.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfPlan)
	if err != nil || digest != plan.SHA256 {
		return errors.New("recovery plan draft digest changed")
	}
	expected, err := customerServiceRecoveryPlanCanonicalJSONSHA256(current)
	if err != nil {
		return err
	}
	bound, err := customerServiceRecoveryPlanCanonicalJSONSHA256(plan.Binding)
	if err != nil || expected != bound {
		return errors.New("recovery plan observed snapshot binding changed")
	}
	return nil
}
