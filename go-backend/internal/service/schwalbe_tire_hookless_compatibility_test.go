package service

import (
	"testing"

	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestSchwalbeTireCatalogRimCompatibilityAlwaysIncludesHooked(t *testing.T) {
	articleNo := "11600385.03"
	modelName := "Rocket Ron"
	rows := []schwalbeTireCatalogSelectorRow{
		{item: repository.SchwalbeTireCatalogItem{ArticleNo: articleNo, ModelName: modelName}},
	}

	compatibility := schwalbeTireCatalogRimCompatibilityForRows(rows, nil)
	require.Equal(t, []SchwalbeTireCatalogSelectorRimCompatibility{
		{RimSystem: "hooked", Status: "supported"},
	}, compatibility[articleNo])
}

func TestSchwalbeTireCatalogRimCompatibilityUsesArticleOverrideBeforeModelFallback(t *testing.T) {
	articleNo := "11600385.03"
	modelName := "Rocket Ron"
	rows := []schwalbeTireCatalogSelectorRow{
		{item: repository.SchwalbeTireCatalogItem{ArticleNo: articleNo, ModelName: modelName}},
		{item: repository.SchwalbeTireCatalogItem{ArticleNo: "other", ModelName: modelName}},
	}
	articleStatus := "not_supported"
	modelStatus := "supported"

	compatibility := schwalbeTireCatalogRimCompatibilityForRows(rows, []repository.SchwalbeTireHooklessCompatibility{
		{ArticleNo: &articleNo, Status: articleStatus},
		{ModelName: &modelName, Status: modelStatus},
	})
	require.Len(t, compatibility[articleNo], 1)
	require.Len(t, compatibility["other"], 2)
	require.Equal(t, "hookless", compatibility["other"][1].RimSystem)
}

func TestSchwalbeTireCatalogRimCompatibilityDoesNotInferUnknownModels(t *testing.T) {
	modelName := "Wicked Will"
	rows := []schwalbeTireCatalogSelectorRow{
		{item: repository.SchwalbeTireCatalogItem{ArticleNo: "11654268", ModelName: modelName}},
	}
	compatibility := schwalbeTireCatalogRimCompatibilityForRows(rows, []repository.SchwalbeTireHooklessCompatibility{
		{ModelName: &modelName, Status: "unknown"},
	})
	require.Len(t, compatibility["11654268"], 1)
}
