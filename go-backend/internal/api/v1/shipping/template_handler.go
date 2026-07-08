package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.shippingService.ListTemplates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": templates})
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid template id")
		return
	}

	template, err := h.shippingService.GetTemplate(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Template")
		return
	}

	response.Success(c, template)
}

func (h *Handler) CalculateShipping(c *gin.Context) {
	var req struct {
		TemplateID uint    `json:"template_id" binding:"required"`
		Weight     float64 `json:"weight"`
		Quantity   int     `json:"quantity"`
		Amount     float64 `json:"amount"`
		Country    string  `json:"country" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	quote, err := h.shippingService.CalculateShipping(service.ShippingCalculationInput{
		TemplateID: req.TemplateID,
		Weight:     req.Weight,
		Quantity:   req.Quantity,
		Amount:     req.Amount,
		Country:    req.Country,
	})
	if err != nil {
		apierror.RespondNotFound(c, "Template")
		return
	}

	response.Success(c, quote)
}
