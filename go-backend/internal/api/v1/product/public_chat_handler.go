package product

import (
	"strconv"
	"strings"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// ListPublicChatProducts powers storefront search and customer-service product
// selection. It shares the hardened catalog response contract so this legacy
// search path cannot become a more verbose data-extraction side channel.
func (h *Handler) ListPublicChatProducts(c *gin.Context) {
	publicContext := h.resolvePublicContext(c)
	locale := publicContext.Locale
	keyword := strings.TrimSpace(c.Query("keyword"))
	productSpecificationTemplateSlug := strings.TrimSpace(c.Query("product_specification_template"))
	categorySlug := strings.TrimSpace(c.Query("product_category"))
	brandSlug := strings.TrimSpace(c.Query("brand"))
	priceMin := parseOptionalFloatQuery(c, "price_min")
	priceMax := parseOptionalFloatQuery(c, "price_max")
	specFilters := parseSpecFilterQuery(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", "20")))
	compact := c.Query("compact") == "true" || c.Query("compact") == "1"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = publicCatalogDefaultPageSize
	}
	if pageSize > publicCatalogMaxPageSize {
		pageSize = publicCatalogMaxPageSize
	}

	searchInput := service.ProductSearchInput{
		Locale:                           locale,
		Keyword:                          keyword,
		ProductSpecificationTemplateSlug: productSpecificationTemplateSlug,
		CategorySlug:                     categorySlug,
		BrandSlug:                        brandSlug,
		PriceMin:                         priceMin,
		PriceMax:                         priceMax,
		SpecFilters:                      specFilters,
		Page:                             page,
		PageSize:                         pageSize,
	}
	var products []productdomain.Product
	var total int64
	var hasMore bool
	var err error
	if compact {
		searchInput.PageSize = pageSize + 1
		searchInput.OffsetPageSize = pageSize
		products, err = h.productService.SearchPublicCompact(searchInput)
	} else {
		products, total, err = h.productService.SearchPublic(searchInput)
	}
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	publicProducts := products
	if compact {
		publicProducts, hasMore = trimPublicProductPage(products, pageSize)
	} else {
		hasMore = hasMorePublicProductPage(total, page, pageSize)
	}
	publicProductResponses := PublicProductsFromDomainWithLocaleAndDisplayCurrency(publicProducts, publicContext.DisplayCurrency, locale, h.mediaService)
	if !compact {
		h.attachReviewSummaries(publicProductResponses)
	}
	payload := gin.H{
		"code":      0,
		"data":      publicProductResponses,
		"context":   publicContext.Response,
		"page":      page,
		"page_size": pageSize,
		"has_more":  hasMore,
	}
	if !compact {
		payload["total"] = total
	}
	c.JSON(200, payload)
}

func makePublicChatProduct(item productdomain.Product) PublicProduct {
	return PublicProductFromDomain(item)
}
