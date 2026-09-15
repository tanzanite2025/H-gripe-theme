package admin

import (
	"errors"
	"net/http"
	"strings"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductProfitabilityHandler struct {
	service *service.ProductProfitabilityService
}

type profitabilitySupplierCostDetailsRequest struct {
	SupplierName         string `json:"supplier_name"`
	SupplierContactName  string `json:"supplier_contact_name"`
	SupplierPhone        string `json:"supplier_phone"`
	SupplierEmail        string `json:"supplier_email"`
	LeadTimeDays         int    `json:"lead_time_days"`
	MinimumOrderQuantity int    `json:"minimum_order_quantity"`
}

type profitabilityItemRequest struct {
	ProductCode string `json:"product_code" binding:"required"`
	ProductName string `json:"product_name" binding:"required"`

	SellingCurrency string `json:"currency"`
	CostCurrency    string `json:"cost_currency"`

	ListPrice           float64  `json:"list_price"`
	SalePrice           *float64 `json:"sale_price"`
	UnitCost            *float64 `json:"unit_cost"`
	UnitCostKnown       bool     `json:"unit_cost_known"`
	LegacyUnitCost      *float64 `json:"purchase_price"`
	LegacyUnitCostKnown bool     `json:"purchase_price_known"`

	InboundShippingUnitCost float64 `json:"inbound_shipping_unit_cost"`
	PackagingUnitCost       float64 `json:"packaging_unit_cost"`
	OtherUnitCost           float64 `json:"other_unit_cost"`

	SupplierCostDetails               *profitabilitySupplierCostDetailsRequest `json:"supplier_cost_details"`
	LegacySupplierCostDetailsEnvelope *profitabilitySupplierCostDetailsRequest `json:"procurement"`
}

type profitabilityItemsRequest struct {
	Items []profitabilityItemRequest `json:"items"`
}

type profitabilityBulkUpsertRequest struct {
	RequestID string                     `json:"request_id"`
	Items     []profitabilityItemRequest `json:"items"`
}

func NewProductProfitabilityHandler(profitabilityService *service.ProductProfitabilityService) *ProductProfitabilityHandler {
	return &ProductProfitabilityHandler{service: profitabilityService}
}

func (h *ProductProfitabilityHandler) Preview(c *gin.Context) {
	var request profitabilityItemsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	results, err := h.service.Preview(toProfitabilityInputs(request.Items))
	if err != nil {
		respondProductProfitabilityError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": results})
}

func (h *ProductProfitabilityHandler) ListByCodes(c *gin.Context) {
	codes := strings.Split(c.Query("codes"), ",")
	records, err := h.service.ListByCodes(codes)
	if err != nil {
		respondProductProfitabilityError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (h *ProductProfitabilityHandler) BulkUpsert(c *gin.Context) {
	var request profitabilityBulkUpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.service.BulkUpsert(toProfitabilityInputs(request.Items))
	if err != nil {
		respondProductProfitabilityError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func toProfitabilityInputs(items []profitabilityItemRequest) []service.ProfitabilityItemInput {
	inputs := make([]service.ProfitabilityItemInput, 0, len(items))
	for _, item := range items {
		input := service.ProfitabilityItemInput{
			ProductCode:             item.ProductCode,
			ProductName:             item.ProductName,
			SellingCurrency:         item.SellingCurrency,
			CostCurrency:            item.CostCurrency,
			ListPrice:               item.ListPrice,
			SalePrice:               item.SalePrice,
			UnitCost:                item.supplierUnitCost(),
			UnitCostKnown:           item.supplierUnitCostKnown(),
			InboundShippingUnitCost: item.InboundShippingUnitCost,
			PackagingUnitCost:       item.PackagingUnitCost,
			OtherUnitCost:           item.OtherUnitCost,
		}
		if supplierCostDetails := item.supplierCostDetails(); supplierCostDetails != nil {
			input.SupplierCostDetails = &service.ProfitabilitySupplierCostDetailsInput{
				SupplierName:         supplierCostDetails.SupplierName,
				SupplierContactName:  supplierCostDetails.SupplierContactName,
				SupplierPhone:        supplierCostDetails.SupplierPhone,
				SupplierEmail:        supplierCostDetails.SupplierEmail,
				LeadTimeDays:         supplierCostDetails.LeadTimeDays,
				MinimumOrderQuantity: supplierCostDetails.MinimumOrderQuantity,
			}
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func (item profitabilityItemRequest) supplierCostDetails() *profitabilitySupplierCostDetailsRequest {
	if item.SupplierCostDetails != nil {
		return item.SupplierCostDetails
	}
	return item.LegacySupplierCostDetailsEnvelope
}

func (item profitabilityItemRequest) supplierUnitCost() *float64 {
	if item.UnitCost != nil {
		return item.UnitCost
	}
	return item.LegacyUnitCost
}

func (item profitabilityItemRequest) supplierUnitCostKnown() bool {
	return item.UnitCostKnown || item.LegacyUnitCostKnown
}

func respondProductProfitabilityError(c *gin.Context, err error) {
	var batchValidationErr *service.ProfitabilityBatchValidationError
	switch {
	case errors.As(err, &batchValidationErr):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "product profitability batch is invalid",
			"items": batchValidationErr.Items,
		})
	case errors.Is(err, service.ErrProductProfitabilityInvalid),
		errors.Is(err, service.ErrProductProfitabilityBatchLarge):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductProfitabilitySupplierCostRecordRepositoryUnavailable):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "product supplier cost record repository is unavailable"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "product profitability record not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage product profitability"})
	}
}
