package seo

import (
	"net/http"
	"strconv"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type googleIndexingService interface {
	Status() service.GoogleIndexingStatus
}

type ProductsHandler struct {
	seoResources   *service.SEOResourceService
	googleIndexing googleIndexingService
	audit          seoAuditRecorder
}

func NewProductsHandler(seoResources *service.SEOResourceService) *ProductsHandler {
	return &ProductsHandler{seoResources: seoResources}
}

func (h *ProductsHandler) ConfigureGoogleIndexingService(indexing *service.GoogleIndexingService) {
	if h == nil {
		return
	}
	if indexing == nil {
		h.googleIndexing = nil
		return
	}
	h.googleIndexing = indexing
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

func (h *ProductsHandler) IndexingStatus(c *gin.Context) {
	if h.googleIndexing == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": service.GoogleIndexingStatus{
				Enabled: false,
				Ready:   false,
				Message: "Google Indexing API is not supported for product pages; use the sitemap workflow instead",
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": h.googleIndexing.Status()})
}

func (h *ProductsHandler) PushIndexing(c *gin.Context) {
	startedAt := time.Now().UTC()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Action:       seoAuditActionIndexing,
			Resource:     seoAuditResourceProduct,
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	err = service.ErrGoogleIndexingProductUnsupported
	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt:    startedAt,
		Action:       seoAuditActionIndexing,
		Resource:     seoAuditResourceProduct,
		ResourceID:   uint(id),
		Status:       seoAuditStatusFailed,
		ErrorMessage: err.Error(),
		NewValue:     googleIndexingAuditValue(uint(id), nil),
	})
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"error":   "google_indexing_product_unsupported",
		"message": err.Error(),
	})
}

func googleIndexingAuditValue(productID uint, result *service.GoogleIndexingPushResult) map[string]interface{} {
	value := map[string]interface{}{
		"product_id": productID,
	}
	if result == nil {
		return value
	}
	value["url"] = result.URL
	value["notification_type"] = result.NotificationType
	value["http_status"] = result.HTTPStatus
	value["accepted"] = result.Accepted
	if !result.SubmittedAt.IsZero() {
		value["submitted_at"] = result.SubmittedAt
	}
	return value
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
