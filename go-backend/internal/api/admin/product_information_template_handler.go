package admin

import (
	"errors"
	"net/http"
	"strconv"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductInformationTemplateHandler struct {
	service *service.ProductInformationTemplateService
}

func NewProductInformationTemplateHandler(templateService *service.ProductInformationTemplateService) *ProductInformationTemplateHandler {
	return &ProductInformationTemplateHandler{service: templateService}
}

type productInformationTemplateRequest struct {
	Kind      string `json:"kind" binding:"required,oneof=after_sales packaging"`
	Name      string `json:"name" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	Content   string `json:"content"`
	Locale    string `json:"locale"`
	IsEnabled *bool  `json:"is_enabled"`
	SortOrder int    `json:"sort_order"`
}

func (h *ProductInformationTemplateHandler) List(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"
	items, err := h.service.List(c.Query("kind"), c.Query("locale"), includeDisabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product information templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *ProductInformationTemplateHandler) Get(c *gin.Context) {
	id, ok := parseProductInformationTemplateID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if errors.Is(err, service.ErrProductInformationTemplateNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product information template not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product information template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ProductInformationTemplateHandler) Create(c *gin.Context) {
	var req productInformationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Create(productInformationTemplateInput(req))
	if err != nil {
		respondProductInformationTemplateError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *ProductInformationTemplateHandler) Update(c *gin.Context) {
	id, ok := parseProductInformationTemplateID(c)
	if !ok {
		return
	}
	var req productInformationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Update(id, productInformationTemplateInput(req))
	if err != nil {
		respondProductInformationTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ProductInformationTemplateHandler) Delete(c *gin.Context) {
	id, ok := parseProductInformationTemplateID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondProductInformationTemplateError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product information template deleted"})
}

func productInformationTemplateInput(req productInformationTemplateRequest) service.ProductInformationTemplateInput {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	return service.ProductInformationTemplateInput{
		Kind:      req.Kind,
		Name:      req.Name,
		Slug:      req.Slug,
		Content:   req.Content,
		Locale:    req.Locale,
		IsEnabled: enabled,
		SortOrder: req.SortOrder,
	}
}

func parseProductInformationTemplateID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product information template ID"})
		return 0, false
	}
	return uint(id), true
}

func respondProductInformationTemplateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductInformationTemplateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product information template not found"})
	case errors.Is(err, service.ErrProductInformationTemplateSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product information template slug already exists"})
	case errors.Is(err, service.ErrProductInformationTemplateInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUnsupportedLocale):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product information template"})
	}
}
