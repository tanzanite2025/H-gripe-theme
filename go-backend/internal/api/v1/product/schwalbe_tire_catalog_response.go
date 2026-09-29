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
	ArticleNo        string                                                `json:"article_no"`
	EAN              *string                                               `json:"ean,omitempty"`
	ModelName        string                                                `json:"model_name"`
	ETRTO            string                                                `json:"etrto"`
	InchDesignation  *string                                               `json:"inch_designation,omitempty"`
	WeightG          *float64                                              `json:"weight_g,omitempty"`
	VersionLabel     *string                                               `json:"version_label,omitempty"`
	Compound         *string                                               `json:"compound,omitempty"`
	Color            *string                                               `json:"color,omitempty"`
	Bead             *string                                               `json:"bead,omitempty"`
	EBikeRating      *string                                               `json:"e_bike_rating,omitempty"`
	EPI              *int                                                  `json:"epi,omitempty"`
	LoadKG           *float64                                              `json:"load_kg,omitempty"`
	Seal             *string                                               `json:"seal,omitempty"`
	Tread            *string                                               `json:"tread,omitempty"`
	MinPressureBar   *float64                                              `json:"min_pressure_bar,omitempty"`
	MaxPressureBar   *float64                                              `json:"max_pressure_bar,omitempty"`
	MinPressurePSI   *float64                                              `json:"min_pressure_psi,omitempty"`
	MaxPressurePSI   *float64                                              `json:"max_pressure_psi,omitempty"`
	SourceCheckedAt  time.Time                                             `json:"source_checked_at"`
	ProductExists    bool                                                  `json:"product_exists"`
	RimWidthGuidance []service.SchwalbeTireCatalogSelectorRimWidthGuidance `json:"rim_width_guidance,omitempty"`
}

func publicSchwalbeTireCatalogItems(items []repository.SchwalbeTireCatalogItem) []publicSchwalbeTireCatalogItem {
	return publicSchwalbeTireCatalogItemsWithRimWidthGuidance(items, nil)
}

func publicSchwalbeTireCatalogItemsWithRimWidthGuidance(
	items []repository.SchwalbeTireCatalogItem,
	guidanceByArticle map[string][]service.SchwalbeTireCatalogSelectorRimWidthGuidance,
) []publicSchwalbeTireCatalogItem {
	publicItems := make([]publicSchwalbeTireCatalogItem, len(items))
	for index, item := range items {
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
			RimWidthGuidance: guidanceByArticle[item.ArticleNo],
		}
	}
	return publicItems
}

type publicSchwalbeTireCatalogSelectorRimWidthContext struct {
	InnerRimWidthMM float64    `json:"inner_rim_width_mm"`
	GuidanceStatus  string     `json:"guidance_status"`
	SourceBasis     string     `json:"source_basis,omitempty"`
	SourceVersion   string     `json:"source_version,omitempty"`
	SourceCheckedAt *time.Time `json:"source_checked_at,omitempty"`
}

type publicSchwalbeTireCatalogSelectorResponse struct {
	Items           []publicSchwalbeTireCatalogItem                   `json:"items"`
	Page            int                                               `json:"page"`
	PageSize        int                                               `json:"page_size"`
	Total           int                                               `json:"total"`
	TotalPages      int                                               `json:"total_pages"`
	FilterOptions   service.SchwalbeTireCatalogSelectorFilterOptions  `json:"filter_options"`
	RimWidthContext *publicSchwalbeTireCatalogSelectorRimWidthContext `json:"rim_width_context,omitempty"`
}

func publicSchwalbeTireCatalogSelectorResponseFromPage(page *service.SchwalbeTireCatalogSelectorPage) publicSchwalbeTireCatalogSelectorResponse {
	var rimWidthContext *publicSchwalbeTireCatalogSelectorRimWidthContext
	if page.RimWidthContext != nil {
		rimWidthContext = &publicSchwalbeTireCatalogSelectorRimWidthContext{
			InnerRimWidthMM: page.RimWidthContext.InnerRimWidthMM,
			GuidanceStatus:  page.RimWidthContext.GuidanceStatus,
			SourceBasis:     page.RimWidthContext.SourceBasis,
			SourceVersion:   page.RimWidthContext.SourceVersion,
			SourceCheckedAt: page.RimWidthContext.SourceCheckedAt,
		}
	}
	return publicSchwalbeTireCatalogSelectorResponse{
		Items:           publicSchwalbeTireCatalogItemsWithRimWidthGuidance(page.Items, page.RimWidthGuidanceByArticle),
		Page:            page.Page,
		PageSize:        page.PageSize,
		Total:           page.Total,
		TotalPages:      page.TotalPages,
		FilterOptions:   page.FilterOptions,
		RimWidthContext: rimWidthContext,
	}
}
