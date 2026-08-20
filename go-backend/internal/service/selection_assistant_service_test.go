package service

import (
	"testing"

	selectionassistant "commerce-platform/internal/domain/selectionassistant"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func TestSelectionAssistantServiceSkipsReservedWheelsetFitFlow(t *testing.T) {
	db, service := newSelectionAssistantTestService(t)

	seedSelectionAssistantFlow(t, db, "wheelset-fit-helper", 100)
	seedSelectionAssistantFlow(t, db, "frame-helper", 200)

	flows, err := service.ListFlows()
	require.NoError(t, err)
	require.Len(t, flows, 1)
	assert.Equal(t, "frame-helper", flows[0].Slug)

	_, err = service.GetPublishedFlowBySlug("wheelset-fit-helper")
	require.ErrorIs(t, err, ErrSelectionAssistantNotFound)

	_, err = service.CreateFlow(SelectionAssistantFlowInput{
		Slug:                "wheelset-fit-helper",
		Name:                "Wheelset fit helper",
		ProductCategorySlug: selectionassistant.WheelsetProductCategorySlug,
		Version: SelectionAssistantVersionInput{
			Config: selectionassistant.Config{
				Kind:          selectionassistant.ConfigKind,
				SchemaVersion: 1,
				EntryNodeKey:  "start",
				BaseProductQuery: selectionassistant.BaseProductQuery{
					CategorySlug: selectionassistant.WheelsetProductCategorySlug,
				},
			},
		},
	})
	require.ErrorIs(t, err, ErrSelectionAssistantInvalid)
}

func TestSelectionAssistantServiceRejectsReservedWheelsetFitMutations(t *testing.T) {
	db, service := newSelectionAssistantTestService(t)
	flow := seedSelectionAssistantFlow(t, db, "wheelset-fit-helper", 100)
	version := selectionassistant.Version{
		FlowID:        flow.ID,
		VersionNumber: 1,
		Status:        selectionassistant.FlowVersionStatusDraft,
		Config:        datatypes.JSON([]byte(`{"kind":"product_selection_assistant","schema_version":1,"entry_node_key":"start","base_product_query":{"category_slug":"wheelset"},"nodes":[{"key":"start","type":"question"}]}`)),
	}
	require.NoError(t, db.Create(&version).Error)

	_, err := service.GetFlow(flow.ID)
	require.ErrorIs(t, err, ErrSelectionAssistantNotFound)

	_, err = service.SaveFlowConfiguration(flow.ID, SelectionAssistantFlowInput{
		Slug:                "wheelset-fit-helper",
		Name:                "Wheelset fit helper",
		ProductCategorySlug: selectionassistant.WheelsetProductCategorySlug,
		Version: SelectionAssistantVersionInput{
			Config: selectionassistant.Config{
				Kind:          selectionassistant.ConfigKind,
				SchemaVersion: 1,
				EntryNodeKey:  "start",
				BaseProductQuery: selectionassistant.BaseProductQuery{
					CategorySlug: selectionassistant.WheelsetProductCategorySlug,
				},
			},
		},
	})
	require.ErrorIs(t, err, ErrSelectionAssistantNotFound)

	loadedVersion, err := service.PublishVersion(version.ID, nil)
	require.ErrorIs(t, err, ErrSelectionAssistantNotFound)
	assert.Nil(t, loadedVersion)
}

func newSelectionAssistantTestService(t *testing.T) (*gorm.DB, *SelectionAssistantService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&selectionassistant.Flow{},
		&selectionassistant.Version{},
	))

	return db, NewSelectionAssistantService(repository.NewSelectionAssistantRepository(db))
}

func seedSelectionAssistantFlow(t *testing.T, db *gorm.DB, slug string, sortOrder int) selectionassistant.Flow {
	t.Helper()

	flow := selectionassistant.Flow{
		Slug:                slug,
		Name:                slug,
		Description:         slug,
		ProductCategorySlug: selectionassistant.WheelsetProductCategorySlug,
		IsEnabled:           true,
		SortOrder:           sortOrder,
	}
	require.NoError(t, db.Create(&flow).Error)
	return flow
}

func hasSelectionAssistantIssue(issues []SelectionAssistantValidationIssue, code, nodeKey string) bool {
	for _, issue := range issues {
		if issue.Code == code && issue.NodeKey == nodeKey {
			return true
		}
	}
	return false
}
