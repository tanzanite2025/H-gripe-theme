package storefront

import (
	"net/http"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type RedirectsHandler struct {
	rules *service.StorefrontRedirectRuleService
}

func NewRedirectsHandler(rules *service.StorefrontRedirectRuleService) *RedirectsHandler {
	return &RedirectsHandler{rules: rules}
}

func (h *RedirectsHandler) ListPublished(c *gin.Context) {
	rules, err := h.rules.ListPublished()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storefront redirect policy is unavailable"})
		return
	}
	c.Header("Cache-Control", "private, max-age=15")
	c.JSON(http.StatusOK, gin.H{"data": rules})
}
