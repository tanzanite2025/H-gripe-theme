package localbackup

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t *testing.T) *CustomerServiceConversationRecoveryPlan {
	t.Helper()
	plan := &CustomerServiceConversationRecoveryPlan{Version: 1, State: "draft_not_executable", InspectionComplete: true, Binding: CustomerServiceRecoveryPlanSnapshotBinding{TaskID: strings.Repeat("1", 32), SiteID: "synthetic-store", Environment: "test", PackageID: strings.Repeat("2", 32), WorkCopyExpiresAt: time.Now().Add(time.Hour)}, BlockingIssueCounts: map[string]int{"conversation_row_content_differs": 2, "target_dependency_row_missing": 1, "target_only_conversation_row": 1, "conversation_natural_key_owned_by_different_row": 1}, RowActions: []CustomerServiceRecoveryPlanRowAction{
		{Table: "ticket_messages", IdentitySHA256: strings.Repeat("a", 64), SourceSHA256: strings.Repeat("b", 64), TargetSHA256: strings.Repeat("c", 64), Action: "propose_replace_after_review", ReviewableIssueCounts: map[string]int{"conversation_row_content_differs": 1}},
		{Table: "ticket_messages", IdentitySHA256: strings.Repeat("d", 64), SourceSHA256: strings.Repeat("e", 64), Action: "propose_insert_after_review", ReviewableIssueCounts: map[string]int{"target_dependency_row_missing": 1}},
		{Table: "tickets", IdentitySHA256: strings.Repeat("f", 64), SourceSHA256: strings.Repeat("3", 64), TargetSHA256: strings.Repeat("4", 64), Action: "blocked_reference_contract_conflict"},
		{Table: "ticket_messages", IdentitySHA256: strings.Repeat("5", 64), TargetSHA256: strings.Repeat("6", 64), Action: "preserve_target_only", ReviewableIssueCounts: map[string]int{"target_only_conversation_row": 1}},
	}}
	plan.ConflictDecisionPolicy = buildCustomerServiceRecoveryConflictDecisionPolicy(plan)
	var err error
	plan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(plan)
	require.NoError(t, err)
	return plan
}

func createCustomerServiceConflictReviewRequestForTest(plan *CustomerServiceConversationRecoveryPlan) CustomerServiceRecoveryConflictReviewRequest {
	return CustomerServiceRecoveryConflictReviewRequest{PlanSHA256: plan.SHA256, Confirmation: "REVIEW " + plan.Binding.SiteID + "/" + plan.Binding.Environment + "/" + plan.Binding.PackageID, Decisions: []CustomerServiceRecoveryContentDecision{{Table: "ticket_messages", IdentitySHA256: strings.Repeat("a", 64), Choice: "preserve_current_row", Reason: "核对后保留当前客服消息"}}}
}

func TestCustomerServiceConflictReviewSeparatesContentChoicesFromHardIssuesAndKeepsAnImmutablePlan(t *testing.T) {
	plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
	before, err := json.Marshal(plan)
	require.NoError(t, err)
	require.Equal(t, map[string]int{"conversation_row_content_differs": 1, "conversation_natural_key_owned_by_different_row": 1}, plan.ConflictDecisionPolicy.RepairRequiredIssueCounts)
	request := createCustomerServiceConflictReviewRequestForTest(plan)
	now := time.Now().UTC()
	first, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "first-operator", now)
	require.NoError(t, err)
	require.Equal(t, 1, first.PendingRowDecisions)
	require.False(t, first.ContentDecisionsComplete)
	require.False(t, first.ExecutionAvailable)
	request.ExpectedReviewSHA256 = first.SHA256
	request.Decisions = []CustomerServiceRecoveryContentDecision{{Table: "ticket_messages", IdentitySHA256: strings.Repeat("d", 64), Choice: "use_package_row", Reason: "已核对缺失的历史消息"}}
	second, err := buildCustomerServiceRecoveryConflictReview(plan, first, request, "second-operator", now.Add(time.Second))
	require.NoError(t, err)
	require.Zero(t, second.PendingRowDecisions)
	require.True(t, second.ContentDecisionsComplete)
	require.False(t, second.ExecutionAvailable)
	require.Equal(t, first.SHA256, second.PreviousReviewSHA256)
	require.Equal(t, "first-operator", second.Decisions[0].Actor)
	require.Equal(t, "second-operator", second.Decisions[1].Actor)
	require.Equal(t, plan.ConflictDecisionPolicy.RepairRequiredIssueCounts, second.RepairRequiredIssueCounts)
	request.ExpectedReviewSHA256 = second.SHA256
	request.Decisions = []CustomerServiceRecoveryContentDecision{{Table: "ticket_messages", IdentitySHA256: strings.Repeat("a", 64), Choice: "undecided", Reason: "重新核对该条消息"}}
	third, err := buildCustomerServiceRecoveryConflictReview(plan, second, request, "third-operator", now.Add(2*time.Second))
	require.NoError(t, err)
	require.Equal(t, 1, third.PendingRowDecisions)
	require.NotEqual(t, second.SHA256, third.SHA256)
	after, err := json.Marshal(plan)
	require.NoError(t, err)
	require.Equal(t, string(before), string(after))
}

func TestCustomerServiceConflictReviewRejectsStaleIncompleteAmbiguousAndNonwaivableChoices(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*CustomerServiceRecoveryConflictReviewRequest)
	}{
		{"stale_plan", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.PlanSHA256 = strings.Repeat("0", 64)
		}},
		{"missing_confirmation", func(request *CustomerServiceRecoveryConflictReviewRequest) { request.Confirmation = "" }},
		{"concurrent_revision", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.ExpectedReviewSHA256 = strings.Repeat("0", 64)
		}},
		{"unknown_choice", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].Choice = "ignore_all_blockers"
		}},
		{"unknown_row", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].IdentitySHA256 = strings.Repeat("0", 64)
		}},
		{"blocked_identity", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].Table = "tickets"
			request.Decisions[0].IdentitySHA256 = strings.Repeat("f", 64)
		}},
		{"fixed_target_preservation", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].IdentitySHA256 = strings.Repeat("5", 64)
		}},
		{"shared_row", func(request *CustomerServiceRecoveryConflictReviewRequest) { request.Decisions[0].Table = "users" }},
		{"source_only_cannot_preserve_absent_target", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].IdentitySHA256 = strings.Repeat("d", 64)
		}},
		{"duplicate_row", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions = append(request.Decisions, request.Decisions[0])
		}},
		{"empty_reason", func(request *CustomerServiceRecoveryConflictReviewRequest) { request.Decisions[0].Reason = "  " }},
		{"long_multibyte_reason", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].Reason = strings.Repeat("核", 180)
		}},
		{"control_character_reason", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			request.Decisions[0].Reason = "synthetic\nreason"
		}},
		{"no_decisions", func(request *CustomerServiceRecoveryConflictReviewRequest) { request.Decisions = nil }},
		{"unbounded_batch", func(request *CustomerServiceRecoveryConflictReviewRequest) {
			for len(request.Decisions) <= customerServiceConflictReviewMaximumDecisionsPerSave {
				request.Decisions = append(request.Decisions, request.Decisions[0])
			}
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
			request := createCustomerServiceConflictReviewRequestForTest(plan)
			scenario.change(&request)
			_, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "operator", time.Now())
			require.Error(t, err)
		})
	}
	t.Run("changed_plan_body", func(t *testing.T) {
		plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
		request := createCustomerServiceConflictReviewRequestForTest(plan)
		plan.RowActions[0].TargetSHA256 = "changed"
		_, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "operator", time.Now())
		require.Error(t, err)
	})
	t.Run("expired_plan", func(t *testing.T) {
		plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
		request := createCustomerServiceConflictReviewRequestForTest(plan)
		_, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "operator", plan.Binding.WorkCopyExpiresAt)
		require.Error(t, err)
	})
	t.Run("missing_full_counter_mapping", func(t *testing.T) {
		plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
		plan.RowActions[0].ReviewableIssueCounts = nil
		plan.SHA256 = ""
		var err error
		plan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(plan)
		require.NoError(t, err)
		request := createCustomerServiceConflictReviewRequestForTest(plan)
		_, err = buildCustomerServiceRecoveryConflictReview(plan, nil, request, "operator", time.Now())
		require.Error(t, err)
	})
	t.Run("new_plan_does_not_inherit_old_approvals", func(t *testing.T) {
		plan := createCustomerServiceConflictReviewPlanWithContentChoicesAndNonwaivableIdentityIssuesForTest(t)
		request := createCustomerServiceConflictReviewRequestForTest(plan)
		first, err := buildCustomerServiceRecoveryConflictReview(plan, nil, request, "operator", time.Now())
		require.NoError(t, err)
		plan.Binding.TargetIndexFingerprint = "new-index"
		plan.SHA256 = ""
		plan.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(plan)
		require.NoError(t, err)
		request = createCustomerServiceConflictReviewRequestForTest(plan)
		request.ExpectedReviewSHA256 = first.SHA256
		request.Decisions = []CustomerServiceRecoveryContentDecision{{Table: "ticket_messages", IdentitySHA256: strings.Repeat("d", 64), Choice: "use_package_row", Reason: "new plan reviewed"}}
		second, err := buildCustomerServiceRecoveryConflictReview(plan, first, request, "operator", time.Now())
		require.NoError(t, err)
		require.Equal(t, 1, second.PendingRowDecisions)
		require.Len(t, second.Decisions, 1)
	})
}
