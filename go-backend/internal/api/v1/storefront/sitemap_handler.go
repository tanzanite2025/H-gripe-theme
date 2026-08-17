package storefront

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/seo"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type SitemapHandler struct {
	catalog *service.StorefrontRouteCatalogService
}

func NewSitemapHandler(catalog *service.StorefrontRouteCatalogService) *SitemapHandler {
	return &SitemapHandler{catalog: catalog}
}

type SitemapRoute struct {
	Loc        string  `json:"loc"`
	LastMod    string  `json:"lastmod,omitempty"`
	ChangeFreq string  `json:"changefreq"`
	Priority   float64 `json:"priority"`
}

func (h *SitemapHandler) List(c *gin.Context) {
	limit := sitemapLimit(c)
	entries, err := h.catalog.ListSitemapEntries(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	routes := make([]SitemapRoute, 0, len(entries))
	for _, entry := range entries {
		if route := sitemapRouteForEntry(entry); route.Loc != "" {
			routes = append(routes, route)
		}
	}

	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, gin.H{
		"items": routes,
		"total": len(routes),
	})
}

func sitemapLimit(c *gin.Context) int {
	limit, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("limit", "5000")))
	if err != nil || limit < 1 {
		return 5000
	}
	if limit > 50000 {
		return 50000
	}
	return limit
}

func sitemapRouteForEntry(entry seo.StorefrontRouteCatalogEntry) SitemapRoute {
	if strings.TrimSpace(entry.Path) == "" {
		return SitemapRoute{}
	}

	route := SitemapRoute{
		Loc:        entry.Path,
		ChangeFreq: "monthly",
		Priority:   0.6,
	}
	switch entry.SourceType {
	case seo.RouteSourceProduct:
		route.ChangeFreq = "weekly"
		route.Priority = 0.8
	case seo.RouteSourceBlog:
		route.ChangeFreq = "monthly"
		route.Priority = 0.6
	case seo.RouteSourceStatic:
		route.ChangeFreq = "monthly"
		route.Priority = 0.7
	case seo.RouteSourceAlias:
		route.ChangeFreq = "yearly"
		route.Priority = 0.1
	}
	if !entry.UpdatedAt.IsZero() {
		route.LastMod = entry.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return route
}
