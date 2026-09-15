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

type ProductSupplierCostRecordHandler struct {
	service *service.ProductSupplierCostRecordService
}

type productSupplierCostRecordDetailsRequest struct {
	UnitCost                *float64 `json:"unit_cost"`
	LegacyUnitCost          *float64 `json:"purchase_price"`
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

type productSupplierCostRecordCreateRequest struct {
	SKU string `json:"sku"`
	productSupplierCostRecordDetailsRequest
}

type productSupplierCostRecordUpdateRequest struct {
	productSupplierCostRecordDetailsRequest
}

func NewProductSupplierCostRecordHandler(supplierCostRecordService *service.ProductSupplierCostRecordService) *ProductSupplierCostRecordHandler {
	return &ProductSupplierCostRecordHandler{service: supplierCostRecordService}
}

func (h *ProductSupplierCostRecordHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	records, total, err := h.service.ListAdmin(service.ProductSupplierCostRecordListInput{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product supplier cost records"})
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

func (h *ProductSupplierCostRecordHandler) ProductOptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	options, total, err := h.service.ListProductOptions(service.ProductSupplierCostRecordListInput{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
		ExactSKU: c.Query("sku"),
	})
	if err != nil {
		if errors.Is(err, service.ErrProductSupplierCostRecordInvalid) {
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

func (h *ProductSupplierCostRecordHandler) Get(c *gin.Context) {
	id, ok := parseProductSupplierCostRecordID(c)
	if !ok {
		return
	}
	record, err := h.service.GetAdmin(id)
	if err != nil {
		respondProductSupplierCostRecordError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"record": record})
}

func (h *ProductSupplierCostRecordHandler) ListByCodes(c *gin.Context) {
	records, err := h.service.ListByProductCodes(strings.Split(c.Query("codes"), ","))
	if err != nil {
		respondProductSupplierCostRecordError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (h *ProductSupplierCostRecordHandler) Create(c *gin.Context) {
	var request productSupplierCostRecordCreateRequest
	if err := bindProductSupplierCostRecordJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := h.service.Create(toProductSupplierCostRecordCreateInput(request))
	if err != nil {
		respondProductSupplierCostRecordError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"record": record})
}

func (h *ProductSupplierCostRecordHandler) Update(c *gin.Context) {
	id, ok := parseProductSupplierCostRecordID(c)
	if !ok {
		return
	}
	var request productSupplierCostRecordUpdateRequest
	if err := bindProductSupplierCostRecordJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := h.service.Update(id, toProductSupplierCostRecordUpdateInput(request))
	if err != nil {
		respondProductSupplierCostRecordError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"record": record})
}

func (h *ProductSupplierCostRecordHandler) Delete(c *gin.Context) {
	id, ok := parseProductSupplierCostRecordID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondProductSupplierCostRecordError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product supplier cost record deleted"})
}

func parseProductSupplierCostRecordID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product supplier cost record ID"})
		return 0, false
	}
	return uint(id), true
}

func bindProductSupplierCostRecordJSON(c *gin.Context, request interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(request)
}

func toProductSupplierCostRecordCreateInput(request productSupplierCostRecordCreateRequest) service.ProductSupplierCostRecordCreateInput {
	return service.ProductSupplierCostRecordCreateInput{
		SKU:                                   request.SKU,
		ProductSupplierCostRecordDetailsInput: toProductSupplierCostRecordDetailsInput(request.productSupplierCostRecordDetailsRequest),
	}
}

func toProductSupplierCostRecordUpdateInput(request productSupplierCostRecordUpdateRequest) service.ProductSupplierCostRecordUpdateInput {
	return service.ProductSupplierCostRecordUpdateInput{
		ProductSupplierCostRecordDetailsInput: toProductSupplierCostRecordDetailsInput(request.productSupplierCostRecordDetailsRequest),
	}
}

func toProductSupplierCostRecordDetailsInput(request productSupplierCostRecordDetailsRequest) service.ProductSupplierCostRecordDetailsInput {
	return service.ProductSupplierCostRecordDetailsInput{
		UnitCost:                request.supplierUnitCost(),
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

func (request productSupplierCostRecordDetailsRequest) supplierUnitCost() *float64 {
	if request.UnitCost != nil {
		return request.UnitCost
	}
	return request.LegacyUnitCost
}

func respondProductSupplierCostRecordError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductSupplierCostRecordNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "product supplier cost record not found"})
	case errors.Is(err, service.ErrProductSupplierCostRecordSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already has a supplier cost record"})
	case errors.Is(err, service.ErrProductSupplierCostRecordInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage product supplier cost record"})
	}
}
