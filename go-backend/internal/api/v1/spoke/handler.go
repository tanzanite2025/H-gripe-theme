package spoke

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	spokeService *service.SpokeService
}

func NewHandler(spokeService *service.SpokeService) *Handler {
	return &Handler{spokeService: spokeService}
}

func (h *Handler) GetExport(c *gin.Context) {
	c.JSON(http.StatusOK, h.spokeService.GetExport())
}

func (h *Handler) ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 5
	}

	items, total, err := h.spokeService.ListHistory(c.Query("search"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_history_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"meta": gin.H{
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"page":        page,
			"per_page":    pageSize,
		},
	})
}

type CalcRequest struct {
	RimID         string `json:"rimId" binding:"required"`
	HubID         string `json:"hubId" binding:"required"`
	WheelPosition string `json:"wheelPosition" binding:"required"`
	SpokeCount    int    `json:"spokeCount" binding:"required"`
	Crossing      int    `json:"crossing"`
}

func (h *Handler) Calculate(c *gin.Context) {
	var req CalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	result, err := h.spokeService.Calculate(service.SpokeCalculationInput{
		RimID:         req.RimID,
		HubID:         req.HubID,
		WheelPosition: req.WheelPosition,
		SpokeCount:    req.SpokeCount,
		Crossing:      req.Crossing,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSpokeGeometryNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Unknown rim or hub geometry"})
		case errors.Is(err, service.ErrSpokeHubGeometryMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Hub geometry not available for requested position"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
