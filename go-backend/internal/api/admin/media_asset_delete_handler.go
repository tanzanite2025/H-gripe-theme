package admin

import (
	"net/http"

	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type mediaAssetDeleteRequest struct {
	Confirm string `json:"confirm"`
}

func (h *MediaHandler) DeleteAsset(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	var req mediaAssetDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "delete confirmation is required",
			"code":  "delete_confirmation_required",
		})
		return
	}
	if err := h.mediaService.DeleteAsset(c.Request.Context(), id, req.Confirm); err != nil {
		respondMediaError(c, err)
		return
	}
	response.NoContent(c)
}
