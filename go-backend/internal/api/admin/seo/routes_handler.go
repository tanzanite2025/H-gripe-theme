package seo

import (
	"context"
	"errors"
	"net/http"
	"strconv"

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
	page, pageSize := resourcePagination(c)
	searchable := parseOptionalBool(c.Query("searchable"))
	entries, total, err := h.catalog.List(repository.StorefrontRouteCatalogListFilter{
		Page:         page,
		PageSize:     pageSize,
		Locale:       c.Query("locale"),
		SourceType:   c.Query("source_type"),
		EntryStatus:  c.Query("entry_status"),
		CheckStatus:  c.Query("check_status"),
		Search:       c.Query("search"),
		Searchable:   searchable,
		ExcludeAlias: c.Query("include_aliases") != "true",
	})
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
	page, pageSize := resourcePagination(c)
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
	summary, err := h.catalog.Check(contextOrBackground(c), repository.StorefrontRouteCatalogListFilter{
		Locale:       c.Query("locale"),
		SourceType:   c.Query("source_type"),
		EntryStatus:  c.Query("entry_status"),
		CheckStatus:  c.Query("check_status"),
		Search:       c.Query("search"),
		Searchable:   parseOptionalBool(c.Query("searchable")),
		ExcludeAlias: c.Query("include_aliases") != "true",
	}, limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
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
