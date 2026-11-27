package admin

import (
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type productCreateRequest struct {
	MerchantProductIdentifier     string `json:"merchant_product_identifier"`
	ManufacturerProductIdentifier string `json:"manufacturer_product_identifier"`
	StandardProductIdentifier     string `json:"standard_product_identifier"`

	ProductType                    *string                                   `json:"product_type"`
	CarbonRimSpecification         *productdomain.CarbonRimSpecification     `json:"carbon_rim_specification"`
	BicycleTireSpecification       *productdomain.BicycleTireSpecification   `json:"bicycle_tire_specification"`
	WheelsetSpecification          *productdomain.WheelsetSpecification      `json:"wheelset_specification"`
	SpokeSpecification             *productdomain.SpokesProductSpecification `json:"spoke_specification"`
	GenericProductSchema           *productdomain.GenericProductSchema       `json:"generic_product_schema"`
	ProductCategoryID              *uint                                     `json:"product_category_id"`
	BrandID                        *uint                                     `json:"brand_id"`
	AfterSalesTemplateID           *uint                                     `json:"after_sales_template_id"`
	PackagingTemplateID            *uint                                     `json:"packaging_template_id"`
	CustomsClassificationProfileID *uint                                     `json:"customs_classification_profile_id"`
	HSCode                         string                                    `json:"hs_code"`
	CNCode                         string                                    `json:"cn_code"`
	CountryOfOrigin                string                                    `json:"country_of_origin"`
	CustomsDescription             string                                    `json:"customs_description"`
	Name                           string                                    `json:"name" binding:"required"`
	Slug                           string                                    `json:"slug" binding:"required"`
	Description                    string                                    `json:"description"`
	ShortDesc                      string                                    `json:"short_description"`
	Currency                       string                                    `json:"currency"`
	FulfillmentMode                string                                    `json:"fulfillment_mode"`
	Status                         string                                    `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	Locale                         string                                    `json:"locale"`
	ParentID                       *uint                                     `json:"parent_id"`
	Featured                       bool                                      `json:"featured"`
	Variants                       []productVariantRequest                   `json:"variants"`
	Media                          []productMediaRequest                     `json:"media"`
	SpokeRepairKitModelKeys        []string                                  `json:"spoke_repair_kit_model_keys"`
}

type productUpdateRequest struct {
	MerchantProductIdentifier     *string `json:"merchant_product_identifier"`
	ManufacturerProductIdentifier *string `json:"manufacturer_product_identifier"`
	StandardProductIdentifier     *string `json:"standard_product_identifier"`

	ProductType                    *string                                   `json:"product_type"`
	CarbonRimSpecification         *productdomain.CarbonRimSpecification     `json:"carbon_rim_specification"`
	BicycleTireSpecification       *productdomain.BicycleTireSpecification   `json:"bicycle_tire_specification"`
	WheelsetSpecification          *productdomain.WheelsetSpecification      `json:"wheelset_specification"`
	SpokeSpecification             *productdomain.SpokesProductSpecification `json:"spoke_specification"`
	GenericProductSchema           *productdomain.GenericProductSchema       `json:"generic_product_schema"`
	ProductCategoryID              *uint                                     `json:"product_category_id"`
	BrandID                        *uint                                     `json:"brand_id"`
	AfterSalesTemplateID           *uint                                     `json:"after_sales_template_id"`
	PackagingTemplateID            *uint                                     `json:"packaging_template_id"`
	CustomsClassificationProfileID *uint                                     `json:"customs_classification_profile_id"`
	HSCode                         *string                                   `json:"hs_code"`
	CNCode                         *string                                   `json:"cn_code"`
	CountryOfOrigin                *string                                   `json:"country_of_origin"`
	CustomsDescription             *string                                   `json:"customs_description"`
	Name                           *string                                   `json:"name" binding:"omitempty,min=1"`
	Slug                           *string                                   `json:"slug" binding:"omitempty,min=1"`
	Description                    *string                                   `json:"description"`
	ShortDesc                      *string                                   `json:"short_description"`
	Currency                       *string                                   `json:"currency"`
	FulfillmentMode                *string                                   `json:"fulfillment_mode"`
	Status                         *string                                   `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
	Locale                         *string                                   `json:"locale"`
	ParentID                       *uint                                     `json:"parent_id"`
	Featured                       *bool                                     `json:"featured"`
	Variants                       []productVariantRequest                   `json:"variants"`
	Media                          []productMediaRequest                     `json:"media"`
	SpokeRepairKitModelKeys        []string                                  `json:"spoke_repair_kit_model_keys"`
}

type productVariantRequest struct {
	CarbonRimSpecification *productdomain.CarbonRimVariantSpecification `json:"carbon_rim_specification"`
	ID                     *uint                                        `json:"id"`
	ShippingTemplateIDs    []uint                                       `json:"shipping_template_ids"`
	SKU                    string                                       `json:"sku"`
	Title                  string                                       `json:"title"`
	Currency               string                                       `json:"currency"`
	PriceMinor             int64                                        `json:"price_minor" binding:"required"`
	Stock                  int                                          `json:"stock"`
	Weight                 int                                          `json:"weight_grams"`
	IsDefault              bool                                         `json:"is_default"`
	IsActive               *bool                                        `json:"is_active"`
	SortOrder              int                                          `json:"sort_order"`
}

type productMediaRequest struct {
	ID           *uint  `json:"id"`
	VariantID    *uint  `json:"variant_id"`
	MediaAssetID *uint  `json:"media_asset_id"`
	MediaType    string `json:"media_type"`
	Role         string `json:"role"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	PosterURL    string `json:"poster_url"`
	Alt          string `json:"alt"`
	Title        string `json:"title"`
	Locale       string `json:"locale"`
	SortOrder    int    `json:"sort_order"`
	IsPrimary    bool   `json:"is_primary"`
	IsVisible    *bool  `json:"is_visible"`
}

func respondProductServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	case errors.Is(err, service.ErrProductSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
	case errors.Is(err, service.ErrProductSlugInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductTranslationExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Translation already exists for this locale"})
	case errors.Is(err, service.ErrProductSpecificationTemplateInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductBrandNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product brand not found"})
	case errors.Is(err, service.ErrProductBrandInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductCategoryNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product category not found"})
	case errors.Is(err, service.ErrProductCategoryInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductLocaleImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductVariantInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSpokeRepairKitModelsInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSpokeRepairKitProductTypeImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductOptionRelationInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductMediaInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductCustomsInfoInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductFulfillmentModeInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductCustomsProfileNotFound),
		errors.Is(err, service.ErrProductCustomsProfileInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductTranslationInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductInformationTemplateInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUnsupportedLocale):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
	}
}

func normalizeSpokeRepairKitModelKeys(raw []string) []string {
	result := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func normalizeShippingTemplateIDs(raw []uint) []uint {
	result := make([]uint, 0, len(raw))
	seen := make(map[uint]struct{}, len(raw))
	for _, id := range raw {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeVariantRequests(raw []productVariantRequest) []service.ProductVariantInput {
	if len(raw) == 0 {
		return nil
	}

	variants := make([]service.ProductVariantInput, 0, len(raw))
	for _, item := range raw {
		variants = append(variants, service.ProductVariantInput{
			CarbonRimSpecification: item.CarbonRimSpecification,
			ID:                     item.ID,
			ShippingTemplateIDs:    normalizeShippingTemplateIDs(item.ShippingTemplateIDs),
			SKU:                    item.SKU,
			Title:                  item.Title,
			Currency:               item.Currency,
			PriceMinor:             item.PriceMinor,
			Stock:                  item.Stock,
			Weight:                 item.Weight,
			IsDefault:              item.IsDefault,
			IsActive:               item.IsActive,
			SortOrder:              item.SortOrder,
		})
	}
	return variants
}

func normalizeMediaRequests(raw []productMediaRequest) []service.ProductMediaInput {
	if len(raw) == 0 {
		return nil
	}

	items := make([]service.ProductMediaInput, 0, len(raw))
	for _, item := range raw {
		items = append(items, service.ProductMediaInput{
			ID:           item.ID,
			VariantID:    item.VariantID,
			MediaAssetID: item.MediaAssetID,
			MediaType:    item.MediaType,
			Role:         item.Role,
			URL:          item.URL,
			ThumbnailURL: item.ThumbnailURL,
			PosterURL:    item.PosterURL,
			Alt:          item.Alt,
			Title:        item.Title,
			Locale:       item.Locale,
			SortOrder:    item.SortOrder,
			IsPrimary:    item.IsPrimary,
			IsVisible:    item.IsVisible,
		})
	}
	return items
}
