package localbackup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomerServiceEffectiveSelectionRejectsIncompleteStaleAndSemanticallyCorruptedDecisions(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*CustomerServiceRecoveryConflictReview, *CustomerServiceEffectiveSelectionInspectionRequest)
	}{
		{"stale_plan", func(_ *CustomerServiceRecoveryConflictReview, request *CustomerServiceEffectiveSelectionInspectionRequest) {
			request.PlanSHA256 = "stale"
		}},
		{"stale_review", func(_ *CustomerServiceRecoveryConflictReview, request *CustomerServiceEffectiveSelectionInspectionRequest) {
			request.ReviewSHA256 = "stale"
		}},
		{"pending_rows", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.PendingRowDecisions = 1
		}},
		{"incomplete_flag", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.ContentDecisionsComplete = false
		}},
		{"missing_row", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions = review.Decisions[:1]
		}},
		{"duplicate_row", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[1] = review.Decisions[0]
		}},
		{"source_only_preserve", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[1].Choice = "preserve_current_row"
		}},
		{"unsupported_choice", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[0].Choice = "ignore_all_blockers"
		}},
		{"wrong_source_digest", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[0].SourceSHA256 = "changed"
		}},
		{"wrong_target_digest", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[0].TargetSHA256 = "changed"
		}},
		{"wrong_row", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Decisions[0].Table = "users"
		}},
		{"wrong_binding", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Binding.TargetMediaFingerprint = "changed"
		}},
		{"wrong_task", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.TaskID = "other"
		}},
		{"wrong_version", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.Version = 2
		}},
		{"execution_flag", func(review *CustomerServiceRecoveryConflictReview, _ *CustomerServiceEffectiveSelectionInspectionRequest) {
			review.ExecutionAvailable = true
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
			request := createCustomerServiceConflictReviewRequestForTest(plan)
			request.Decisions = append(request.Decisions, CustomerServiceRecoveryContentDecision{Table: plan.RowActions[1].Table, IdentitySHA256: plan.RowActions[1].IdentitySHA256, Choice: "use_package_row", Reason: "verified missing message"})
			review, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "synthetic-operator", time.Now().UTC())
			require.NoError(t, err)
			selectionRequest := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256, ReviewSHA256: review.SHA256}
			decisions, err := validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, review, selectionRequest, time.Now().UTC())
			require.NoError(t, err)
			require.Len(t, decisions, 2)
			scenario.change(review, &selectionRequest)
			// A matching JSON digest alone must not hide invalid selection semantics.
			if scenario.name != "stale_plan" && scenario.name != "stale_review" {
				review.SHA256 = ""
				review.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(review)
				require.NoError(t, err)
				selectionRequest.ReviewSHA256 = review.SHA256
			}
			_, err = validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, review, selectionRequest, time.Now().UTC())
			require.Error(t, err)
		})
	}
}

func TestCustomerServiceEffectiveSelectionNeedsNoOldReviewWhenNewPlanHasNoContentDifferences(t *testing.T) {
	oldPlan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
	oldRequest := createCustomerServiceConflictReviewRequestForTest(oldPlan)
	oldReview, err := buildCustomerServiceRecoveryConflictReview(oldPlan, nil, oldRequest, "synthetic-operator", time.Now().UTC())
	require.NoError(t, err)
	newPlan := *oldPlan
	newPlan.RowActions = nil
	newPlan.BlockingIssueCounts = map[string]int{}
	newPlan.ConflictDecisionPolicy = buildCustomerServiceRecoveryConflictDecisionPolicy(&newPlan)
	newPlan.SHA256 = ""
	newPlan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(&newPlan)
	require.NoError(t, err)
	request := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: newPlan.SHA256}
	decisions, err := validateCustomerServiceSavedDecisionsAgainstLatestPlan(&newPlan, oldReview, request, time.Now().UTC())
	require.NoError(t, err)
	require.Empty(t, decisions)
	request.ReviewSHA256 = oldReview.SHA256
	_, err = validateCustomerServiceSavedDecisionsAgainstLatestPlan(&newPlan, oldReview, request, time.Now().UTC())
	require.Error(t, err)
}
