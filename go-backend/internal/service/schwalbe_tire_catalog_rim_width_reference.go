package service

import (
	"commerce-platform/internal/domain/tirerim"
)

// schwalbeTireCatalogRimWidthReferenceForRows derives display guidance only
// from the selected item's nominal ETRTO width and explicit rim-system facts.
// It never reads the retired Schwalbe broad combination matrix.
func schwalbeTireCatalogRimWidthReferenceForRows(
	rows []schwalbeTireCatalogSelectorRow,
	rimCompatibilityByArticle map[string][]SchwalbeTireCatalogSelectorRimCompatibility,
) map[string][]SchwalbeTireCatalogSelectorRimWidthReference {
	guidanceByArticle := make(map[string][]SchwalbeTireCatalogSelectorRimWidthReference, len(rows))
	for _, row := range rows {
		if row.dimensions.NominalTireWidthMM == nil {
			continue
		}

		systems := make([]tirerim.RimSystem, 0, 2)
		for _, compatibility := range rimCompatibilityByArticle[row.item.ArticleNo] {
			if compatibility.Status != "supported" {
				continue
			}
			system := tirerim.NormalizeRimSystem(compatibility.RimSystem)
			if system != tirerim.RimSystemHooked && system != tirerim.RimSystemHookless {
				continue
			}
			systems = append(systems, system)
		}
		guidance := tirerim.BuildTireRimWidthGuidanceForNominalWidth(*row.dimensions.NominalTireWidthMM, systems)
		if len(guidance) > 0 {
			guidanceByArticle[row.item.ArticleNo] = guidance
		}
	}
	return guidanceByArticle
}
