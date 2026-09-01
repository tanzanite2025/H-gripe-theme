package content

import (
	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRefundReturnPolicy(c *gin.Context) {
	if h.refundReturnPolicyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and return policy service unavailable"})
		return
	}

	result, err := h.refundReturnPolicyService.GetPublic(c.DefaultQuery("locale", "en"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.publicRefundReturnPolicy(result))
}

func (h *Handler) publicRefundReturnPolicy(result service.RefundReturnPolicyResult) service.RefundReturnPolicyResult {
	for index := range result.Policy.Sections {
		if result.Policy.Sections[index].Image == nil {
			continue
		}
		image := *result.Policy.Sections[index].Image
		image.URL = publicmedia.URL(h.mediaService, image.URL)
		result.Policy.Sections[index].Image = &image
	}
	return result
}
