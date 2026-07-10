package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type productCreateRequest struct {
	ProductTypeID *uint                   `json:"product_type_id"`
	Name          string                  `json:"name" binding:"required"`
	Slug          string                  `json:"slug" binding:"required"`
	Description   string                  `json:"description"`
	ShortDesc     string                  `json:"short_description"`
	Weight        int                     `json:"weight_grams" binding:"gte=0"`
	Status        string                  `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	Locale        string                  `json:"locale"`
	ParentID      *uint                   `json:"parent_id"`
	Featured      bool                    `json:"featured"`
	MetaTitle     string                  `json:"meta_title"`
	MetaDesc      string                  `json:"meta_description"`
	Specs         map[string]interface{}  `json:"specs"`
	Variants      []productVariantRequest `json:"variants"`
}

type productUpdateRequest struct {
	ProductTypeID *uint                   `json:"product_type_id"`
	Name          *string                 `json:"name" binding:"omitempty,min=1"`
	Slug          *string                 `json:"slug" binding:"omitempty,min=1"`
	Description   *string                 `json:"description"`
	ShortDesc     *string                 `json:"short_description"`
	Weight        *int                    `json:"weight_grams" binding:"omitempty,gte=0"`
	Status        *string                 `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
	Locale        *string                 `json:"locale"`
	ParentID      *uint                   `json:"parent_id"`
	Featured      *bool                   `json:"featured"`
	MetaTitle     *string                 `json:"meta_title"`
	MetaDesc      *string                 `json:"meta_description"`
	Specs         map[string]interface{}  `json:"specs"`
	Variants      []productVariantRequest `json:"variants"`
}

type productVariantRequest struct {
	ID           *uint                  `json:"id"`
	SKU          string                 `json:"sku"`
	Title        string                 `json:"title"`
	OptionValues map[string]interface{} `json:"option_values"`
	Price        float64                `json:"price"`
	SalePrice    *float64               `json:"sale_price"`
	Stock        int                    `json:"stock"`
	Weight       int                    `json:"weight_grams"`
	IsDefault    bool                   `json:"is_default"`
	IsActive     *bool                  `json:"is_active"`
	SortOrder    int                    `json:"sort_order"`
}

func respondProductServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	case errors.Is(err, service.ErrProductSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
	case errors.Is(err, service.ErrProductTypeNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product type not found"})
	case errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductVariantInvalid):
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
			ID:           item.ID,
			SKU:          item.SKU,
			Title:        item.Title,
			OptionValues: normalizeRequestSpecs(item.OptionValues),
			Price:        item.Price,
			SalePrice:    item.SalePrice,
			Stock:        item.Stock,
			Weight:       item.Weight,
			IsDefault:    item.IsDefault,
			IsActive:     item.IsActive,
			SortOrder:    item.SortOrder,
		})
	}
	return variants
}
