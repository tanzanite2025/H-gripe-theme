package content

import (
	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRefundCancellationPolicy(c *gin.Context) {
	if h.refundCancellationPolicyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and cancellation policy service unavailable"})
		return
	}

	result, err := h.refundCancellationPolicyService.GetPublic(c.DefaultQuery("locale", "en"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.publicRefundCancellationPolicy(result))
}

func (h *Handler) publicRefundCancellationPolicy(result service.RefundCancellationPolicyResult) service.RefundCancellationPolicyResult {
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
