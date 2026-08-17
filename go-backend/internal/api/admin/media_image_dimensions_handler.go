package admin

import (
	"net/http"

	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type MediaImageDimensionsHandler struct {
	engine *service.MediaImageDimensionEngine
}

func NewMediaImageDimensionsHandler(
	engine *service.MediaImageDimensionEngine,
) *MediaImageDimensionsHandler {
	return &MediaImageDimensionsHandler{engine: engine}
}

func (h *MediaImageDimensionsHandler) List(c *gin.Context) {
	if h == nil || h.engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image dimension engine is unavailable"})
		return
	}

	params := pagination.ParsePagination(c)
	result, err := h.engine.List(service.MediaImageDimensionListInput{
		Page:     params.Page,
		PageSize: params.PageSize,
		Search:   c.Query("search"),
		State:    c.DefaultQuery("state", service.MediaImageDimensionStateAttention),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response.Paged(c, gin.H{
		"items":   result.Items,
		"summary": result.Summary,
		"presets": result.Presets,
	}, params.Page, params.PageSize, result.Total)
}

func (h *MediaImageDimensionsHandler) Reconcile(c *gin.Context) {
	if h == nil || h.engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image dimension engine is unavailable"})
		return
	}

	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}
	item, err := h.engine.Reconcile(c.Request.Context(), id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"item": item})
}
