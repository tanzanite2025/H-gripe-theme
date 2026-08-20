package admin

import (
	refundreturndomain "commerce-platform/internal/domain/refundreturn"
	"commerce-platform/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RefundReturnPolicyHandler struct {
	policyService *service.RefundReturnPolicyService
}

func NewRefundReturnPolicyHandler(policyService *service.RefundReturnPolicyService) *RefundReturnPolicyHandler {
	return &RefundReturnPolicyHandler{policyService: policyService}
}

func (h *RefundReturnPolicyHandler) Get(c *gin.Context) {
	if h == nil || h.policyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and return policy service unavailable"})
		return
	}
	result, err := h.policyService.GetAdmin(c.DefaultQuery("locale", "en"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *RefundReturnPolicyHandler) Update(c *gin.Context) {
	if h == nil || h.policyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund and return policy service unavailable"})
		return
	}

	var request struct {
		Locale string                    `json:"locale"`
		Policy refundreturndomain.Policy `json:"policy"`
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
