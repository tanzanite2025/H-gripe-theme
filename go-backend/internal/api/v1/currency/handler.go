package currency

import (
	"net/http"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	policyService *service.CurrencyPolicyService
}

func NewHandler(policyService *service.CurrencyPolicyService) *Handler {
	return &Handler{policyService: policyService}
}

func (h *Handler) GetPolicy(c *gin.Context) {
	policy, err := h.policyService.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policy})
}
