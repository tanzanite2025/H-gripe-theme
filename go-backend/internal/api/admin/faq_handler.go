package admin

import (
	"commerce-platform/internal/service"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type FAQHandler struct {
	faqService   *service.FAQService
	routeCatalog *service.StorefrontRouteCatalogService
}

func NewFAQHandler(faqService *service.FAQService, routeCatalog ...*service.StorefrontRouteCatalogService) *FAQHandler {
	var catalog *service.StorefrontRouteCatalogService
	if len(routeCatalog) > 0 {
		catalog = routeCatalog[0]
	}
	return &FAQHandler{
		faqService:   faqService,
		routeCatalog: catalog,
	}
}

// SyncRoutes refreshes the Nuxt route manifest and reconciles FAQ route
// metadata. It is safe to run repeatedly and never deletes FAQ answers.
func (h *FAQHandler) SyncRoutes(c *gin.Context) {
	if h.routeCatalog == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storefront route catalog is unavailable"})
		return
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	summary, err := h.routeCatalog.Sync(ctx)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func isFAQValidationError(err error) bool {
	message := err.Error()
	return errors.Is(err, service.ErrUnsupportedLocale) ||
		errors.Is(err, service.ErrFAQLocaleImmutable) ||
		strings.Contains(message, "required") ||
		strings.Contains(message, "does not exist") ||
		strings.Contains(message, "must be created") ||
		strings.Contains(message, "hidden") ||
		strings.Contains(message, "answer image")
}
