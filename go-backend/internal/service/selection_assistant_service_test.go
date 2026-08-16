package service

import (
	"testing"

	selectionassistant "commerce-platform/internal/domain/selectionassistant"
)

func TestValidateSelectionAssistantConfigRejectsUnreachableBranch(t *testing.T) {
	config := selectionassistant.Config{
		Kind:          selectionassistant.ConfigKind,
		SchemaVersion: 1,
		EntryNodeKey:  "start",
		BaseProductQuery: selectionassistant.BaseProductQuery{
			CategorySlug: selectionassistant.WheelsetProductCategorySlug,
		},
		Nodes: []selectionassistant.Node{
			{
				Key:  "start",
				Type: selectionassistant.NodeTypeQuestion,
				Options: []selectionassistant.Option{
					{Key: "yes", NextNodeKey: "finish"},
				},
			},
			{
				Key:  "finish",
				Type: selectionassistant.NodeTypeSupport,
			},
			{
				Key:  "orphan",
				Type: selectionassistant.NodeTypeTerminal,
			},
		},
	}

	_, result := ValidateConfig(config)

	if result.Valid {
		t.Fatal("expected unreachable branch to invalidate the config")
	}
	if !hasSelectionAssistantIssue(result.Issues, "unreachable_node", "orphan") {
		t.Fatalf("expected unreachable_node issue for orphan, got %#v", result.Issues)
	}
}

func TestValidateSelectionAssistantConfigRejectsCycles(t *testing.T) {
	config := selectionassistant.Config{
		Kind:          selectionassistant.ConfigKind,
		SchemaVersion: 1,
		EntryNodeKey:  "start",
		BaseProductQuery: selectionassistant.BaseProductQuery{
			CategorySlug: selectionassistant.WheelsetProductCategorySlug,
		},
		Nodes: []selectionassistant.Node{
			{
				Key:  "start",
				Type: selectionassistant.NodeTypeQuestion,
				Options: []selectionassistant.Option{
					{Key: "next", NextNodeKey: "loop"},
				},
			},
			{
				Key:  "loop",
				Type: selectionassistant.NodeTypeQuestion,
				Options: []selectionassistant.Option{
					{Key: "back", NextNodeKey: "start"},
				},
			},
		},
	}

	_, result := ValidateConfig(config)

	if result.Valid {
		t.Fatal("expected cycle to invalidate the config")
	}
	if !hasSelectionAssistantIssue(result.Issues, "cycle_detected", "start") {
		t.Fatalf("expected cycle_detected issue, got %#v", result.Issues)
	}
}

func hasSelectionAssistantIssue(issues []SelectionAssistantValidationIssue, code, nodeKey string) bool {
	for _, issue := range issues {
		if issue.Code == code && issue.NodeKey == nodeKey {
			return true
		}
	}
	return false
}
