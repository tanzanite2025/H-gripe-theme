package admin

import (
	"net/http"

	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type mediaAssetUpdateRequest struct {
	Alt        string `json:"alt"`
	Caption    string `json:"caption"`
	Status     string `json:"status"`
	Visibility string `json:"visibility"`
}

func (h *MediaHandler) UpdateAsset(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	var req mediaAssetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media asset payload"})
		return
	}

	asset, err := h.mediaService.UpdateAsset(id, service.MediaAssetUpdateInput{
		Alt:        req.Alt,
		Caption:    req.Caption,
		Status:     req.Status,
		Visibility: req.Visibility,
	})
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"asset": asset})
}
