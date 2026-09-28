package repository

import (
	"strings"
	"time"
)

// SchwalbeTireCatalogItem is a sourced candidate record. ProductExists is
// derived from actual products using the schwalbe_tire template and is not a
// property of the catalog row itself.
type SchwalbeTireCatalogItem struct {
	ArticleNo       string    `json:"article_no"`
	EAN             *string   `json:"ean,omitempty"`
	ModelName       string    `json:"model_name"`
	ETRTO           string    `json:"etrto"`
	InchDesignation *string   `json:"inch_designation,omitempty"`
	WeightG         *float64  `json:"weight_g,omitempty"`
	VersionLabel    *string   `json:"version_label,omitempty"`
	Compound        *string   `json:"compound,omitempty"`
	Color           *string   `json:"color,omitempty"`
	Bead            *string   `json:"bead,omitempty"`
	EBikeRating     *string   `json:"e_bike_rating,omitempty"`
	EPI             *int      `json:"epi,omitempty"`
	LoadKG          *float64  `json:"load_kg,omitempty"`
	Seal            *string   `json:"seal,omitempty"`
	Tread           *string   `json:"tread,omitempty"`
	MinPressureBar  *float64  `json:"min_pressure_bar,omitempty"`
	MaxPressureBar  *float64  `json:"max_pressure_bar,omitempty"`
	MinPressurePSI  *float64  `json:"min_pressure_psi,omitempty"`
	MaxPressurePSI  *float64  `json:"max_pressure_psi,omitempty"`
	SourceURL       string    `json:"source_url"`
	SourceCheckedAt time.Time `json:"source_checked_at"`
	ProductExists   bool      `json:"product_exists"`
}

func (r *ProductRepository) ListSchwalbeTireCatalog(search string) ([]SchwalbeTireCatalogItem, error) {
	items := make([]SchwalbeTireCatalogItem, 0)
	query := r.db.Table("schwalbe_tire_specifications AS catalog").
		Select(`
			catalog.article_no,
			catalog.ean,
			catalog.model_name,
			catalog.etrto,
			catalog.inch_designation,
			catalog.weight_g,
			catalog.version_label,
			catalog.compound,
			catalog.color,
			catalog.bead,
			catalog.e_bike_rating,
			catalog.epi,
			catalog.load_kg,
			catalog.seal,
			catalog.tread,
			catalog.min_pressure_bar,
			catalog.max_pressure_bar,
			catalog.min_pressure_psi,
			catalog.max_pressure_psi,
			catalog.source_url,
			catalog.source_checked_at,
			EXISTS (
				SELECT 1
				FROM products p
				JOIN product_specification_templates t
				  ON t.id = p.product_specification_template_id
				 AND t.slug = 'schwalbe_tire'
				JOIN product_spec_definitions d
				  ON d.product_specification_template_id = t.id
				 AND d.slug = 'article_no'
				JOIN product_spec_values v
				  ON v.product_id = p.id
				 AND v.spec_definition_id = d.id
				WHERE LOWER(TRIM(v.value)) = LOWER(TRIM(catalog.article_no))
			) AS product_exists`)

	if normalizedSearch := strings.TrimSpace(search); normalizedSearch != "" {
		pattern := "%" + strings.ToLower(normalizedSearch) + "%"
		query = query.Where(`
			LOWER(catalog.article_no) LIKE ?
			OR LOWER(catalog.model_name) LIKE ?
			OR LOWER(catalog.etrto) LIKE ?
			OR LOWER(COALESCE(catalog.inch_designation, '')) LIKE ?`,
			pattern, pattern, pattern, pattern,
		)
	}

	err := query.Order("catalog.model_name ASC").Order("catalog.etrto ASC").Order("catalog.article_no ASC").Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
