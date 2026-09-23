package admin

import (
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productCreateRequest struct {
	ProductSpecificationTemplateID *uint                               `json:"product_specification_template_id"`
	ProductCategoryID              *uint                               `json:"product_category_id"`
	BrandID                        *uint                               `json:"brand_id"`
	ShippingTemplateID             *uint                               `json:"shipping_template_id"`
	AfterSalesTemplateID           *uint                               `json:"after_sales_template_id"`
	PackagingTemplateID            *uint                               `json:"packaging_template_id"`
	CustomsClassificationProfileID *uint                               `json:"customs_classification_profile_id"`
	HSCode                         string                              `json:"hs_code"`
	CNCode                         string                              `json:"cn_code"`
	CountryOfOrigin                string                              `json:"country_of_origin"`
	CustomsDescription             string                              `json:"customs_description"`
	Name                           string                              `json:"name" binding:"required"`
	Slug                           string                              `json:"slug" binding:"required"`
	Description                    string                              `json:"description"`
	ShortDesc                      string                              `json:"short_description"`
	Currency                       string                              `json:"currency"`
	FulfillmentMode                string                              `json:"fulfillment_mode"`
	Status                         string                              `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	Locale                         string                              `json:"locale"`
	ParentID                       *uint                               `json:"parent_id"`
	Featured                       bool                                `json:"featured"`
	Specs                          map[string]interface{}              `json:"specs"`
	Variants                       []productVariantRequest             `json:"variants"`
	VariantOptionValues            []productVariantOptionValueRequest  `json:"variant_option_values"`
	Media                          []productMediaRequest               `json:"media"`
	OptionValueRelations           []productOptionValueRelationRequest `json:"option_value_relations"`
}

type productUpdateRequest struct {
	ProductSpecificationTemplateID *uint                               `json:"product_specification_template_id"`
	ProductCategoryID              *uint                               `json:"product_category_id"`
	BrandID                        *uint                               `json:"brand_id"`
	ShippingTemplateID             *uint                               `json:"shipping_template_id"`
	AfterSalesTemplateID           *uint                               `json:"after_sales_template_id"`
	PackagingTemplateID            *uint                               `json:"packaging_template_id"`
	CustomsClassificationProfileID *uint                               `json:"customs_classification_profile_id"`
	HSCode                         *string                             `json:"hs_code"`
	CNCode                         *string                             `json:"cn_code"`
	CountryOfOrigin                *string                             `json:"country_of_origin"`
	CustomsDescription             *string                             `json:"customs_description"`
	Name                           *string                             `json:"name" binding:"omitempty,min=1"`
	Slug                           *string                             `json:"slug" binding:"omitempty,min=1"`
	Description                    *string                             `json:"description"`
	ShortDesc                      *string                             `json:"short_description"`
	Currency                       *string                             `json:"currency"`
	FulfillmentMode                *string                             `json:"fulfillment_mode"`
	Status                         *string                             `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
	Locale                         *string                             `json:"locale"`
	ParentID                       *uint                               `json:"parent_id"`
	Featured                       *bool                               `json:"featured"`
	Specs                          map[string]interface{}              `json:"specs"`
	Variants                       []productVariantRequest             `json:"variants"`
	VariantOptionValues            []productVariantOptionValueRequest  `json:"variant_option_values"`
	Media                          []productMediaRequest               `json:"media"`
	OptionValueRelations           []productOptionValueRelationRequest `json:"option_value_relations"`
}

type productOptionValueRelationRequest struct {
	ID                  *uint  `json:"id"`
	SourceOptionValueID uint   `json:"source_option_value_id" binding:"required"`
	TargetOptionValueID uint   `json:"target_option_value_id" binding:"required"`
	RelationType        string `json:"relation_type" binding:"required,oneof=requires conflicts"`
}

type productVariantRequest struct {
	ID                 *uint                                  `json:"id"`
	ShippingTemplateID *uint                                  `json:"shipping_template_id"`
	SKU                string                                 `json:"sku"`
	Title              string                                 `json:"title"`
	OptionValues       map[string]interface{}                 `json:"option_values"`
	Currency           string                                 `json:"currency"`
	PriceMinor         int64                                  `json:"price_minor" binding:"required"`
	SalePriceMinor     *int64                                 `json:"sale_price_minor"`
	Stock              int                                    `json:"stock"`
	Weight             int                                    `json:"weight_grams"`
	IsDefault          bool                                   `json:"is_default"`
	IsActive           *bool                                  `json:"is_active"`
	SortOrder          int                                    `json:"sort_order"`
	OptionGroupRules   []productOptionGroupVariantRuleRequest `json:"option_group_rules"`
	OptionValueRules   []productOptionValueVariantRuleRequest `json:"option_value_rules"`
}

type productOptionGroupVariantRuleRequest struct {
	ID                    *uint `json:"id"`
	SpecDefinitionID      uint  `json:"spec_definition_id"`
	IsApplicable          bool  `json:"is_applicable"`
	MinSelectionsOverride *int  `json:"min_selections_override"`
	MaxSelectionsOverride *int  `json:"max_selections_override"`
}

type productOptionValueVariantRuleRequest struct {
	ID                          *uint  `json:"id"`
	ProductVariantOptionValueID uint   `json:"product_variant_option_value_id"`
	IsEnabled                   bool   `json:"is_enabled"`
	PriceDeltaMinorOverride     *int64 `json:"price_delta_minor_override"`
	UnavailableReason           string `json:"unavailable_reason"`
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
	ID                        *uint  `json:"id"`
	SpecDefinitionID          uint   `json:"spec_definition_id" binding:"required"`
	TemplateOptionItemID      *uint  `json:"template_option_item_id"`
	SourceTemplateRevision    int    `json:"source_template_revision"`
	ValueKey                  string `json:"value_key"`
	Label                     string `json:"label"`
	ColorHex                  string `json:"color_hex"`
	SwatchMediaAssetID        *uint  `json:"swatch_media_asset_id"`
	SwatchURL                 string `json:"swatch_url"`
	SortOrder                 int    `json:"sort_order"`
	IsEnabled                 *bool  `json:"is_enabled"`
	PriceDeltaMinor           *int64 `json:"price_delta_minor"`
	WeightDeltaGrams          int    `json:"weight_delta_grams"`
	PackagingWeightDeltaGrams int    `json:"packaging_weight_delta_grams"`
	ProductionLeadTimeDays    int    `json:"production_lead_time_days"`
	RequiresProduction        bool   `json:"requires_production"`
	CancellationPolicy        string `json:"cancellation_policy"`
	ReturnPolicy              string `json:"return_policy"`
	IsDefault                 bool   `json:"is_default"`
	InventoryPolicy           string `json:"inventory_policy"`
	ComponentVariantID        *uint  `json:"component_variant_id"`
	ComponentQuantity         int    `json:"component_quantity"`
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
	case errors.Is(err, service.ErrProductSpecificationTemplateNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product specification template not found"})
	case errors.Is(err, service.ErrProductTemplateSyncRevisionConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
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
			PriceMinor:         item.PriceMinor,
			SalePriceMinor:     item.SalePriceMinor,
			Stock:              item.Stock,
			Weight:             item.Weight,
			IsDefault:          item.IsDefault,
			IsActive:           item.IsActive,
			SortOrder:          item.SortOrder,
			OptionGroupRules:   normalizeOptionGroupVariantRules(item.OptionGroupRules),
			OptionValueRules:   normalizeOptionValueVariantRules(item.OptionValueRules),
		})
	}
	return variants
}

func normalizeOptionGroupVariantRules(raw []productOptionGroupVariantRuleRequest) []productdomain.ProductOptionGroupVariantRule {
	if len(raw) == 0 {
		return nil
	}
	result := make([]productdomain.ProductOptionGroupVariantRule, 0, len(raw))
	for _, item := range raw {
		result = append(result, productdomain.ProductOptionGroupVariantRule{ID: valueOrZero(item.ID), SpecDefinitionID: item.SpecDefinitionID, IsApplicable: item.IsApplicable, MinSelectionsOverride: item.MinSelectionsOverride, MaxSelectionsOverride: item.MaxSelectionsOverride})
	}
	return result
}

func normalizeOptionValueVariantRules(raw []productOptionValueVariantRuleRequest) []productdomain.ProductOptionValueVariantRule {
	if len(raw) == 0 {
		return nil
	}
	result := make([]productdomain.ProductOptionValueVariantRule, 0, len(raw))
	for _, item := range raw {
		result = append(result, productdomain.ProductOptionValueVariantRule{ID: valueOrZero(item.ID), ProductVariantOptionValueID: item.ProductVariantOptionValueID, IsEnabled: item.IsEnabled, PriceDeltaMinorOverride: item.PriceDeltaMinorOverride, UnavailableReason: item.UnavailableReason})
	}
	return result
}

func valueOrZero(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
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

func normalizeProductOptionValueRelationRequests(raw []productOptionValueRelationRequest) []service.ProductOptionValueRelationInput {
	if len(raw) == 0 {
		return nil
	}
	items := make([]service.ProductOptionValueRelationInput, 0, len(raw))
	for _, item := range raw {
		items = append(items, service.ProductOptionValueRelationInput{
			ID: item.ID, SourceOptionValueID: item.SourceOptionValueID, TargetOptionValueID: item.TargetOptionValueID, RelationType: item.RelationType,
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
			ID:                        item.ID,
			SpecDefinitionID:          item.SpecDefinitionID,
			TemplateOptionItemID:      item.TemplateOptionItemID,
			SourceTemplateRevision:    item.SourceTemplateRevision,
			ValueKey:                  item.ValueKey,
			Label:                     item.Label,
			ColorHex:                  item.ColorHex,
			SwatchMediaAssetID:        item.SwatchMediaAssetID,
			SwatchURL:                 item.SwatchURL,
			SortOrder:                 item.SortOrder,
			IsEnabled:                 item.IsEnabled,
			PriceDeltaMinor:           item.PriceDeltaMinor,
			WeightDeltaGrams:          item.WeightDeltaGrams,
			PackagingWeightDeltaGrams: item.PackagingWeightDeltaGrams,
			ProductionLeadTimeDays:    item.ProductionLeadTimeDays,
			RequiresProduction:        item.RequiresProduction,
			CancellationPolicy:        item.CancellationPolicy,
			ReturnPolicy:              item.ReturnPolicy,
			IsDefault:                 item.IsDefault,
			InventoryPolicy:           item.InventoryPolicy,
			ComponentVariantID:        item.ComponentVariantID,
			ComponentQuantity:         item.ComponentQuantity,
		})
	}
	return items
}
