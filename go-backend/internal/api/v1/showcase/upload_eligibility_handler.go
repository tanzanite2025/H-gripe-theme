package showcase

import (
	"errors"
	"net/http"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *ShowcaseHandler) ListUploadOrders(c *gin.Context) {
	userID, ok := showcaseAuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}
	if h == nil || h.uploadEligibility == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload eligibility is temporarily unavailable"})
		return
	}

	orders, err := h.uploadEligibility.ListOrderOptions(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrShowcaseUploadEligibilityUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Upload eligibility is temporarily unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load upload orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
