package warranty

import (
	"encoding/json"
	"strings"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetWarrantyStatus(c *gin.Context) {
	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		apierror.RespondBadRequest(c, "Order number is required")
		return
	}

	userID, ok := warrantyUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}
	if h.shipmentSvc == nil {
		apierror.RespondInternalError(c, service.ErrShipmentRecordUnavailable)
		return
	}

	shipment, err := h.shipmentSvc.GetForUser(orderNumber, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"success": true,
		"data":    shipmentWarrantyStatusResponse(shipment),
	})
}

func warrantyUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch typed := value.(type) {
	case uint:
		return typed, typed > 0
	case uint64:
		return uint(typed), typed > 0
	case int:
		return uint(typed), typed > 0
	default:
		return 0, false
	}
}

func shipmentWarrantyStatusResponse(record *shippingdomain.ShipmentRecord) gin.H {
	status, remainingDays := service.ShipmentRecordWarrantyStatus(record, time.Now())
	items := make([]gin.H, 0)
	if record != nil && len(record.ItemsSnapshot) > 0 {
		var rawItems []map[string]interface{}
		if err := json.Unmarshal(record.ItemsSnapshot, &rawItems); err == nil {
			for _, item := range rawItems {
				items = append(items, gin.H{
					"product_name": item["product_name"],
					"sku":          item["sku"],
					"quantity":     item["quantity"],
					"variant_id":   item["variant_id"],
				})
			}
		}
	}

	remaining := gin.H{"months": 0, "days": 0, "total_days": 0}
	if status == "valid" {
		remaining["months"] = remainingDays / 30
		remaining["days"] = remainingDays % 30
		remaining["total_days"] = remainingDays
	} else {
		remaining["expired_days"] = remainingDays
	}

	firstProductName := ""
	if len(items) > 0 {
		if name, ok := items[0]["product_name"].(string); ok {
			firstProductName = name
		}
	}

	return gin.H{
		"order_number":  record.OrderNumber,
		"product_type":  gin.H{"code": "order-shipment", "name": "Order shipment", "name_zh": "订单发货"},
		"product_name":  firstProductName,
		"items":         items,
		"ship_date":     record.ShippedAt.Format("2006-01-02"),
		"warranty_months": record.WarrantyMonths,
		"warranty_end":  record.WarrantyExpires.Format("2006-01-02"),
		"status":        status,
		"remaining":     remaining,
		"records":       []gin.H{},
	}
}
