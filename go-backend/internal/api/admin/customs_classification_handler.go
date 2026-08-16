package admin

import (
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomsClassificationHandler struct {
	service *service.CustomsClassificationService
}

type customsClassificationRequest struct {
	ProductSpecificationTemplateID *uint  `json:"product_specification_template_id"`
	Name                           string `json:"name" binding:"required"`
	Slug                           string `json:"slug" binding:"required"`
	ComponentKind                  string `json:"component_kind"`
	Material                       string `json:"material"`
	HSCode                         string `json:"hs_code" binding:"required"`
	CNCode                         string `json:"cn_code"`
	CountryOfOrigin                string `json:"country_of_origin"`
	CustomsDescription             string `json:"customs_description"`
	Source                         string `json:"source"`
	SourceCode                     string `json:"source_code"`
	SourceURL                      string `json:"source_url"`
	Notes                          string `json:"notes"`
	Status                         string `json:"status"`
}

func NewCustomsClassificationHandler(customsService *service.CustomsClassificationService) *CustomsClassificationHandler {
	return &CustomsClassificationHandler{service: customsService}
}

func (h *CustomsClassificationHandler) List(c *gin.Context) {
	productSpecificationTemplateID, _ := strconv.ParseUint(c.Query("product_specification_template_id"), 10, 32)
	items, err := h.service.List(service.CustomsClassificationListInput{
		ProductSpecificationTemplateID: uint(productSpecificationTemplateID),
		ComponentKind:                  c.Query("component_kind"),
		Material:                       c.Query("material"),
		Status:                         c.Query("status"),
		Search:                         c.Query("q"),
		IncludePaused:                  c.Query("include_paused") == "true" || c.Query("include_paused") == "1",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customs classifications"})
		return
	}
	response.Success(c, items)
}

func (h *CustomsClassificationHandler) Get(c *gin.Context) {
	id, ok := parseCustomsClassificationID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		respondCustomsClassificationError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *CustomsClassificationHandler) Create(c *gin.Context) {
	var req customsClassificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Create(customsClassificationInput(req))
	if err != nil {
		respondCustomsClassificationError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *CustomsClassificationHandler) Update(c *gin.Context) {
	id, ok := parseCustomsClassificationID(c)
	if !ok {
		return
	}
	var req customsClassificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Update(id, customsClassificationInput(req))
	if err != nil {
		respondCustomsClassificationError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *CustomsClassificationHandler) Delete(c *gin.Context) {
	id, ok := parseCustomsClassificationID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondCustomsClassificationError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Customs classification deleted"})
}

func (h *CustomsClassificationHandler) Lookup(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, err := h.service.Lookup(service.CustomsClassificationLookupInput{
		Provider: c.DefaultQuery("provider", service.CustomsLookupProviderUSHTS),
		Query:    c.Query("q"),
		Limit:    limit,
	})
	if err != nil {
		respondCustomsClassificationError(c, err)
		return
	}
	response.Success(c, items)
}

func customsClassificationInput(req customsClassificationRequest) service.CustomsClassificationInput {
	return service.CustomsClassificationInput{
		ProductSpecificationTemplateID: req.ProductSpecificationTemplateID,
		Name:                           req.Name,
		Slug:                           req.Slug,
		ComponentKind:                  req.ComponentKind,
		Material:                       req.Material,
		HSCode:                         req.HSCode,
		CNCode:                         req.CNCode,
		CountryOfOrigin:                req.CountryOfOrigin,
		CustomsDescription:             req.CustomsDescription,
		Source:                         req.Source,
		SourceCode:                     req.SourceCode,
		SourceURL:                      req.SourceURL,
		Notes:                          req.Notes,
		Status:                         req.Status,
	}
}

func parseCustomsClassificationID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customs classification ID"})
		return 0, false
	}
	return uint(id), true
}

func respondCustomsClassificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCustomsClassificationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Customs classification not found"})
	case errors.Is(err, service.ErrCustomsClassificationSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Customs classification slug already exists"})
	case errors.Is(err, service.ErrCustomsClassificationInvalid),
		errors.Is(err, service.ErrCustomsLookupInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrCustomsLookupUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Customs classification operation failed"})
	}
}
