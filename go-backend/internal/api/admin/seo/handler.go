package seo

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func localeFromQuery(c *gin.Context) string {
	return c.DefaultQuery("locale", "en")
}

func writeGetError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func writeUpdateError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidSEOSettings) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func resourcePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return page, pageSize
}

func writeResourceError(c *gin.Context, err error, notFound error) {
	if errors.Is(err, notFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "SEO resource not found"})
		return
	}
	if errors.Is(err, service.ErrInvalidSEOCanonicalURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "canonical_url must be an absolute same-site URL without query or fragment"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
