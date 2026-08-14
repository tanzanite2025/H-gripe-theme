package admin

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productCreateRequest struct {
	ProductTypeID        *uint                              `json:"product_type_id"`
	BrandID              *uint                              `json:"brand_id"`
	ShippingTemplateID   *uint                              `json:"shipping_template_id"`
	AfterSalesTemplateID *uint                              `json:"after_sales_template_id"`
	PackagingTemplateID  *uint                              `json:"packaging_template_id"`
	Name                 string                             `json:"name" binding:"required"`
	Slug                 string                             `json:"slug" binding:"required"`
	Description          string                             `json:"description"`
	ShortDesc            string                             `json:"short_description"`
	Currency             string                             `json:"currency"`
	Status               string                             `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	Locale               string                             `json:"locale"`
	ParentID             *uint                              `json:"parent_id"`
	Featured             bool                               `json:"featured"`
	Specs                map[string]interface{}             `json:"specs"`
	Variants             []productVariantRequest            `json:"variants"`
	VariantOptionValues  []productVariantOptionValueRequest `json:"variant_option_values"`
	Media                []productMediaRequest              `json:"media"`
}

type productUpdateRequest struct {
	ProductTypeID        *uint                              `json:"product_type_id"`
	BrandID              *uint                              `json:"brand_id"`
	ShippingTemplateID   *uint                              `json:"shipping_template_id"`
	AfterSalesTemplateID *uint                              `json:"after_sales_template_id"`
	PackagingTemplateID  *uint                              `json:"packaging_template_id"`
	Name                 *string                            `json:"name" binding:"omitempty,min=1"`
	Slug                 *string                            `json:"slug" binding:"omitempty,min=1"`
	Description          *string                            `json:"description"`
	ShortDesc            *string                            `json:"short_description"`
	Currency             *string                            `json:"currency"`
	Status               *string                            `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
	Locale               *string                            `json:"locale"`
	ParentID             *uint                              `json:"parent_id"`
	Featured             *bool                              `json:"featured"`
	Specs                map[string]interface{}             `json:"specs"`
	Variants             []productVariantRequest            `json:"variants"`
	VariantOptionValues  []productVariantOptionValueRequest `json:"variant_option_values"`
	Media                []productMediaRequest              `json:"media"`
}

type productVariantRequest struct {
	ID                 *uint                           `json:"id"`
	ShippingTemplateID *uint                           `json:"shipping_template_id"`
	SKU                string                          `json:"sku"`
	Title              string                          `json:"title"`
	OptionValues       map[string]interface{}          `json:"option_values"`
	Currency           string                          `json:"currency"`
	Price              float64                         `json:"price"`
	SalePrice          *float64                        `json:"sale_price"`
	DisplayPrices      []currency.DisplayPriceSnapshot `json:"display_prices"`
	Stock              int                             `json:"stock"`
	Weight             int                             `json:"weight_grams"`
	IsDefault          bool                            `json:"is_default"`
	IsActive           *bool                           `json:"is_active"`
	SortOrder          int                             `json:"sort_order"`
}

type productMediaRequest struct {
	ID                   *uint  `json:"id"`
	VariantID            *uint  `json:"variant_id"`
	VariantOptionValueID *uint  `json:"variant_option_value_id"`
	MediaAssetID         *uint  `json:"media_asset_id"`
	MediaType            string `json:"media_type"`
	Role                 string `json:"role"`
	URL                  string `json:"url"`
	ThumbnailURL         string `json:"thumbnail_url"`
	PosterURL            string `json:"poster_url"`
	Alt                  string `json:"alt"`
	Title                string `json:"title"`
	Locale               string `json:"locale"`
	SortOrder            int    `json:"sort_order"`
	IsPrimary            bool   `json:"is_primary"`
	IsVisible            *bool  `json:"is_visible"`
}

type productVariantOptionValueRequest struct {
	ID                 *uint  `json:"id"`
	SpecDefinitionID   uint   `json:"spec_definition_id" binding:"required"`
	ValueKey           string `json:"value_key"`
	Label              string `json:"label"`
	ColorHex           string `json:"color_hex"`
	SwatchMediaAssetID *uint  `json:"swatch_media_asset_id"`
	SwatchURL          string `json:"swatch_url"`
	SortOrder          int    `json:"sort_order"`
	IsEnabled          *bool  `json:"is_enabled"`
}

func respondProductServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	case errors.Is(err, service.ErrProductSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
	case errors.Is(err, service.ErrProductTranslationExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Translation already exists for this locale"})
	case errors.Is(err, service.ErrProductTypeNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product type not found"})
	case errors.Is(err, service.ErrProductBrandNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product brand not found"})
	case errors.Is(err, service.ErrProductBrandInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductLocaleImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductVariantInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductMediaInvalid):
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

func normalizeRequestSpecs(raw map[string]interface{}) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	specs := make(map[string]string, len(raw))
	for key, value := range raw {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			specs[key] = typed
		case bool:
			specs[key] = strconv.FormatBool(typed)
		case float64:
			specs[key] = strconv.FormatFloat(typed, 'f', -1, 64)
		case int:
			specs[key] = strconv.Itoa(typed)
		default:
			encoded, err := json.Marshal(typed)
			if err != nil {
				continue
			}
			specs[key] = string(encoded)
		}
	}
	return specs
}

func normalizeVariantRequests(raw []productVariantRequest) []service.ProductVariantInput {
	if len(raw) == 0 {
		return nil
	}

	variants := make([]service.ProductVariantInput, 0, len(raw))
	for _, item := range raw {
		variants = append(variants, service.ProductVariantInput{
			ID:                 item.ID,
			ShippingTemplateID: item.ShippingTemplateID,
			SKU:                item.SKU,
			Title:              item.Title,
			OptionValues:       normalizeRequestSpecs(item.OptionValues),
			Currency:           item.Currency,
			Price:              item.Price,
			SalePrice:          item.SalePrice,
			DisplayPrices:      item.DisplayPrices,
			Stock:              item.Stock,
			Weight:             item.Weight,
			IsDefault:          item.IsDefault,
			IsActive:           item.IsActive,
			SortOrder:          item.SortOrder,
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
			ID:                   item.ID,
			VariantID:            item.VariantID,
			VariantOptionValueID: item.VariantOptionValueID,
			MediaAssetID:         item.MediaAssetID,
			MediaType:            item.MediaType,
			Role:                 item.Role,
			URL:                  item.URL,
			ThumbnailURL:         item.ThumbnailURL,
			PosterURL:            item.PosterURL,
			Alt:                  item.Alt,
			Title:                item.Title,
			Locale:               item.Locale,
			SortOrder:            item.SortOrder,
			IsPrimary:            item.IsPrimary,
			IsVisible:            item.IsVisible,
		})
	}
	return items
}

func normalizeVariantOptionValueRequests(raw []productVariantOptionValueRequest) []service.ProductVariantOptionValueInput {
	if len(raw) == 0 {
		return nil
	}

	items := make([]service.ProductVariantOptionValueInput, 0, len(raw))
	for _, item := range raw {
		items = append(items, service.ProductVariantOptionValueInput{
			ID:                 item.ID,
			SpecDefinitionID:   item.SpecDefinitionID,
			ValueKey:           item.ValueKey,
			Label:              item.Label,
			ColorHex:           item.ColorHex,
			SwatchMediaAssetID: item.SwatchMediaAssetID,
			SwatchURL:          item.SwatchURL,
			SortOrder:          item.SortOrder,
			IsEnabled:          item.IsEnabled,
		})
	}
	return items
}
