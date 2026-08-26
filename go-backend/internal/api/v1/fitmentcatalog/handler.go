package fitmentcatalog

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	frameEntries      *service.FrameFitmentEntryService
	forkEntries       *service.ForkFitmentEntryService
	hubSpecifications *service.FitmentHubSpecificationService
}

func NewHandler(
	frameEntries *service.FrameFitmentEntryService,
	forkEntries *service.ForkFitmentEntryService,
	hubSpecifications *service.FitmentHubSpecificationService,
) *Handler {
	return &Handler{
		frameEntries:      frameEntries,
		forkEntries:       forkEntries,
		hubSpecifications: hubSpecifications,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/frame-entries", h.ListFrameEntries)
	group.GET("/frame-entries/:id", h.GetFrameEntry)
	group.GET("/fork-entries", h.ListForkEntries)
	group.GET("/fork-entries/:id", h.GetForkEntry)
	group.GET("/hub-specifications", h.ListHubSpecifications)
	group.GET("/hub-specifications/:id", h.GetHubSpecification)
}

func (h *Handler) ListFrameEntries(c *gin.Context) {
	page, pageSize := parsePagination(c)
	year, ok := parseYear(c)
	if !ok {
		return
	}
	enabled := true
	entries, total, err := h.frameEntries.List(service.FrameFitmentEntryListInput{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		IsEnabled: &enabled,
		Year:      year,
	})
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}

	responses := make([]frameEntryResponse, 0, len(entries))
	for _, entry := range entries {
		detail, err := h.frameEntries.Get(entry.ID)
		if err != nil {
			respondFitmentCatalogError(c, err)
			return
		}
		if detail == nil || !detail.IsEnabled {
			continue
		}
		responses = append(responses, newFrameEntryResponse(detail))
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"frame_entries": responses,
		"pagination":    newPaginationResponse(page, pageSize, total),
	}})
}

func (h *Handler) GetFrameEntry(c *gin.Context) {
	id, ok := parseID(c, "frame fitment entry ID")
	if !ok {
		return
	}

	entry, err := h.frameEntries.Get(id)
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}
	if entry == nil || !entry.IsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "frame fitment entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entry": newFrameEntryResponse(entry)}})
}

func (h *Handler) ListForkEntries(c *gin.Context) {
	page, pageSize := parsePagination(c)
	year, ok := parseYear(c)
	if !ok {
		return
	}
	enabled := true
	entries, total, err := h.forkEntries.List(service.ForkFitmentEntryListInput{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		IsEnabled: &enabled,
		Year:      year,
	})
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}

	responses := make([]forkEntryResponse, 0, len(entries))
	for _, entry := range entries {
		detail, err := h.forkEntries.Get(entry.ID)
		if err != nil {
			respondFitmentCatalogError(c, err)
			return
		}
		if detail == nil || !detail.IsEnabled {
			continue
		}
		responses = append(responses, newForkEntryResponse(detail))
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"fork_entries": responses,
		"pagination":   newPaginationResponse(page, pageSize, total),
	}})
}

func (h *Handler) GetForkEntry(c *gin.Context) {
	id, ok := parseID(c, "fork fitment entry ID")
	if !ok {
		return
	}

	entry, err := h.forkEntries.Get(id)
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}
	if entry == nil || !entry.IsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "fork fitment entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entry": newForkEntryResponse(entry)}})
}

func (h *Handler) ListHubSpecifications(c *gin.Context) {
	page, pageSize := parsePagination(c)
	enabled := true
	specifications, total, err := h.hubSpecifications.List(service.FitmentHubSpecificationListInput{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		Position:  fitmentcatalogdomain.HubPosition(normalizeQueryToken(c.Query("position"))),
		AxleType:  fitmentcatalogdomain.HubAxleType(normalizeQueryToken(c.Query("axle_type"))),
		IsEnabled: &enabled,
	})
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}

	responses := make([]hubSpecificationResponse, 0, len(specifications))
	for _, specification := range specifications {
		if !specification.IsEnabled {
			continue
		}
		responses = append(responses, newHubSpecificationResponse(specification))
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"hub_specifications": responses,
		"pagination":         newPaginationResponse(page, pageSize, total),
	}})
}

func (h *Handler) GetHubSpecification(c *gin.Context) {
	id, ok := parseID(c, "hub specification ID")
	if !ok {
		return
	}

	specification, err := h.hubSpecifications.Get(id)
	if err != nil {
		respondFitmentCatalogError(c, err)
		return
	}
	if specification == nil || !specification.IsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "hub specification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"hub_specification": newHubSpecificationResponse(*specification),
	}})
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func parseID(c *gin.Context, label string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + label})
		return 0, false
	}
	return uint(id), true
}

func parseYear(c *gin.Context) (*int, bool) {
	raw := strings.TrimSpace(c.Query("year"))
	if raw == "" {
		return nil, true
	}
	year, err := strconv.Atoi(raw)
	if err != nil || year < 1800 || year > 2200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year must be an integer between 1800 and 2200"})
		return nil, false
	}
	return &year, true
}

func normalizeQueryToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func respondFitmentCatalogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFrameFitmentEntryNotFound),
		errors.Is(err, service.ErrForkFitmentEntryNotFound),
		errors.Is(err, service.ErrFitmentHubSpecificationNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "fitment catalog record not found"})
	case errors.Is(err, service.ErrFrameFitmentEntryInvalid),
		errors.Is(err, service.ErrForkFitmentEntryInvalid),
		errors.Is(err, service.ErrFitmentHubSpecificationInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load fitment catalog"})
	}
}
