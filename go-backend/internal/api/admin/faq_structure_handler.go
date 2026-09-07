package admin

import (
	"commerce-platform/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListStructure 获取 FAQ 页面与页面内 FAQ 结构
// GET /api/admin/faqs/structure
func (h *FAQHandler) ListStructure(c *gin.Context) {
	locale := c.Query("locale")
	if locale == "" {
		c.JSON(http.StatusOK, gin.H{
			"pages": []service.FAQPageAdminView{},
		})
		return
	}

	pages, err := h.faqService.ListAdminStructure(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get FAQ structure"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pages": pages,
	})
}

// UpdatePage 更新 FAQ 页面元信息
// PUT /api/admin/faqs/pages/:page_id
func (h *FAQHandler) UpdatePage(c *gin.Context) {
	pageID := c.Param("page_id")
	var req faqPageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := h.faqService.UpsertAdminPage(pageID, service.FAQPageAdminInput{
		RoutePath: req.RoutePath,
		Domain:    req.Domain,
		Locale:    req.Locale,
		Title:     req.Title,
		Subtitle:  req.Subtitle,
		Status:    req.Status,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		if isFAQValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FAQ page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ page updated successfully",
		"page":    page,
	})
}
