package admin

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type productTypeRequest struct {
	Name            string                         `json:"name" binding:"required"`
	Slug            string                         `json:"slug" binding:"required"`
	Description     string                         `json:"description"`
	SortOrder       int                            `json:"sort_order"`
	IsEnabled       *bool                          `json:"is_enabled" binding:"required"`
	SpecDefinitions []productSpecDefinitionRequest `json:"spec_definitions"`
}

type productSpecDefinitionRequest struct {
	ID              uint   `json:"id"`
	Group           string `json:"group"`
	Name            string `json:"name" binding:"required"`
	Slug            string `json:"slug" binding:"required"`
	FieldType       string `json:"field_type" binding:"required,oneof=text number select boolean"`
	Unit            string `json:"unit"`
	IsRequired      bool   `json:"is_required"`
	IsFilterable    bool   `json:"is_filterable"`
	IsVisible       bool   `json:"is_visible"`
	IsVariantOption bool   `json:"is_variant_option"`
	SortOrder       int    `json:"sort_order"`
	Options         string `json:"options"`
	Validation      string `json:"validation"`
}

func (h *ProductHandler) GetProductType(c *gin.Context) {
	id, ok := parseProductTypeID(c)
	if !ok {
		return
	}
	productType, err := h.productService.GetProductType(id)
	if err != nil {
		respondProductTypeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productType})
}

func (h *ProductHandler) CreateProductType(c *gin.Context) {
	var request productTypeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productType, err := h.productService.CreateProductType(productTypeInputFromRequest(request))
	if err != nil {
		respondProductTypeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": productType})
}

func (h *ProductHandler) UpdateProductType(c *gin.Context) {
	id, ok := parseProductTypeID(c)
	if !ok {
		return
	}
	var request productTypeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productType, err := h.productService.UpdateProductType(id, productTypeInputFromRequest(request))
	if err != nil {
		respondProductTypeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productType})
}

func (h *ProductHandler) DeleteProductType(c *gin.Context) {
	id, ok := parseProductTypeID(c)
	if !ok {
		return
	}
	if err := h.productService.DeleteProductType(id); err != nil {
		respondProductTypeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product type deleted"})
}

func parseProductTypeID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product type id"})
		return 0, false
	}
	return uint(id), true
}

func productTypeInputFromRequest(request productTypeRequest) service.ProductTypeInput {
	definitions := make([]service.ProductSpecDefinitionInput, 0, len(request.SpecDefinitions))
	for _, definition := range request.SpecDefinitions {
		definitions = append(definitions, service.ProductSpecDefinitionInput{
			ID:              definition.ID,
			Group:           definition.Group,
			Name:            definition.Name,
			Slug:            definition.Slug,
			FieldType:       definition.FieldType,
			Unit:            definition.Unit,
			IsRequired:      definition.IsRequired,
			IsFilterable:    definition.IsFilterable,
			IsVisible:       definition.IsVisible,
			IsVariantOption: definition.IsVariantOption,
			SortOrder:       definition.SortOrder,
			Options:         definition.Options,
			Validation:      definition.Validation,
		})
	}
	return service.ProductTypeInput{
		Name:            request.Name,
		Slug:            request.Slug,
		Description:     request.Description,
		SortOrder:       request.SortOrder,
		IsEnabled:       request.IsEnabled != nil && *request.IsEnabled,
		SpecDefinitions: definitions,
	}
}

func respondProductTypeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductTypeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product type not found"})
	case errors.Is(err, service.ErrProductTypeSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product type slug already exists"})
	case errors.Is(err, service.ErrProductTypeInvalid), errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to manage product type"})
	}
}
