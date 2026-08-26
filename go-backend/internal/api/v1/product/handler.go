package product

import (
	"strings"

	"commerce-platform/internal/api/middleware"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	productService         *service.ProductService
	productCategoryService *service.ProductCategoryService
	mediaService           *service.MediaService
	storefrontContext      *service.StorefrontContextService
	reviewService          *service.ReviewService
	shippingService        *service.ShippingService
}

const (
	publicCatalogDefaultPageSize = 12
	publicCatalogMaxPageSize     = 24
)

func NewHandler(productService *service.ProductService) *Handler {
	return &Handler{
		productService: productService,
	}
}

func (h *Handler) ConfigureProductCategoryService(categoryService *service.ProductCategoryService) {
	if h == nil {
		return
	}
	h.productCategoryService = categoryService
}

func (h *Handler) ConfigureMediaService(mediaService *service.MediaService) {
	if h == nil {
		return
	}
	h.mediaService = mediaService
}

func (h *Handler) ConfigureStorefrontContext(contextService *service.StorefrontContextService) {
	if h == nil {
		return
	}
	h.storefrontContext = contextService
}

func (h *Handler) ConfigureReviewService(reviewService *service.ReviewService) {
	if h == nil {
		return
	}
	h.reviewService = reviewService
}

func (h *Handler) ConfigureShippingService(shippingService *service.ShippingService) {
	if h == nil {
		return
	}
	h.shippingService = shippingService
}

func (h *Handler) ListProducts(c *gin.Context) {
	publicContext := h.resolvePublicContext(c)
	locale := publicContext.Locale
	featured := c.Query("featured") == "true"
	categorySlug := strings.TrimSpace(c.Query("product_category"))
	params := pagination.ParsePagination(c)

	if c.Query("page_size") == "" {
		params.PageSize = publicCatalogDefaultPageSize
	}
	if params.PageSize > publicCatalogMaxPageSize {
		params.PageSize = publicCatalogMaxPageSize
	}

	var products []productdomain.Product
	var total int64
	var err error
	if categorySlug == "" {
		products, total, err = h.productService.ListPublic(
			locale,
			featured,
			params.Page,
			params.PageSize+1,
		)
	} else {
		var featuredFilter *bool
		if featured {
			featuredFilter = &featured
		}
		products, total, err = h.productService.SearchPublic(service.ProductSearchInput{
			Locale:       locale,
			Featured:     featuredFilter,
			CategorySlug: categorySlug,
			Page:         params.Page,
			PageSize:     params.PageSize + 1,
		})
	}
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	publicProducts, hasMore := trimPublicProductPage(products, params.PageSize)
	publicProductResponses := PublicProductsFromDomainWithLocaleAndDisplayCurrency(publicProducts, publicContext.DisplayCurrency, locale, h.mediaService)
	h.attachReviewSummaries(publicProductResponses)
	c.JSON(200, gin.H{
		"code":      0,
		"data":      publicProductResponses,
		"context":   publicContext.Response,
		"page":      params.Page,
		"page_size": params.PageSize,
		"total":     total,
		"has_more":  hasMore,
	})
}

func (h *Handler) attachReviewSummaries(items []PublicProduct) {
	if h == nil || h.reviewService == nil || len(items) == 0 {
		return
	}

	productIDs := make([]uint, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		productIDs = append(productIDs, item.ID)
	}
	if len(productIDs) == 0 {
		return
	}

	summaries, err := h.reviewService.GetReviewSummaries(productIDs)
	if err != nil {
		return
	}
	for index := range items {
		if summary, ok := summaries[items[index].ID]; ok {
			items[index].ReviewSummary = PublicProductReviewSummaryFromDomain(&summary)
		}
	}
}

func (h *Handler) GetProduct(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("id"))
	publicContext := h.resolvePublicContext(c)
	locale := publicContext.Locale

	product, translationRoutes, err := h.productService.GetPublicBySlugWithRoutesContext(c.Request.Context(), slug, locale)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	publicProduct := PublicProductFromDomainWithLocaleAndRoutes(*product, publicContext.DisplayCurrency, publicContext.Locale, translationRoutes, h.mediaService)
	if h.productCategoryService != nil {
		if breadcrumb, breadcrumbErr := h.productCategoryService.BuildProductBreadcrumb(*product, publicContext.Locale); breadcrumbErr == nil {
			publicProduct.Breadcrumb = breadcrumb
		}
	}
	if h.reviewService != nil {
		if summary, summaryErr := h.reviewService.GetReviewSummary(product.ID); summaryErr == nil {
			publicProduct.ReviewSummary = PublicProductReviewSummaryFromDomain(summary)
		}
	}
	publicProduct.ShippingDetails = h.resolvePublicShippingDetails(*product, publicContext)

	c.JSON(200, gin.H{
		"code":    0,
		"data":    publicProduct,
		"context": publicContext.Response,
	})
}

func (h *Handler) resolvePublicShippingDetails(item productdomain.Product, publicContext publicProductContext) *PublicProductShippingDetails {
	if h == nil || h.shippingService == nil || publicContext.Response == nil {
		return nil
	}

	country := strings.ToUpper(strings.TrimSpace(publicContext.Response.Country.Code))
	if country == "" || country == "ZZ" {
		return nil
	}

	variant := item.DefaultVariant()
	if variant == nil || variant.Weight <= 0 {
		return nil
	}

	currency := strings.TrimSpace(publicContext.Response.Currency.Base)
	if currency == "" {
		currency = item.DisplayPriceCurrency()
	}

	quote, err := h.shippingService.QuoteCart(service.ShippingQuoteInput{
		Country:         country,
		Currency:        currency,
		DisplayCurrency: publicContext.DisplayCurrency,
		Items: []service.ShippingQuoteItemInput{
			{
				ProductID: item.ID,
				VariantID: &variant.ID,
				Quantity:  1,
			},
		},
	})
	if err != nil || quote == nil || quote.SelectedOption == nil {
		return nil
	}

	option := quote.SelectedOption
	if option.EtaMinDays <= 0 && option.EtaMaxDays <= 0 {
		return nil
	}

	amount := option.ShippingFee
	amountCurrency := option.Currency
	if option.DisplayPrice != nil {
		amount = option.DisplayPrice.Amount
		amountCurrency = option.DisplayPrice.Currency
	}
	if strings.TrimSpace(amountCurrency) == "" {
		amountCurrency = currency
	}

	return &PublicProductShippingDetails{
		Country:      country,
		Amount:       amount,
		Currency:     strings.ToUpper(strings.TrimSpace(amountCurrency)),
		FreeShipping: option.FreeShipping,
		EtaMinDays:   option.EtaMinDays,
		EtaMaxDays:   option.EtaMaxDays,
	}
}

func (h *Handler) GetFilterableAttributes(c *gin.Context) {
	attrs, err := h.productService.GetFilterableAttributes()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    PublicProductAttributesFromDomain(attrs),
	})
}

func (h *Handler) ListProductSpecificationTemplates(c *gin.Context) {
	productSpecificationTemplates, err := h.productService.ListPublicProductSpecificationTemplates(false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, PublicProductSpecificationTemplatesFromDomainWithLocale(productSpecificationTemplates, middleware.GetLocale(c)))
}

func (h *Handler) ListCategories(c *gin.Context) {
	if h == nil || h.productCategoryService == nil {
		c.JSON(500, gin.H{
			"code":    apierror.ErrCodeInternal,
			"message": "product category service is unavailable",
		})
		return
	}

	locale := middleware.GetLocale(c)
	categories, err := h.productCategoryService.ListPublic(locale)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"code":   0,
		"data":   PublicProductCategoryListWithMedia(categories, h.mediaService),
		"locale": locale,
	})
}

func (h *Handler) GetCategory(c *gin.Context) {
	if h == nil || h.productCategoryService == nil {
		c.JSON(500, gin.H{
			"code":    apierror.ErrCodeInternal,
			"message": "product category service is unavailable",
		})
		return
	}

	slug := strings.TrimSpace(c.Param("slug"))
	category, err := h.productCategoryService.GetPublicBySlug(slug, middleware.GetLocale(c))
	if err != nil {
		apierror.RespondNotFound(c, "Product category")
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": PublicProductCategoryViewWithMedia(*category, h.mediaService),
	})
}

func trimPublicProductPage(products []productdomain.Product, pageSize int) ([]productdomain.Product, bool) {
	if pageSize < 1 || len(products) <= pageSize {
		return products, false
	}
	return products[:pageSize], true
}
