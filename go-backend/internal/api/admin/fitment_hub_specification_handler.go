package admin

import (
	"errors"
	"net/http"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FitmentHubSpecificationHandler struct {
	service *service.FitmentHubSpecificationService
}

type hubSpecificationRequest struct {
	SpecCode      string                           `json:"spec_code"`
	DisplayName   string                           `json:"display_name"`
	Position      fitmentcatalogdomain.HubPosition `json:"position"`
	AxleType      fitmentcatalogdomain.HubAxleType `json:"axle_type"`
	AxleSpacingMM int                              `json:"axle_spacing_mm"`
	WRMM          *float64                         `json:"wr_mm"`
	WLMM          *float64                         `json:"wl_mm"`
	PCDRMM        *float64                         `json:"pcdr_mm"`
	PCDLMM        *float64                         `json:"pcdl_mm"`
	Notes         string                           `json:"notes"`
	IsEnabled     bool                             `json:"is_enabled"`
	SortOrder     int                              `json:"sort_order"`
}

func NewFitmentHubSpecificationHandler(
	service *service.FitmentHubSpecificationService,
) *FitmentHubSpecificationHandler {
	return &FitmentHubSpecificationHandler{service: service}
}

func (h *FitmentHubSpecificationHandler) List(c *gin.Context) {
	page, pageSize := parseFitmentPagination(c)

	isEnabled, err := parseFitmentOptionalBool(c.Query("is_enabled"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_enabled must be true or false"})
		return
	}

	specifications, total, err := h.service.List(service.FitmentHubSpecificationListInput{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		Position:  fitmentcatalogdomain.HubPosition(c.Query("position")),
		AxleType:  fitmentcatalogdomain.HubAxleType(c.Query("axle_type")),
		IsEnabled: isEnabled,
	})
	if err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"specifications": specifications,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *FitmentHubSpecificationHandler) Get(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "fitment hub specification ID")
	if !ok {
		return
	}
	specification, err := h.service.Get(id)
	if err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"specification": specification})
}

func (h *FitmentHubSpecificationHandler) Create(c *gin.Context) {
	var request hubSpecificationRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specification, err := h.service.Create(toFitmentHubSpecificationInput(request))
	if err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"specification": specification})
}

func (h *FitmentHubSpecificationHandler) Update(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "fitment hub specification ID")
	if !ok {
		return
	}

	var request hubSpecificationRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specification, err := h.service.Update(id, toFitmentHubSpecificationInput(request))
	if err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"specification": specification})
}

func (h *FitmentHubSpecificationHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "fitment hub specification ID")
	if !ok {
		return
	}

	var request fitmentCatalogStatusRequest
	if err := bindFitmentJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specification, err := h.service.UpdateStatus(id, request.IsEnabled)
	if err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"specification": specification})
}

func (h *FitmentHubSpecificationHandler) Delete(c *gin.Context) {
	id, ok := parseFitmentCatalogID(c, "fitment hub specification ID")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondFitmentHubSpecificationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "fitment hub specification deleted"})
}

func toFitmentHubSpecificationInput(request hubSpecificationRequest) service.FitmentHubSpecificationInput {
	return service.FitmentHubSpecificationInput{
		SpecCode:      request.SpecCode,
		DisplayName:   request.DisplayName,
		Position:      request.Position,
		AxleType:      request.AxleType,
		AxleSpacingMM: request.AxleSpacingMM,
		WRMM:          request.WRMM,
		WLMM:          request.WLMM,
		PCDRMM:        request.PCDRMM,
		PCDLMM:        request.PCDLMM,
		Notes:         request.Notes,
		IsEnabled:     request.IsEnabled,
		SortOrder:     request.SortOrder,
	}
}

func respondFitmentHubSpecificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFitmentHubSpecificationNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "fitment hub specification not found"})
	case errors.Is(err, service.ErrFitmentHubSpecificationDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "fitment hub specification already exists"})
	case errors.Is(err, service.ErrFitmentHubSpecificationInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "fitment hub specification is in use; remove references or disable it instead"})
	case errors.Is(err, service.ErrFitmentHubSpecificationInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage fitment hub specification"})
	}
}
