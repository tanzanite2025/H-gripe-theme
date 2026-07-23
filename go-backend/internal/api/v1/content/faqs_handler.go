package content

import (
	"net/http"
	"strconv"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

// ListFAQs 获取FAQ列表
func (h *Handler) ListFAQs(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	pageID := c.Query("page_id")
	category := c.Query("category")
	status := c.DefaultQuery("status", "published")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	faqs, total, err := h.faqService.List(locale, pageID, category, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        faqs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetFAQ 获取单个FAQ
func (h *Handler) GetFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	faq, err := h.faqService.GetPublicByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "faq not found"})
		return
	}

	c.JSON(http.StatusOK, faq)
}

// GetFAQCategories 获取FAQ分类列表
func (h *Handler) GetFAQCategories(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	categories, err := h.faqService.GetCategories(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// ListFAQPages 获取结构化 FAQ 页面数据
func (h *Handler) ListFAQPages(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	pages, err := h.faqService.ListPublicPageData(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages})
}

// GetFAQPage 获取单个页面的结构化 FAQ 数据
func (h *Handler) GetFAQPage(c *gin.Context) {
	pageID := c.Param("page_id")
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	page, err := h.faqService.GetPublicPageData(pageID, locale)
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "faq page not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// GetFAQPageByRoute 获取当前路由对应的结构化 FAQ 数据
func (h *Handler) GetFAQPageByRoute(c *gin.Context) {
	routePath := c.Query("route_path")
	if routePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route_path is required"})
		return
	}
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	page, err := h.faqService.GetPublicPageDataByRoutePath(routePath, locale)
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "faq page not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// SearchFAQs 搜索FAQ
func (h *Handler) SearchFAQs(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search keyword is required"})
		return
	}

	locale := middleware.GetLocale(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	faqs, total, err := h.faqService.Search(keyword, locale, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        faqs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		"keyword":     keyword,
	})
}

// GetFAQsByCategory 获取分类下的FAQ
func (h *Handler) GetFAQsByCategory(c *gin.Context) {
	category := c.Param("category")
	locale := middleware.GetLocale(c)

	faqs, err := h.faqService.GetByCategory(category, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"data":     faqs,
		"total":    len(faqs),
	})
}

// GetPopularFAQs 获取热门FAQ
func (h *Handler) GetPopularFAQs(c *gin.Context) {
	locale := middleware.GetLocale(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if limit < 1 || limit > 50 {
		limit = 10
	}

	faqs, err := h.faqService.GetPopular(locale, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  faqs,
		"total": len(faqs),
	})
}
