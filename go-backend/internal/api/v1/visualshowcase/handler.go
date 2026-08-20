package visualshowcase

import (
	"net/http"

	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	visualShowcaseService *service.VisualShowcaseService
	mediaResolver         publicmedia.Resolver
}

func NewHandler(visualShowcaseService *service.VisualShowcaseService, mediaResolvers ...publicmedia.Resolver) *Handler {
	var mediaResolver publicmedia.Resolver
	if len(mediaResolvers) > 0 {
		mediaResolver = mediaResolvers[0]
	}
	return &Handler{
		visualShowcaseService: visualShowcaseService,
		mediaResolver:         mediaResolver,
	}
}

func (h *Handler) Get(c *gin.Context) {
	showcaseKey := c.Param("showcase_key")
	locale := c.DefaultQuery("locale", "en")
	result, err := h.visualShowcaseService.GetPublishedResult(showcaseKey, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "visual showcase unavailable"})
		return
	}

	response.Success(c, gin.H{
		"showcase_key":     showcaseKey,
		"locale":           result.Locale,
		"requested_locale": result.RequestedLocale,
		"fallback":         result.Fallback,
		"configured_count": result.ConfiguredCount,
		"items":            publicItemsFromDomain(result.Items, h.mediaResolver),
	})
}
