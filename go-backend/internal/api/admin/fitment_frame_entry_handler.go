package admin

import (
	"errors"
	"net/http"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FrameFitmentEntryHandler struct {
	service *service.FrameFitmentEntryService
}

type frameFitmentEntryRequest struct {
	BrandName           string                        `json:"brand_name"`
	ModelName           string                        `json:"model_name"`
	SeriesName          string                        `json:"series_name"`
	GenerationName      string                        `json:"generation_name"`
	YearMode            fitmentcatalogdomain.YearMode `json:"year_mode"`
	YearFrom            *int                          `json:"year_from"`
	YearTo              *int                          `json:"year_to"`
	MarketCode          string                        `json:"market_code"`
	Notes               string                        `json:"notes"`
	IsEnabled           bool                          `json:"is_enabled"`
	SortOrder           int                           `json:"sort_order"`
	HubSpecificationIDs []uint                        `json:"hub_specification_ids"`
}

func NewFrameFitmentEntryHandler(service *service.FrameFitmentEntryService) *FrameFitmentEntryHandler {
	return &FrameFitmentEntryHandler{service: service}
}

func (h *FrameFitmentEntryHandler) List(c *gin.Context) {
	page, pageSize := parseFitmentPagination(c)

	isEnabled, err := parseFitmentOptionalBool(c.Query("is_enabled"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_enabled must be true or false"})
		return
	}

	entries, total, err := h.service.List(service.FrameFitmentEntryListInput{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		IsEnabled: isEnabled,
	})
	if err != nil {
		respondFrameFitmentError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *FrameFitmentEntryHandler) Get(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "frame fitment entry ID")
	if !ok {
		return
	}
	entry, err := h.service.Get(id)
	if err != nil {
		respondFrameFitmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *FrameFitmentEntryHandler) Create(c *gin.Context) {
	var request frameFitmentEntryRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.Create(toFrameFitmentInput(request))
	if err != nil {
		respondFrameFitmentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entry": entry})
}

func (h *FrameFitmentEntryHandler) Update(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "frame fitment entry ID")
	if !ok {
		return
	}

	var request frameFitmentEntryRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.Update(id, toFrameFitmentInput(request))
	if err != nil {
		respondFrameFitmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *FrameFitmentEntryHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "frame fitment entry ID")
	if !ok {
		return
	}

	var request fitmentCatalogStatusRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.UpdateStatus(id, request.IsEnabled)
	if err != nil {
		respondFrameFitmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *FrameFitmentEntryHandler) Delete(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "frame fitment entry ID")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondFrameFitmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "frame fitment entry deleted"})
}

func toFrameFitmentInput(request frameFitmentEntryRequest) service.FrameFitmentEntryInput {
	return service.FrameFitmentEntryInput{
		BrandName:           request.BrandName,
		ModelName:           request.ModelName,
		SeriesName:          request.SeriesName,
		GenerationName:      request.GenerationName,
		YearMode:            request.YearMode,
		YearFrom:            request.YearFrom,
		YearTo:              request.YearTo,
		MarketCode:          request.MarketCode,
		Notes:               request.Notes,
		IsEnabled:           request.IsEnabled,
		SortOrder:           request.SortOrder,
		HubSpecificationIDs: request.HubSpecificationIDs,
	}
}

func respondFrameFitmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFrameFitmentEntryNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "frame fitment entry not found"})
	case errors.Is(err, service.ErrFrameFitmentEntryDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "frame fitment entry already exists"})
	case errors.Is(err, service.ErrFrameFitmentEntryInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage frame fitment entry"})
	}
}
