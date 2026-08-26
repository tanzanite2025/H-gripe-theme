package seo

import (
	"context"
	"errors"
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
	PushProduct(ctx context.Context, productID uint) (*service.GoogleIndexingPushResult, error)
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
				Message: "Google Indexing service is unavailable",
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
	if h.googleIndexing == nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Action:       seoAuditActionIndexing,
			Resource:     seoAuditResourceProduct,
			ResourceID:   uint(id),
			Status:       seoAuditStatusFailed,
			ErrorMessage: service.ErrGoogleIndexingNotConfigured.Error(),
			NewValue:     googleIndexingAuditValue(uint(id), nil),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": service.ErrGoogleIndexingNotConfigured.Error()})
		return
	}

	result, err := h.googleIndexing.PushProduct(c.Request.Context(), uint(id))
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Action:       seoAuditActionIndexing,
			Resource:     seoAuditResourceProduct,
			ResourceID:   uint(id),
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
			NewValue:     googleIndexingAuditValue(uint(id), result),
		})
		switch {
		case errors.Is(err, service.ErrGoogleIndexingProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrGoogleIndexingProductNotPublic):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrGoogleIndexingRecentlyNotified):
			retryAfter := 60
			var cooldownErr *service.GoogleIndexingCooldownError
			if errors.As(err, &cooldownErr) && cooldownErr.RetryAfter > 0 {
				retryAfter = int((cooldownErr.RetryAfter + time.Second - 1) / time.Second)
				if retryAfter < 1 {
					retryAfter = 1
				}
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusConflict, gin.H{
				"error":   "google_indexing_recently_notified",
				"message": err.Error(),
			})
		case errors.Is(err, service.ErrGoogleIndexingDisabled),
			errors.Is(err, service.ErrGoogleIndexingNotConfigured):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrGoogleIndexingProtection):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "google_indexing_protection_unavailable",
				"message": "Google Indexing duplicate protection is temporarily unavailable",
			})
		case errors.Is(err, service.ErrGoogleIndexingInvalidURL):
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrGoogleIndexingUpstream):
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt:  startedAt,
		Action:     seoAuditActionIndexing,
		Resource:   seoAuditResourceProduct,
		ResourceID: uint(id),
		Status:     seoAuditStatusOK,
		NewValue:   googleIndexingAuditValue(uint(id), result),
	})
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func googleIndexingAuditValue(productID uint, result *service.GoogleIndexingPushResult) map[string]interface{} {
	value := map[string]interface{}{
		"product_id":        productID,
		"notification_type": "URL_UPDATED",
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
