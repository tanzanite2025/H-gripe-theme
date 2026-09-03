package admin

import (
	refundcancellationdomain "commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RefundCancellationPolicyHandler struct {
	policyService *service.RefundCancellationPolicyService
}

func NewRefundCancellationPolicyHandler(policyService *service.RefundCancellationPolicyService) *RefundCancellationPolicyHandler {
	return &RefundCancellationPolicyHandler{policyService: policyService}
}

func (h *RefundCancellationPolicyHandler) Get(c *gin.Context) {
	if h == nil || h.policyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and cancellation policy service unavailable"})
		return
	}
	result, err := h.policyService.GetAdmin(c.DefaultQuery("locale", "en"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *RefundCancellationPolicyHandler) Update(c *gin.Context) {
	if h == nil || h.policyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and cancellation policy service unavailable"})
		return
	}

	var request struct {
		Locale string                          `json:"locale"`
		Policy refundcancellationdomain.Policy `json:"policy"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.policyService.Update(strings.TrimSpace(request.Locale), request.Policy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
