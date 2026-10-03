package service

import (
	"strings"

	"commerce-platform/internal/repository"
)

// SchwalbeTireCatalogSelectorRimCompatibility is the small public verdict
// needed by a catalog card. Source provenance remains in the maintenance table
// and is deliberately not exposed in the selector item response.
type SchwalbeTireCatalogSelectorRimCompatibility struct {
	RimSystem string `json:"rim_system"` // hooked or hookless
	Status    string `json:"status"`     // supported
}

const (
	schwalbeRimSystemHooked   = "hooked"
	schwalbeRimSystemHookless = "hookless"
	schwalbeHooklessSupported = "supported"
)

// schwalbeTireCatalogRimCompatibilityForRows resolves article-level facts
// first, then model-level fallbacks. An explicit article-level not_supported
// or unknown record therefore blocks a broader model fallback. Missing and
// invalid records stay unknown and never create a Hookless layer.
func schwalbeTireCatalogRimCompatibilityForRows(
	rows []schwalbeTireCatalogSelectorRow,
	records []repository.SchwalbeTireHooklessCompatibility,
) map[string][]SchwalbeTireCatalogSelectorRimCompatibility {
	articleStatuses := make(map[string]string)
	modelStatuses := make(map[string]string)
	for _, record := range records {
		status := strings.ToLower(strings.TrimSpace(record.Status))
		if status != "supported" && status != "not_supported" && status != "unknown" {
			continue
		}
		articleNo := normalizeSchwalbeHooklessLookupValue(record.ArticleNo)
		modelName := normalizeSchwalbeHooklessLookupValue(record.ModelName)
		if articleNo != "" {
			articleStatuses[articleNo] = status
			continue
		}
		if modelName != "" {
			modelStatuses[modelName] = status
		}
	}

	compatibilityByArticle := make(map[string][]SchwalbeTireCatalogSelectorRimCompatibility, len(rows))
	for _, row := range rows {
		compatibility := []SchwalbeTireCatalogSelectorRimCompatibility{
			{RimSystem: schwalbeRimSystemHooked, Status: schwalbeHooklessSupported},
		}
		status, ok := articleStatuses[strings.TrimSpace(row.item.ArticleNo)]
		if !ok {
			status, ok = modelStatuses[strings.TrimSpace(row.item.ModelName)]
		}
		if ok && status == schwalbeHooklessSupported {
			compatibility = append(compatibility, SchwalbeTireCatalogSelectorRimCompatibility{
				RimSystem: schwalbeRimSystemHookless,
				Status:    schwalbeHooklessSupported,
			})
		}
		compatibilityByArticle[row.item.ArticleNo] = compatibility
	}
	return compatibilityByArticle
}

func normalizeSchwalbeHooklessLookupValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
