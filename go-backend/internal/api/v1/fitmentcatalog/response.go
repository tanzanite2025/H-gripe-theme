package fitmentcatalog

import fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type hubSpecificationResponse struct {
	ID            uint                             `json:"id"`
	SpecCode      string                           `json:"spec_code"`
	DisplayName   string                           `json:"display_name"`
	Position      fitmentcatalogdomain.HubPosition `json:"position"`
	AxleType      fitmentcatalogdomain.HubAxleType `json:"axle_type"`
	AxleSpacingMM int                              `json:"axle_spacing_mm"`
	Notes         string                           `json:"notes"`
}

type frameEntryResponse struct {
	ID                    uint                          `json:"id"`
	BrandName             string                        `json:"brand_name"`
	ModelName             string                        `json:"model_name"`
	SeriesName            string                        `json:"series_name"`
	GenerationName        string                        `json:"generation_name"`
	YearMode              fitmentcatalogdomain.YearMode `json:"year_mode"`
	YearFrom              *int                          `json:"year_from"`
	YearTo                *int                          `json:"year_to"`
	MarketCode            string                        `json:"market_code"`
	Notes                 string                        `json:"notes"`
	HubSpecifications     []hubSpecificationResponse    `json:"hub_specifications"`
	HubSpecificationCount int                           `json:"hub_specification_count"`
}

type forkEntryResponse struct {
	ID                    uint                          `json:"id"`
	BrandName             string                        `json:"brand_name"`
	ModelName             string                        `json:"model_name"`
	SeriesName            string                        `json:"series_name"`
	GenerationName        string                        `json:"generation_name"`
	YearMode              fitmentcatalogdomain.YearMode `json:"year_mode"`
	YearFrom              *int                          `json:"year_from"`
	YearTo                *int                          `json:"year_to"`
	MarketCode            string                        `json:"market_code"`
	Notes                 string                        `json:"notes"`
	HubSpecifications     []hubSpecificationResponse    `json:"hub_specifications"`
	HubSpecificationCount int                           `json:"hub_specification_count"`
}

func newPaginationResponse(page, pageSize int, total int64) paginationResponse {
	return paginationResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}
}

func newFrameEntryResponse(entry *fitmentcatalogdomain.FrameFitmentEntry) frameEntryResponse {
	if entry == nil {
		return frameEntryResponse{HubSpecifications: []hubSpecificationResponse{}}
	}
	hubs := newEnabledHubSpecificationResponses(entry.HubSpecifications)
	return frameEntryResponse{
		ID:                    entry.ID,
		BrandName:             entry.BrandName,
		ModelName:             entry.ModelName,
		SeriesName:            entry.SeriesName,
		GenerationName:        entry.GenerationName,
		YearMode:              entry.YearMode,
		YearFrom:              entry.YearFrom,
		YearTo:                entry.YearTo,
		MarketCode:            entry.MarketCode,
		Notes:                 entry.Notes,
		HubSpecifications:     hubs,
		HubSpecificationCount: len(hubs),
	}
}

func newForkEntryResponse(entry *fitmentcatalogdomain.ForkFitmentEntry) forkEntryResponse {
	if entry == nil {
		return forkEntryResponse{HubSpecifications: []hubSpecificationResponse{}}
	}
	hubs := newEnabledHubSpecificationResponses(entry.HubSpecifications)
	return forkEntryResponse{
		ID:                    entry.ID,
		BrandName:             entry.BrandName,
		ModelName:             entry.ModelName,
		SeriesName:            entry.SeriesName,
		GenerationName:        entry.GenerationName,
		YearMode:              entry.YearMode,
		YearFrom:              entry.YearFrom,
		YearTo:                entry.YearTo,
		MarketCode:            entry.MarketCode,
		Notes:                 entry.Notes,
		HubSpecifications:     hubs,
		HubSpecificationCount: len(hubs),
	}
}

func newHubSpecificationResponse(specification fitmentcatalogdomain.HubSpecification) hubSpecificationResponse {
	return hubSpecificationResponse{
		ID:            specification.ID,
		SpecCode:      specification.SpecCode,
		DisplayName:   specification.DisplayName,
		Position:      specification.Position,
		AxleType:      specification.AxleType,
		AxleSpacingMM: specification.AxleSpacingMM,
		Notes:         specification.Notes,
	}
}

func newEnabledHubSpecificationResponses(
	specifications []fitmentcatalogdomain.HubSpecification,
) []hubSpecificationResponse {
	responses := make([]hubSpecificationResponse, 0, len(specifications))
	for _, specification := range specifications {
		if !specification.IsEnabled {
			continue
		}
		responses = append(responses, newHubSpecificationResponse(specification))
	}
	return responses
}
