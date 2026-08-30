package urlmanagement

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoutesHandler struct {
	catalog *service.StorefrontRouteCatalogService
}

func NewRoutesHandler(catalog *service.StorefrontRouteCatalogService) *RoutesHandler {
	return &RoutesHandler{catalog: catalog}
}

func (h *RoutesHandler) Stats(c *gin.Context) {
	stats, err := h.catalog.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *RoutesHandler) List(c *gin.Context) {
	page, pageSize := urlPagination(c)
	filter := routeCatalogFilter(c)
	filter.Page = page
	filter.PageSize = pageSize
	entries, total, err := h.catalog.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": entries,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *RoutesHandler) Get(c *gin.Context) {
	id, err := parseRouteEntryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.catalog.Get(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entry})
}

func (h *RoutesHandler) History(c *gin.Context) {
	id, err := parseRouteEntryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page, pageSize := urlPagination(c)
	results, total, err := h.catalog.ListChecks(id, page, pageSize)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": results,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *RoutesHandler) Sync(c *gin.Context) {
	summary, err := h.catalog.Sync(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *RoutesHandler) Sitemap(c *gin.Context) {
	overview, err := h.catalog.SitemapOverview()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Header("Cache-Control", "private, max-age=15")
	c.JSON(http.StatusOK, gin.H{"data": overview})
}

func (h *RoutesHandler) SyncSitemap(c *gin.Context) {
	syncSummary, err := h.catalog.Sync(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	overview, err := h.catalog.SitemapOverview()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"sync":    syncSummary,
			"sitemap": overview,
		},
	})
}

func (h *RoutesHandler) CheckOne(c *gin.Context) {
	id, err := parseRouteEntryID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.catalog.CheckEntry(contextOrBackground(c), id)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *RoutesHandler) Check(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	summary, err := h.catalog.Check(contextOrBackground(c), routeCatalogFilter(c), limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func routeCatalogFilter(c *gin.Context) repository.StorefrontRouteCatalogListFilter {
	searchProfileStatus := strings.ToLower(strings.TrimSpace(c.Query("search_profile_status")))
	switch searchProfileStatus {
	case "configured", "unconfigured":
	default:
		searchProfileStatus = ""
	}

	return repository.StorefrontRouteCatalogListFilter{
		Locale:              c.Query("locale"),
		SourceType:          c.Query("source_type"),
		EntryStatus:         c.Query("entry_status"),
		CheckStatus:         c.Query("check_status"),
		Search:              c.Query("search"),
		Searchable:          parseOptionalBool(c.Query("searchable")),
		SearchProfileStatus: searchProfileStatus,
		NeedsAttention:      parseOptionalBool(c.Query("needs_attention")),
		ProblemScope:        c.Query("problem_scope"),
		ExcludeAlias:        c.Query("include_aliases") != "true",
	}
}

func urlPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	return page, pageSize
}

func parseRouteEntryID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid route entry ID")
	}
	return uint(parsed), nil
}

func parseOptionalBool(value string) *bool {
	switch value {
	case "true":
		result := true
		return &result
	case "false":
		result := false
		return &result
	default:
		return nil
	}
}

func contextOrBackground(c *gin.Context) context.Context {
	if c != nil && c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context()
	}
	return context.Background()
}
