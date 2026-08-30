package storefront

import (
	"net/http"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type URLSearchHandler struct {
	profiles *service.StorefrontURLSearchProfileService
}

func NewURLSearchHandler(profiles *service.StorefrontURLSearchProfileService) *URLSearchHandler {
	return &URLSearchHandler{profiles: profiles}
}

func (h *URLSearchHandler) List(c *gin.Context) {
	locale := c.Query("locale")
	entries, err := h.profiles.PublicIndex(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, gin.H{
		"items": entries,
		"total": len(entries),
	})
}
