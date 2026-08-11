package content

import (
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/locales"

	"github.com/gin-gonic/gin"
)

// ListPosts 获取文章列表
func (h *Handler) ListPosts(c *gin.Context) {
	locale, ok := resolvePostLocale(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported locale"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.postService.ListPublic(locale, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        PublicPostsFromDomain(posts),
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetPost 获取单篇文章
func (h *Handler) GetPost(c *gin.Context) {
	idOrSlug := c.Param("id")

	if id, err := strconv.ParseUint(idOrSlug, 10, 32); err == nil {
		post, routes, err := h.postService.GetPublicByIDWithRoutes(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusOK, PublicPostFromDomainWithRoutes(*post, routes))
		return
	}

	locale, ok := resolvePostLocale(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported locale"})
		return
	}

	post, routes, err := h.postService.GetPublicBySlugWithRoutes(idOrSlug, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, PublicPostFromDomainWithRoutes(*post, routes))
}

func resolvePostLocale(c *gin.Context) (string, bool) {
	requested := strings.TrimSpace(c.Query("locale"))
	if requested == "" {
		return middleware.GetLocale(c), true
	}

	resolved := locales.ResolveSupported(requested)
	return resolved, resolved != ""
}
