package service

import (
	"sort"

	"commerce-platform/internal/repository"
)

// SchwalbeTireCatalogSelectorRimWidthGuidance describes the source-backed
// possible-combination range that matched one catalog row. It is guidance for
// narrowing a catalog search, not a model-specific compatibility certificate.
type SchwalbeTireCatalogSelectorRimWidthGuidance struct {
	TireWidthMinMM     int `json:"tire_width_min_mm"`
	TireWidthMaxMM     int `json:"tire_width_max_mm"`
	InnerRimWidthMinMM int `json:"inner_rim_width_min_mm"`
	InnerRimWidthMaxMM int `json:"inner_rim_width_max_mm"`
}

// schwalbeTireRimWidthGuidanceForTireWidth returns every official
// possible-combination range that contains the catalog row's ETRTO width.
// Boundary values are intentionally inclusive. The result is reference data
// for the card; it is not a model-specific compatibility certificate.
func schwalbeTireRimWidthGuidanceForTireWidth(
	nominalTireWidthMM *int,
	rules []repository.SchwalbeTireRimWidthCombinationRule,
) []SchwalbeTireCatalogSelectorRimWidthGuidance {
	if nominalTireWidthMM == nil || *nominalTireWidthMM <= 0 {
		return nil
	}

	matches := make([]SchwalbeTireCatalogSelectorRimWidthGuidance, 0)
	seen := make(map[SchwalbeTireCatalogSelectorRimWidthGuidance]struct{})
	for _, rule := range rules {
		if !validSchwalbeTireRimWidthRule(rule) ||
			*nominalTireWidthMM < rule.TireWidthMinMM || *nominalTireWidthMM > rule.TireWidthMaxMM {
			continue
		}

		match := schwalbeTireCatalogSelectorRimWidthGuidanceFromRule(rule)
		if _, exists := seen[match]; exists {
			continue
		}
		seen[match] = struct{}{}
		matches = append(matches, match)
	}

	sortSchwalbeTireRimWidthGuidance(matches)
	return matches
}

// schwalbeTireRimWidthGuidanceFor returns every official possible-combination
// range that contains both the catalog row's ETRTO width and the user's rim
// inner width. Boundary values are intentionally inclusive.
func schwalbeTireRimWidthGuidanceFor(
	nominalTireWidthMM *int,
	innerRimWidthMM *float64,
	rules []repository.SchwalbeTireRimWidthCombinationRule,
) []SchwalbeTireCatalogSelectorRimWidthGuidance {
	if nominalTireWidthMM == nil || innerRimWidthMM == nil {
		return nil
	}
	if *nominalTireWidthMM <= 0 || *innerRimWidthMM <= 0 {
		return nil
	}

	matches := make([]SchwalbeTireCatalogSelectorRimWidthGuidance, 0)
	seen := make(map[SchwalbeTireCatalogSelectorRimWidthGuidance]struct{})
	for _, rule := range rules {
		if !validSchwalbeTireRimWidthRule(rule) {
			continue
		}
		if *nominalTireWidthMM < rule.TireWidthMinMM || *nominalTireWidthMM > rule.TireWidthMaxMM ||
			*innerRimWidthMM < float64(rule.InnerRimWidthMinMM) ||
			*innerRimWidthMM > float64(rule.InnerRimWidthMaxMM) {
			continue
		}

		match := schwalbeTireCatalogSelectorRimWidthGuidanceFromRule(rule)
		if _, exists := seen[match]; exists {
			continue
		}
		seen[match] = struct{}{}
		matches = append(matches, match)
	}

	sortSchwalbeTireRimWidthGuidance(matches)
	return matches
}

func validSchwalbeTireRimWidthRule(rule repository.SchwalbeTireRimWidthCombinationRule) bool {
	return rule.TireWidthMinMM > 0 && rule.TireWidthMaxMM > 0 &&
		rule.InnerRimWidthMinMM > 0 && rule.InnerRimWidthMaxMM > 0 &&
		rule.TireWidthMinMM <= rule.TireWidthMaxMM &&
		rule.InnerRimWidthMinMM <= rule.InnerRimWidthMaxMM
}

func schwalbeTireCatalogSelectorRimWidthGuidanceFromRule(
	rule repository.SchwalbeTireRimWidthCombinationRule,
) SchwalbeTireCatalogSelectorRimWidthGuidance {
	return SchwalbeTireCatalogSelectorRimWidthGuidance{
		TireWidthMinMM:     rule.TireWidthMinMM,
		TireWidthMaxMM:     rule.TireWidthMaxMM,
		InnerRimWidthMinMM: rule.InnerRimWidthMinMM,
		InnerRimWidthMaxMM: rule.InnerRimWidthMaxMM,
	}
}

func sortSchwalbeTireRimWidthGuidance(matches []SchwalbeTireCatalogSelectorRimWidthGuidance) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].TireWidthMinMM != matches[j].TireWidthMinMM {
			return matches[i].TireWidthMinMM < matches[j].TireWidthMinMM
		}
		if matches[i].TireWidthMaxMM != matches[j].TireWidthMaxMM {
			return matches[i].TireWidthMaxMM < matches[j].TireWidthMaxMM
		}
		if matches[i].InnerRimWidthMinMM != matches[j].InnerRimWidthMinMM {
			return matches[i].InnerRimWidthMinMM < matches[j].InnerRimWidthMinMM
		}
		return matches[i].InnerRimWidthMaxMM < matches[j].InnerRimWidthMaxMM
	})
}

func schwalbeTireRimWidthInputCovered(
	innerRimWidthMM *float64,
	rules []repository.SchwalbeTireRimWidthCombinationRule,
) bool {
	if innerRimWidthMM == nil || *innerRimWidthMM <= 0 {
		return false
	}
	for _, rule := range rules {
		if rule.InnerRimWidthMinMM > 0 && rule.InnerRimWidthMaxMM >= rule.InnerRimWidthMinMM &&
			*innerRimWidthMM >= float64(rule.InnerRimWidthMinMM) &&
			*innerRimWidthMM <= float64(rule.InnerRimWidthMaxMM) {
			return true
		}
	}
	return false
}
