package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductBrandHandler struct {
	service *service.ProductBrandService
}

type productBrandRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	WebsiteURL  string `json:"website_url"`
	IsEnabled   *bool  `json:"is_enabled"`
	SortOrder   int    `json:"sort_order"`
}

func NewProductBrandHandler(brandService *service.ProductBrandService) *ProductBrandHandler {
	return &ProductBrandHandler{service: brandService}
}

func (h *ProductBrandHandler) List(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"
	brands, err := h.service.List(includeDisabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product brands"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": brands})
}

func (h *ProductBrandHandler) Get(c *gin.Context) {
	id, ok := parseProductBrandID(c)
	if !ok {
		return
	}
	brand, err := h.service.Get(id)
	if err != nil {
		respondProductBrandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": brand})
}

func (h *ProductBrandHandler) Create(c *gin.Context) {
	var req productBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	brand, err := h.service.Create(productBrandInput(req))
	if err != nil {
		respondProductBrandError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": brand})
}

func (h *ProductBrandHandler) Update(c *gin.Context) {
	id, ok := parseProductBrandID(c)
	if !ok {
		return
	}
	var req productBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	brand, err := h.service.Update(id, productBrandInput(req))
	if err != nil {
		respondProductBrandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": brand})
}

func (h *ProductBrandHandler) Delete(c *gin.Context) {
	id, ok := parseProductBrandID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondProductBrandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product brand deleted"})
}

func productBrandInput(req productBrandRequest) service.ProductBrandInput {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	return service.ProductBrandInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		WebsiteURL:  req.WebsiteURL,
		IsEnabled:   enabled,
		SortOrder:   req.SortOrder,
	}
}

func parseProductBrandID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product brand ID"})
		return 0, false
	}
	return uint(id), true
}

func respondProductBrandError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductBrandNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product brand not found"})
	case errors.Is(err, service.ErrProductBrandSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Product brand slug already exists"})
	case errors.Is(err, service.ErrProductBrandInUse):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductBrandInSpokeRimCatalog):
		c.JSON(http.StatusConflict, gin.H{"error": "Product brand is referenced by the spoke rim catalog and cannot be deleted"})
	case errors.Is(err, service.ErrProductBrandInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to manage product brand"})
	}
}
