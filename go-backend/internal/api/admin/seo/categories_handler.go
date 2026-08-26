package seo

import (
	"net/http"
	"strconv"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoriesHandler struct {
	seoResources *service.SEOResourceService
	audit        seoAuditRecorder
}

func NewCategoriesHandler(seoResources *service.SEOResourceService) *CategoriesHandler {
	return &CategoriesHandler{seoResources: seoResources}
}

func (h *CategoriesHandler) ConfigureAuditService(recorder seoAuditRecorder) {
	if h == nil {
		return
	}
	h.audit = recorder
}

func (h *CategoriesHandler) Get(c *gin.Context) {
	page, pageSize := resourcePagination(c)
	categories, total, err := h.seoResources.ListCategories(
		page,
		pageSize,
		c.Query("locale"),
		c.Query("search"),
	)
	if err != nil {
		writeResourceError(c, err, service.ErrProductCategoryNotFound)
		return
	}

	items := make([]gin.H, 0, len(categories))
	for _, category := range categories {
		items = append(items, categoryResourcePayload(category))
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *CategoriesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var request seodomain.CategoryResourceUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Locale == "" {
		request.Locale = c.DefaultQuery("locale", "en")
	}

	startedAt := time.Now().UTC()
	existing, _ := h.seoResources.GetCategory(uint(id), request.Locale)
	category, err := h.seoResources.UpdateCategory(uint(id), request)
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Resource:     seoAuditResourceCategory,
			ResourceID:   uint(id),
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     categoryAuditValue(existing),
			Changes: seoFieldChanges(map[string]interface{}{
				"locale":           request.Locale,
				"meta_title":       request.MetaTitle != nil,
				"meta_description": request.MetaDescription != nil,
				"intro":            request.Intro != nil,
			}),
		})
		writeResourceError(c, err, service.ErrProductCategoryNotFound)
		return
	}

	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt:  startedAt,
		Resource:   seoAuditResourceCategory,
		ResourceID: category.ID,
		Status:     seoAuditStatusOK,
		Changes: seoFieldChanges(map[string]interface{}{
			"locale":           category.Locale,
			"meta_title":       request.MetaTitle != nil,
			"meta_description": request.MetaDescription != nil,
			"intro":            request.Intro != nil,
		}),
		OldValue: categoryAuditValue(existing),
		NewValue: categoryAuditValue(category),
	})

	c.JSON(http.StatusOK, gin.H{"data": categoryResourcePayload(*category)})
}

func categoryResourcePayload(category service.ProductCategorySEOView) gin.H {
	return gin.H{
		"id":               category.ID,
		"name":             category.Name,
		"slug":             category.Slug,
		"route_path":       category.RoutePath,
		"locale":           category.Locale,
		"status":           category.Status,
		"meta_title":       category.MetaTitle,
		"meta_description": category.MetaDescription,
		"intro":            category.Intro,
	}
}

func categoryAuditValue(category *service.ProductCategorySEOView) map[string]string {
	if category == nil {
		return nil
	}
	return map[string]string{
		"locale":           category.Locale,
		"meta_title":       category.MetaTitle,
		"meta_description": category.MetaDescription,
		"intro":            category.Intro,
	}
}
