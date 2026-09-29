package repository

import "time"

// SchwalbeTireRimWidthCombinationRule is an official possible-combination
// range from migration 362. It is guidance for selector filtering, not a
// model-specific compatibility certification.
type SchwalbeTireRimWidthCombinationRule struct {
	ID                 uint      `json:"id"`
	TireWidthMinMM     int       `json:"tire_width_min_mm"`
	TireWidthMaxMM     int       `json:"tire_width_max_mm"`
	InnerRimWidthMinMM int       `json:"inner_rim_width_min_mm"`
	InnerRimWidthMaxMM int       `json:"inner_rim_width_max_mm"`
	SourceBasis        string    `json:"source_basis"`
	SourceVersion      string    `json:"source_version"`
	SourceURL          string    `json:"source_url"`
	SourceCheckedAt    time.Time `json:"source_checked_at"`
}

// ListSchwalbeTireRimWidthCombinationRules returns the source-backed width
// ranges in deterministic tire-width order for selector and modal consumers.
func (r *ProductRepository) ListSchwalbeTireRimWidthCombinationRules() ([]SchwalbeTireRimWidthCombinationRule, error) {
	rules := make([]SchwalbeTireRimWidthCombinationRule, 0)
	err := r.db.Table("schwalbe_tire_rim_width_combination_rules").
		Select(`
			id,
			tire_width_min_mm,
			tire_width_max_mm,
			inner_rim_width_min_mm,
			inner_rim_width_max_mm,
			source_basis,
			source_version,
			source_url,
			source_checked_at`).
		Order("tire_width_min_mm ASC").
		Order("inner_rim_width_min_mm ASC").
		Scan(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}
