package product

import (
	"time"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

// publicSchwalbeTireCatalogItem is the storefront/admin selector projection.
// Source URLs stay in the database for provenance but are not part of this
// public response contract.
type publicSchwalbeTireCatalogItem struct {
	ArticleNo        string                                                 `json:"article_no"`
	EAN              *string                                                `json:"ean,omitempty"`
	ModelName        string                                                 `json:"model_name"`
	ETRTO            string                                                 `json:"etrto"`
	InchDesignation  *string                                                `json:"inch_designation,omitempty"`
	WeightG          *float64                                               `json:"weight_g,omitempty"`
	VersionLabel     *string                                                `json:"version_label,omitempty"`
	Compound         *string                                                `json:"compound,omitempty"`
	Color            *string                                                `json:"color,omitempty"`
	Bead             *string                                                `json:"bead,omitempty"`
	EBikeRating      *string                                                `json:"e_bike_rating,omitempty"`
	EPI              *int                                                   `json:"epi,omitempty"`
	LoadKG           *float64                                               `json:"load_kg,omitempty"`
	Seal             *string                                                `json:"seal,omitempty"`
	Tread            *string                                                `json:"tread,omitempty"`
	MinPressureBar   *float64                                               `json:"min_pressure_bar,omitempty"`
	MaxPressureBar   *float64                                               `json:"max_pressure_bar,omitempty"`
	MinPressurePSI   *float64                                               `json:"min_pressure_psi,omitempty"`
	MaxPressurePSI   *float64                                               `json:"max_pressure_psi,omitempty"`
	SourceCheckedAt  time.Time                                              `json:"source_checked_at"`
	ProductExists    bool                                                   `json:"product_exists"`
	WheelSize        *service.SchwalbeTireCatalogSelectorWheelSizeOption    `json:"wheel_size,omitempty"`
	RimCompatibility []service.SchwalbeTireCatalogSelectorRimCompatibility  `json:"rim_compatibility,omitempty"`
	RimWidthGuidance []service.SchwalbeTireCatalogSelectorRimWidthReference `json:"rim_width_guidance,omitempty"`
}

func publicSchwalbeTireCatalogItems(items []repository.SchwalbeTireCatalogItem) []publicSchwalbeTireCatalogItem {
	return publicSchwalbeTireCatalogItemsWithProjection(items, nil, nil, nil)
}

func publicSchwalbeTireCatalogItemsWithProjection(
	items []repository.SchwalbeTireCatalogItem,
	wheelSizeByArticle map[string]service.SchwalbeTireCatalogSelectorWheelSizeOption,
	rimCompatibilityByArticle map[string][]service.SchwalbeTireCatalogSelectorRimCompatibility,
	rimWidthReferenceByArticle map[string][]service.SchwalbeTireCatalogSelectorRimWidthReference,
) []publicSchwalbeTireCatalogItem {
	publicItems := make([]publicSchwalbeTireCatalogItem, len(items))
	for index, item := range items {
		var wheelSize *service.SchwalbeTireCatalogSelectorWheelSizeOption
		if value, ok := wheelSizeByArticle[item.ArticleNo]; ok {
			wheelSize = &value
		}
		publicItems[index] = publicSchwalbeTireCatalogItem{
			ArticleNo:        item.ArticleNo,
			EAN:              item.EAN,
			ModelName:        item.ModelName,
			ETRTO:            item.ETRTO,
			InchDesignation:  item.InchDesignation,
			WeightG:          item.WeightG,
			VersionLabel:     item.VersionLabel,
			Compound:         item.Compound,
			Color:            item.Color,
			Bead:             item.Bead,
			EBikeRating:      item.EBikeRating,
			EPI:              item.EPI,
			LoadKG:           item.LoadKG,
			Seal:             item.Seal,
			Tread:            item.Tread,
			MinPressureBar:   item.MinPressureBar,
			MaxPressureBar:   item.MaxPressureBar,
			MinPressurePSI:   item.MinPressurePSI,
			MaxPressurePSI:   item.MaxPressurePSI,
			SourceCheckedAt:  item.SourceCheckedAt,
			ProductExists:    item.ProductExists,
			WheelSize:        wheelSize,
			RimCompatibility: rimCompatibilityByArticle[item.ArticleNo],
			RimWidthGuidance: rimWidthReferenceByArticle[item.ArticleNo],
		}
	}
	return publicItems
}

type publicSchwalbeTireCatalogSelectorResponse struct {
	Items         []publicSchwalbeTireCatalogItem                  `json:"items"`
	Page          int                                              `json:"page"`
	PageSize      int                                              `json:"page_size"`
	Total         int                                              `json:"total"`
	TotalPages    int                                              `json:"total_pages"`
	FilterOptions service.SchwalbeTireCatalogSelectorFilterOptions `json:"filter_options"`
}

func publicSchwalbeTireCatalogSelectorResponseFromPage(page *service.SchwalbeTireCatalogSelectorPage) publicSchwalbeTireCatalogSelectorResponse {
	return publicSchwalbeTireCatalogSelectorResponse{
		Items:         publicSchwalbeTireCatalogItemsWithProjection(page.Items, page.WheelSizeByArticle, page.RimCompatibilityByArticle, page.RimWidthReferenceByArticle),
		Page:          page.Page,
		PageSize:      page.PageSize,
		Total:         page.Total,
		TotalPages:    page.TotalPages,
		FilterOptions: page.FilterOptions,
	}
}
