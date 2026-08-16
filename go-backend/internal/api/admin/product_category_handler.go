package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductCategoryHandler struct {
	service *service.ProductCategoryService
}

type productCategoryRequest struct {
	ParentID          *uint  `json:"parent_id"`
	Name              string `json:"name" binding:"required"`
	Slug              string `json:"slug" binding:"required"`
	Description       string `json:"description"`
	ImageMediaAssetID *uint  `json:"image_media_asset_id"`
	IsEnabled         *bool  `json:"is_enabled"`
	SortOrder         int    `json:"sort_order"`
}

type productCategoryTranslationsRequest struct {
	Translations []productCategoryTranslationRequest `json:"translations"`
}

type productCategoryTranslationRequest struct {
	ID          uint   `json:"id"`
	Locale      string `json:"locale"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewProductCategoryHandler(categoryService *service.ProductCategoryService) *ProductCategoryHandler {
	return &ProductCategoryHandler{service: categoryService}
}

func (h *ProductCategoryHandler) List(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"
	categories, err := h.service.List(includeDisabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *ProductCategoryHandler) Get(c *gin.Context) {
	id, ok := parseProductCategoryID(c)
	if !ok {
		return
	}
	category, err := h.service.Get(id)
	if err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func (h *ProductCategoryHandler) ListTranslations(c *gin.Context) {
	id, ok := parseProductCategoryID(c)
	if !ok {
		return
	}
	translations, err := h.service.ListTranslations(id)
	if err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": translations})
}

func (h *ProductCategoryHandler) UpdateTranslations(c *gin.Context) {
	id, ok := parseProductCategoryID(c)
	if !ok {
		return
	}

	var req productCategoryTranslationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inputs := make([]service.ProductCategoryTranslationInput, 0, len(req.Translations))
	for _, translation := range req.Translations {
		inputs = append(inputs, service.ProductCategoryTranslationInput{
			ID:          translation.ID,
			Locale:      translation.Locale,
			Name:        translation.Name,
			Description: translation.Description,
		})
	}
	translations, err := h.service.UpdateTranslations(id, inputs)
	if err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": translations})
}

func (h *ProductCategoryHandler) Create(c *gin.Context) {
	var req productCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := h.service.Create(productCategoryInput(req))
	if err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": category})
}

func (h *ProductCategoryHandler) Update(c *gin.Context) {
	id, ok := parseProductCategoryID(c)
	if !ok {
		return
	}
	var req productCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := h.service.Update(id, productCategoryInput(req))
	if err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func (h *ProductCategoryHandler) Delete(c *gin.Context) {
	id, ok := parseProductCategoryID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondProductCategoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product category deleted"})
}

func productCategoryInput(req productCategoryRequest) service.ProductCategoryInput {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	return service.ProductCategoryInput{
		ParentID:          req.ParentID,
		Name:              req.Name,
		Slug:              req.Slug,
		Description:       req.Description,
		ImageMediaAssetID: req.ImageMediaAssetID,
		IsEnabled:         enabled,
		SortOrder:         req.SortOrder,
	}
}

func parseProductCategoryID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product category ID"})
		return 0, false
	}
	return uint(id), true
}

func respondProductCategoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductCategoryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
	case errors.Is(err, service.ErrProductCategorySlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product category slug already exists"})
	case errors.Is(err, service.ErrProductCategoryHasChildren):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductCategoryImageInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product category image must be an active public image from the media library"})
	case errors.Is(err, service.ErrProductCategoryInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductCategoryTranslationInvalid), errors.Is(err, service.ErrUnsupportedLocale):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to manage product category"})
	}
}
