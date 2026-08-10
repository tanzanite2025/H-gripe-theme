package seo

import (
	"net/http"
	"strconv"
	"time"

	productdomain "tanzanite/internal/domain/product"
	seodomain "tanzanite/internal/domain/seo"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductsHandler struct {
	seoResources *service.SEOResourceService
	audit        seoAuditRecorder
}

func NewProductsHandler(seoResources *service.SEOResourceService) *ProductsHandler {
	return &ProductsHandler{seoResources: seoResources}
}

func (h *ProductsHandler) ConfigureAuditService(recorder seoAuditRecorder) {
	if h == nil {
		return
	}
	h.audit = recorder
}

func (h *ProductsHandler) Get(c *gin.Context) {
	page, pageSize := resourcePagination(c)
	products, total, err := h.seoResources.ListProducts(
		page,
		pageSize,
		c.Query("status"),
		c.Query("locale"),
		c.Query("search"),
	)
	if err != nil {
		writeResourceError(c, err, service.ErrProductNotFound)
		return
	}

	items := make([]gin.H, 0, len(products))
	for _, product := range products {
		diagnostics, err := h.seoResources.ProductDiagnostics(product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, gin.H{
			"id":               product.ID,
			"name":             product.Name,
			"slug":             product.Slug,
			"route_path":       seodomain.BuildProductRoute(product.Locale, product.Slug).Path,
			"status":           product.Status,
			"locale":           product.Locale,
			"meta_title":       product.MetaTitle,
			"meta_description": product.MetaDesc,
			"diagnostics":      diagnostics,
			"created_at":       product.CreatedAt,
		})
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

func (h *ProductsHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var request seodomain.ProductResourceUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startedAt := time.Now().UTC()
	existing, _ := h.seoResources.GetProduct(uint(id))
	product, err := h.seoResources.UpdateProduct(uint(id), request)
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Resource:     seoAuditResourceProduct,
			ResourceID:   uint(id),
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     productAuditValue(existing),
			Changes: seoFieldChanges(map[string]interface{}{
				"meta_title":       request.MetaTitle != nil,
				"meta_description": request.MetaDescription != nil,
			}),
		})
		writeResourceError(c, err, service.ErrProductNotFound)
		return
	}

	diagnostics, err := h.seoResources.ProductDiagnostics(*product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt:  startedAt,
		Resource:   seoAuditResourceProduct,
		ResourceID: product.ID,
		Status:     seoAuditStatusOK,
		Changes: seoFieldChanges(map[string]interface{}{
			"meta_title":       request.MetaTitle != nil,
			"meta_description": request.MetaDescription != nil,
		}),
		OldValue: productAuditValue(existing),
		NewValue: productAuditValue(product),
	})

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":               product.ID,
			"name":             product.Name,
			"slug":             product.Slug,
			"route_path":       seodomain.BuildProductRoute(product.Locale, product.Slug).Path,
			"status":           product.Status,
			"locale":           product.Locale,
			"meta_title":       product.MetaTitle,
			"meta_description": product.MetaDesc,
			"diagnostics":      diagnostics,
		},
	})
}

func productAuditValue(product *productdomain.Product) map[string]string {
	if product == nil {
		return nil
	}
	return map[string]string{
		"meta_title":       product.MetaTitle,
		"meta_description": product.MetaDesc,
	}
}
