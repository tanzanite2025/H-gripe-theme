package product

import (
	"strconv"
	"strings"

	"commerce-platform/internal/api/middleware"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// ListPublicChatProducts powers storefront search and customer-service product
// selection. It shares the hardened catalog response contract so this legacy
// search path cannot become a more verbose data-extraction side channel.
func (h *Handler) ListPublicChatProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
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
		PageSize:                         pageSize + 1,
	}
	var products []productdomain.Product
	var err error
	if compact {
		products, err = h.productService.SearchPublicCompact(searchInput)
	} else {
		products, _, err = h.productService.SearchPublic(searchInput)
	}
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	publicProducts, hasMore := trimPublicProductPage(products, pageSize)
	publicProductResponses := PublicProductsFromDomainWithLocale(publicProducts, locale, h.mediaService)
	if !compact {
		h.attachReviewSummaries(publicProductResponses)
	}
	c.JSON(200, gin.H{
		"code":      0,
		"data":      publicProductResponses,
		"page_size": pageSize,
		"has_more":  hasMore,
	})
}

func makePublicChatProduct(item productdomain.Product) PublicProduct {
	return PublicProductFromDomain(item)
}
