package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const customerServiceConflictReviewCurrentFilename = "customer-service-conflict-review-current-v1.json"
const customerServiceConflictReviewMaximumDecisionsPerSave = 30

type CustomerServiceRecoveryConflictDecisionPolicy struct {
	Version                      int            `json:"version"`
	MappingComplete              bool           `json:"mapping_complete"`
	RequiredRowDecisions         int            `json:"required_row_decisions"`
	ReviewableIssueCounts        map[string]int `json:"reviewable_issue_counts"`
	FixedPreservationIssueCounts map[string]int `json:"fixed_preservation_issue_counts"`
	RepairRequiredIssueCounts    map[string]int `json:"repair_required_issue_counts"`
}

type CustomerServiceRecoveryContentDecision struct {
	Table          string `json:"table"`
	IdentitySHA256 string `json:"identity_sha256"`
	Choice         string `json:"choice"`
	Reason         string `json:"reason"`
}

type CustomerServiceRecoveryRecordedContentDecision struct {
	CustomerServiceRecoveryContentDecision
	SourceSHA256 string    `json:"source_sha256"`
	TargetSHA256 string    `json:"target_sha256,omitempty"`
	Actor        string    `json:"actor"`
	RecordedAt   time.Time `json:"recorded_at"`
}

type CustomerServiceRecoveryConflictReviewRequest struct {
	PlanSHA256           string                                   `json:"plan_sha256"`
	ExpectedReviewSHA256 string                                   `json:"expected_review_sha256"`
	Confirmation         string                                   `json:"confirmation"`
	Decisions            []CustomerServiceRecoveryContentDecision `json:"decisions"`
}

type CustomerServiceRecoveryConflictReview struct {
	Version                   int                                              `json:"version"`
	State                     string                                           `json:"state"`
	ExecutionAvailable        bool                                             `json:"execution_available"`
	TaskID                    string                                           `json:"task_id"`
	PlanSHA256                string                                           `json:"plan_sha256"`
	Binding                   CustomerServiceRecoveryPlanSnapshotBinding       `json:"binding"`
	SHA256                    string                                           `json:"sha256"`
	PreviousReviewSHA256      string                                           `json:"previous_review_sha256,omitempty"`
	Actor                     string                                           `json:"actor"`
	RecordedAt                time.Time                                        `json:"recorded_at"`
	Decisions                 []CustomerServiceRecoveryRecordedContentDecision `json:"decisions"`
	RequiredRowDecisions      int                                              `json:"required_row_decisions"`
	PendingRowDecisions       int                                              `json:"pending_row_decisions"`
	ContentDecisionsComplete  bool                                             `json:"content_decisions_complete"`
	ChoiceCounts              map[string]int                                   `json:"choice_counts"`
	RepairRequiredIssueCounts map[string]int                                   `json:"repair_required_issue_counts"`
}

func (graph *moduleRecoveryRowDependencyGraph) recordCustomerServiceDecisionIssue(table, identity, code string) {
	if graph.CustomerServiceDecisionIssueCounts == nil {
		graph.CustomerServiceDecisionIssueCounts = map[string]map[string]map[string]int{}
	}
	if graph.CustomerServiceDecisionIssueCounts[table] == nil {
		graph.CustomerServiceDecisionIssueCounts[table] = map[string]map[string]int{}
	}
	if graph.CustomerServiceDecisionIssueCounts[table][identity] == nil {
		graph.CustomerServiceDecisionIssueCounts[table][identity] = map[string]int{}
	}
	graph.CustomerServiceDecisionIssueCounts[table][identity][code]++
}

func customerServiceRecoveryRowRequiresContentDecision(row CustomerServiceRecoveryPlanRowAction) bool {
	return customerServiceRecoveryPlanProposedRowReplacementAllowed(row.Table) && (row.Action == "propose_insert_after_review" || row.Action == "propose_replace_after_review")
}

// Classification uses the full per-row comparison counters, never the truncated
// issue samples. Anything not explicitly assigned remains repair-required.
func buildCustomerServiceRecoveryConflictDecisionPolicy(plan *CustomerServiceConversationRecoveryPlan) *CustomerServiceRecoveryConflictDecisionPolicy {
	policy := &CustomerServiceRecoveryConflictDecisionPolicy{Version: 1, MappingComplete: true, ReviewableIssueCounts: map[string]int{}, FixedPreservationIssueCounts: map[string]int{}, RepairRequiredIssueCounts: map[string]int{}}
	for code, count := range plan.BlockingIssueCounts {
		policy.RepairRequiredIssueCounts[code] = count
	}
	for _, row := range plan.RowActions {
		if customerServiceRecoveryRowRequiresContentDecision(row) {
			policy.RequiredRowDecisions++
			expectedCode := "conversation_row_content_differs"
			if row.Action == "propose_insert_after_review" {
				expectedCode = "target_dependency_row_missing"
			}
			if len(row.ReviewableIssueCounts) != 1 || row.ReviewableIssueCounts[expectedCode] != 1 {
				policy.MappingComplete = false
				continue
			}
			policy.ReviewableIssueCounts[expectedCode]++
		} else if row.Action == "preserve_target_only" && customerServiceRecoveryPlanProposedRowReplacementAllowed(row.Table) {
			for code, count := range row.ReviewableIssueCounts {
				if code != "target_only_conversation_row" || count != 1 {
					policy.MappingComplete = false
					continue
				}
				policy.FixedPreservationIssueCounts[code] += count
			}
		}
	}
	for _, counts := range []map[string]int{policy.ReviewableIssueCounts, policy.FixedPreservationIssueCounts} {
		for code, count := range counts {
			if policy.RepairRequiredIssueCounts[code] < count {
				policy.MappingComplete = false
				continue
			}
			policy.RepairRequiredIssueCounts[code] -= count
			if policy.RepairRequiredIssueCounts[code] == 0 {
				delete(policy.RepairRequiredIssueCounts, code)
			}
		}
	}
	return policy
}

func customerServiceRecoveryContentDecisionKey(table, identity string) string {
	return table + ":" + identity
}

func buildCustomerServiceRecoveryConflictReview(plan *CustomerServiceConversationRecoveryPlan, previous *CustomerServiceRecoveryConflictReview, request CustomerServiceRecoveryConflictReviewRequest, actor string, now time.Time) (*CustomerServiceRecoveryConflictReview, error) {
	if plan == nil {
		return nil, errors.New("customer service recovery plan is missing")
	}
	if err := verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(plan, plan.Binding, now); err != nil {
		return nil, err
	}
	policy := buildCustomerServiceRecoveryConflictDecisionPolicy(plan)
	if plan.ConflictDecisionPolicy == nil || plan.ConflictDecisionPolicy.Version != 1 || !policy.MappingComplete {
		return nil, errors.New("conflict decision mapping is missing or incomplete; inspect again")
	}
	if request.PlanSHA256 != plan.SHA256 {
		return nil, errors.New("recovery plan changed; inspect again and review the new plan")
	}
	if strings.TrimSpace(actor) == "" || request.Confirmation != fmt.Sprintf("REVIEW %s/%s/%s", plan.Binding.SiteID, plan.Binding.Environment, plan.Binding.PackageID) {
		return nil, errors.New("explicit content-review confirmation is required")
	}
	if len(request.Decisions) == 0 || len(request.Decisions) > customerServiceConflictReviewMaximumDecisionsPerSave {
		return nil, errors.New("save between 1 and 30 row decisions at a time")
	}
	previousDigest := ""
	if previous != nil {
		previousDigest = previous.SHA256
	}
	if request.ExpectedReviewSHA256 != previousDigest {
		return nil, errors.New("another review revision was saved; reload before changing decisions")
	}
	allowed := map[string]CustomerServiceRecoveryPlanRowAction{}
	for _, row := range plan.RowActions {
		if customerServiceRecoveryRowRequiresContentDecision(row) {
			allowed[customerServiceRecoveryContentDecisionKey(row.Table, row.IdentitySHA256)] = row
		}
	}
	decisions := map[string]CustomerServiceRecoveryRecordedContentDecision{}
	if previous != nil && previous.PlanSHA256 == plan.SHA256 {
		for _, decision := range previous.Decisions {
			key := customerServiceRecoveryContentDecisionKey(decision.Table, decision.IdentitySHA256)
			row, exists := allowed[key]
			if !exists || row.SourceSHA256 != decision.SourceSHA256 || row.TargetSHA256 != decision.TargetSHA256 {
				return nil, errors.New("stored content decision does not match the plan")
			}
			decisions[key] = decision
		}
	}
	seen := map[string]bool{}
	for _, decision := range request.Decisions {
		key := customerServiceRecoveryContentDecisionKey(decision.Table, decision.IdentitySHA256)
		row, exists := allowed[key]
		if !exists || seen[key] {
			return nil, errors.New("decision targets an unknown, duplicate, preserved or blocked row")
		}
		seen[key] = true
		decision.Reason = strings.TrimSpace(decision.Reason)
		if decision.Reason == "" || len(decision.Reason) > 512 || strings.IndexFunc(decision.Reason, unicode.IsControl) >= 0 {
			return nil, errors.New("each decision requires a reason of at most 512 UTF-8 bytes without control characters")
		}
		switch decision.Choice {
		case "use_package_row":
		case "preserve_current_row":
			if row.Action != "propose_replace_after_review" {
				return nil, errors.New("a source-only row has no current row to preserve")
			}
		case "undecided":
			delete(decisions, key)
			continue
		default:
			return nil, errors.New("unsupported content decision; identity or reference conflicts cannot be waived")
		}
		decisions[key] = CustomerServiceRecoveryRecordedContentDecision{CustomerServiceRecoveryContentDecision: decision, SourceSHA256: row.SourceSHA256, TargetSHA256: row.TargetSHA256, Actor: actor, RecordedAt: now}
	}
	review := &CustomerServiceRecoveryConflictReview{Version: 1, State: "recorded_not_executable", TaskID: plan.Binding.TaskID, PlanSHA256: plan.SHA256, Binding: plan.Binding, PreviousReviewSHA256: previousDigest, Actor: actor, RecordedAt: now, Decisions: []CustomerServiceRecoveryRecordedContentDecision{}, RequiredRowDecisions: len(allowed), ChoiceCounts: map[string]int{}, RepairRequiredIssueCounts: policy.RepairRequiredIssueCounts}
	for _, decision := range decisions {
		review.Decisions = append(review.Decisions, decision)
		review.ChoiceCounts[decision.Choice]++
	}
	sort.Slice(review.Decisions, func(i, j int) bool {
		return customerServiceRecoveryContentDecisionKey(review.Decisions[i].Table, review.Decisions[i].IdentitySHA256) < customerServiceRecoveryContentDecisionKey(review.Decisions[j].Table, review.Decisions[j].IdentitySHA256)
	})
	review.PendingRowDecisions = review.RequiredRowDecisions - len(review.Decisions)
	review.ContentDecisionsComplete = review.PendingRowDecisions == 0
	var err error
	review.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(review)
	return review, err
}

func (manager *Manager) loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(id string) (*CustomerServiceRecoveryConflictReview, error) {
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filepath.Join(directory, customerServiceConflictReviewCurrentFilename))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 32<<20+1))
	if err != nil || len(data) > 32<<20 {
		return nil, errors.New("stored conflict review exceeds its size limit")
	}
	var review CustomerServiceRecoveryConflictReview
	if err := decodeStrictJSON(data, &review); err != nil {
		return nil, errors.New("stored conflict review is invalid")
	}
	copyOfReview := review
	copyOfReview.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfReview)
	if err != nil || digest != review.SHA256 || review.Version != 1 || review.State != "recorded_not_executable" || review.ExecutionAvailable || review.TaskID != id {
		return nil, errors.New("stored conflict review integrity or task identity changed")
	}
	return &review, nil
}

func (manager *Manager) ReadCustomerServiceRecoveryConflictReview(id string) (*CustomerServiceRecoveryConflictReview, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	if _, err := manager.GetTask(id); err != nil {
		return nil, err
	}
	return manager.loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(id)
}

// Every save observes the current source, target, schema, indexes and media.
// These observations remain unfenced: this records choices, never write authority.
func (manager *Manager) SaveCustomerServiceRecoveryConflictReview(ctx context.Context, id, actor string, request CustomerServiceRecoveryConflictReviewRequest) (*CustomerServiceRecoveryConflictReview, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	previous, err := manager.loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(id)
	if err != nil {
		return nil, err
	}
	report, err := manager.inspectVerifiedPackageModuleDependenciesWhileOperationLocked(ctx, id, "customer_service")
	if err != nil {
		return nil, err
	}
	plan := report.ConversationRows.ConversationRecoveryPlan
	review, err := buildCustomerServiceRecoveryConflictReview(plan, previous, request, actor, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	directory, _ := manager.taskDirectory(id)
	encodedReview, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(encodedReview) > 32<<20 {
		return nil, errors.New("conflict review exceeds its size limit")
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(encodedReview)) * 2); err != nil {
		return nil, err
	}
	// Preserve each revision before replacing the current pointer. A failed
	// pointer write leaves an unused revision and cannot advance the review.
	if err := writeDurableFile(filepath.Join(directory, "customer-service-conflict-review-revision-"+review.SHA256+".json"), encodedReview); err != nil {
		return nil, errors.New("conflict review revision could not be persisted")
	}
	if err := writeDurableFile(filepath.Join(directory, customerServiceConflictReviewCurrentFilename), encodedReview); err != nil {
		return nil, errors.New("current conflict review could not be persisted")
	}
	return review, nil
}
