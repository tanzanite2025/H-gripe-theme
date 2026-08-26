package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductProcurementHandler struct {
	service *service.ProductProcurementService
}

type productProcurementDetailsRequest struct {
	PurchasePrice           *float64 `json:"purchase_price"`
	Currency                string   `json:"currency"`
	SupplierName            string   `json:"supplier_name"`
	SupplierContactName     string   `json:"supplier_contact_name"`
	SupplierPhone           string   `json:"supplier_phone"`
	SupplierEmail           string   `json:"supplier_email"`
	LeadTimeDays            int      `json:"lead_time_days"`
	MinimumOrderQuantity    int      `json:"minimum_order_quantity"`
	InboundShippingUnitCost float64  `json:"inbound_shipping_unit_cost"`
	PackagingUnitCost       float64  `json:"packaging_unit_cost"`
	OtherUnitCost           float64  `json:"other_unit_cost"`
}

type productProcurementCreateRequest struct {
	SKU string `json:"sku"`
	productProcurementDetailsRequest
}

type productProcurementUpdateRequest struct {
	productProcurementDetailsRequest
}

func NewProductProcurementHandler(procurementService *service.ProductProcurementService) *ProductProcurementHandler {
	return &ProductProcurementHandler{service: procurementService}
}

func (h *ProductProcurementHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	records, total, err := h.service.ListAdmin(service.ProductProcurementListInput{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product procurement records"})
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *ProductProcurementHandler) ProductOptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	options, total, err := h.service.ListProductOptions(service.ProductProcurementListInput{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
		ExactSKU: c.Query("sku"),
	})
	if err != nil {
		if errors.Is(err, service.ErrProductProcurementInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product options"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options": options,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *ProductProcurementHandler) Get(c *gin.Context) {
	id, ok := parseProductProcurementID(c)
	if !ok {
		return
	}
	record, err := h.service.GetAdmin(id)
	if err != nil {
		respondProductProcurementError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"record": record})
}

func (h *ProductProcurementHandler) ListByCodes(c *gin.Context) {
	records, err := h.service.ListByProductCodes(strings.Split(c.Query("codes"), ","))
	if err != nil {
		respondProductProcurementError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (h *ProductProcurementHandler) Create(c *gin.Context) {
	var request productProcurementCreateRequest
	if err := bindProductProcurementJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := h.service.Create(toProductProcurementCreateInput(request))
	if err != nil {
		respondProductProcurementError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"record": record})
}

func (h *ProductProcurementHandler) Update(c *gin.Context) {
	id, ok := parseProductProcurementID(c)
	if !ok {
		return
	}
	var request productProcurementUpdateRequest
	if err := bindProductProcurementJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := h.service.Update(id, toProductProcurementUpdateInput(request))
	if err != nil {
		respondProductProcurementError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"record": record})
}

func (h *ProductProcurementHandler) Delete(c *gin.Context) {
	id, ok := parseProductProcurementID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondProductProcurementError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product procurement record deleted"})
}

func parseProductProcurementID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product procurement record ID"})
		return 0, false
	}
	return uint(id), true
}

func bindProductProcurementJSON(c *gin.Context, request interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(request)
}

func toProductProcurementCreateInput(request productProcurementCreateRequest) service.ProductProcurementCreateInput {
	return service.ProductProcurementCreateInput{
		SKU:                            request.SKU,
		ProductProcurementDetailsInput: toProductProcurementDetailsInput(request.productProcurementDetailsRequest),
	}
}

func toProductProcurementUpdateInput(request productProcurementUpdateRequest) service.ProductProcurementUpdateInput {
	return service.ProductProcurementUpdateInput{
		ProductProcurementDetailsInput: toProductProcurementDetailsInput(request.productProcurementDetailsRequest),
	}
}

func toProductProcurementDetailsInput(request productProcurementDetailsRequest) service.ProductProcurementDetailsInput {
	return service.ProductProcurementDetailsInput{
		PurchasePrice:           request.PurchasePrice,
		Currency:                request.Currency,
		SupplierName:            request.SupplierName,
		SupplierContactName:     request.SupplierContactName,
		SupplierPhone:           request.SupplierPhone,
		SupplierEmail:           request.SupplierEmail,
		LeadTimeDays:            request.LeadTimeDays,
		MinimumOrderQuantity:    request.MinimumOrderQuantity,
		InboundShippingUnitCost: request.InboundShippingUnitCost,
		PackagingUnitCost:       request.PackagingUnitCost,
		OtherUnitCost:           request.OtherUnitCost,
	}
}

func respondProductProcurementError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductProcurementNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "product procurement record not found"})
	case errors.Is(err, service.ErrProductProcurementSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already has a procurement record"})
	case errors.Is(err, service.ErrProductProcurementInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage product procurement record"})
	}
}
