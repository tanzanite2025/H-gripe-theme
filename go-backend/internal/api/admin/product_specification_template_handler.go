package admin

import (
	"commerce-platform/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productSpecificationTemplateRequest struct {
	Name            string                         `json:"name" binding:"required"`
	Slug            string                         `json:"slug" binding:"required"`
	Description     string                         `json:"description"`
	SortOrder       int                            `json:"sort_order"`
	IsEnabled       *bool                          `json:"is_enabled" binding:"required"`
	SpecDefinitions []productSpecDefinitionRequest `json:"spec_definitions"`
}

type productSpecDefinitionRequest struct {
	ID            uint                           `json:"id"`
	Group         string                         `json:"group"`
	Name          string                         `json:"name" binding:"required"`
	Slug          string                         `json:"slug" binding:"required"`
	FieldType     string                         `json:"field_type" binding:"required,oneof=text number select boolean"`
	Role          string                         `json:"role" binding:"omitempty,oneof=attribute variant custom_option"`
	SelectionMode string                         `json:"selection_mode" binding:"omitempty,oneof=single multiple"`
	MinSelections int                            `json:"min_selections" binding:"min=0"`
	MaxSelections *int                           `json:"max_selections" binding:"omitempty,min=0"`
	Presentation  string                         `json:"presentation"`
	Unit          string                         `json:"unit"`
	IsRequired    bool                           `json:"is_required"`
	IsFilterable  bool                           `json:"is_filterable"`
	IsVisible     bool                           `json:"is_visible"`
	SortOrder     int                            `json:"sort_order"`
	Validation    string                         `json:"validation"`
	OptionItems   []productSpecOptionItemRequest `json:"option_items"`
}

type productSpecOptionItemRequest struct {
	ID                     uint   `json:"id"`
	ValueKey               string `json:"value_key" binding:"required"`
	DefaultLabel           string `json:"default_label"`
	ColorHex               string `json:"color_hex"`
	SwatchMediaAssetID     *uint  `json:"swatch_media_asset_id"`
	SwatchURL              string `json:"swatch_url"`
	IsEnabledByDefault     *bool  `json:"is_enabled_by_default"`
	IsDefault              bool   `json:"is_default"`
	DefaultPriceDeltaMinor *int64 `json:"default_price_delta_minor"`
	DefaultPriceCurrency   string `json:"default_price_currency"`
	SortOrder              int    `json:"sort_order"`
	Revision               int    `json:"revision"`
}

func (h *ProductHandler) GetProductSpecificationTemplate(c *gin.Context) {
	id, ok := parseProductSpecificationTemplateID(c)
	if !ok {
		return
	}
	productSpecificationTemplate, err := h.productService.GetProductSpecificationTemplate(id)
	if err != nil {
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productSpecificationTemplate})
}

func (h *ProductHandler) CreateProductSpecificationTemplate(c *gin.Context) {
	var request productSpecificationTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productSpecificationTemplate, err := h.productService.CreateProductSpecificationTemplate(productSpecificationTemplateInputFromRequest(request))
	if err != nil {
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": productSpecificationTemplate})
}

func (h *ProductHandler) UpdateProductSpecificationTemplate(c *gin.Context) {
	id, ok := parseProductSpecificationTemplateID(c)
	if !ok {
		return
	}
	var request productSpecificationTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productSpecificationTemplate, err := h.productService.UpdateProductSpecificationTemplate(id, productSpecificationTemplateInputFromRequest(request))
	if err != nil {
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productSpecificationTemplate})
}

func (h *ProductHandler) DeleteProductSpecificationTemplate(c *gin.Context) {
	id, ok := parseProductSpecificationTemplateID(c)
	if !ok {
		return
	}
	if err := h.productService.DeleteProductSpecificationTemplate(id); err != nil {
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product specification template deleted"})
}

func parseProductSpecificationTemplateID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product specification template id"})
		return 0, false
	}
	return uint(id), true
}

func productSpecificationTemplateInputFromRequest(request productSpecificationTemplateRequest) service.ProductSpecificationTemplateInput {
	definitions := make([]service.ProductSpecDefinitionInput, 0, len(request.SpecDefinitions))
	for _, definition := range request.SpecDefinitions {
		definitions = append(definitions, service.ProductSpecDefinitionInput{
			ID:            definition.ID,
			Group:         definition.Group,
			Name:          definition.Name,
			Slug:          definition.Slug,
			FieldType:     definition.FieldType,
			Role:          definition.Role,
			SelectionMode: definition.SelectionMode,
			MinSelections: definition.MinSelections,
			MaxSelections: definition.MaxSelections,
			Presentation:  definition.Presentation,
			Unit:          definition.Unit,
			IsRequired:    definition.IsRequired,
			IsFilterable:  definition.IsFilterable,
			IsVisible:     definition.IsVisible,
			SortOrder:     definition.SortOrder,
			Validation:    definition.Validation,
			OptionItems:   productSpecOptionItemsFromRequest(definition.OptionItems),
		})
	}
	input := service.ProductSpecificationTemplateInput{
		Name:            request.Name,
		Slug:            request.Slug,
		Description:     request.Description,
		SortOrder:       request.SortOrder,
		IsEnabled:       request.IsEnabled != nil && *request.IsEnabled,
		SpecDefinitions: definitions,
	}
	return input
}

func productSpecOptionItemsFromRequest(items []productSpecOptionItemRequest) []service.ProductSpecOptionItemInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]service.ProductSpecOptionItemInput, 0, len(items))
	for _, item := range items {
		result = append(result, service.ProductSpecOptionItemInput{
			ID:                     item.ID,
			ValueKey:               item.ValueKey,
			DefaultLabel:           item.DefaultLabel,
			ColorHex:               item.ColorHex,
			SwatchMediaAssetID:     item.SwatchMediaAssetID,
			SwatchURL:              item.SwatchURL,
			IsEnabledByDefault:     item.IsEnabledByDefault,
			IsDefault:              item.IsDefault,
			DefaultPriceDeltaMinor: item.DefaultPriceDeltaMinor,
			DefaultPriceCurrency:   item.DefaultPriceCurrency,
			SortOrder:              item.SortOrder,
			Revision:               item.Revision,
		})
	}
	return result
}

func respondProductSpecificationTemplateServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductSpecificationTemplateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product specification template not found"})
	case errors.Is(err, service.ErrProductSpecificationTemplateSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product specification template slug already exists"})
	case errors.Is(err, service.ErrProductSpecificationTemplateSystemManaged):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductSpecificationTemplateInvalid), errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUnsupportedLocale):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to manage product specification template"})
	}
}
