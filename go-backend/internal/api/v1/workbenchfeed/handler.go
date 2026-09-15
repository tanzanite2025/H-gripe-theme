package workbenchfeed

import (
	"net/http"
	"strconv"
	"strings"

	feed "commerce-platform/internal/workbenchfeed"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *feed.Service
}

func NewHandler(service *feed.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "workbench feed is unavailable"})
		return
	}
	page, pageSize := parsePagination(c)
	result, err := h.service.List(feed.ListInput{
		Page:     page,
		PageSize: pageSize,
		Locale:   strings.TrimSpace(c.DefaultQuery("locale", feed.DefaultLocale)),
		Tag:      strings.TrimSpace(c.Query("tag")),
		Admin:    false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workbench feed is unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return page, pageSize
}
