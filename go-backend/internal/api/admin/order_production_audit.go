package admin

import orderdomain "commerce-platform/internal/domain/order"

const adminAuditResourceOrderProduction = "order_production"

func orderProductionAuditValue(record *orderdomain.Order) map[string]interface{} {
	if record == nil {
		return nil
	}
	return map[string]interface{}{
		"order_number":      record.OrderNumber,
		"status":            record.Status,
		"payment_status":    record.PaymentStatus,
		"fulfillment_mode":  orderdomain.NormalizeFulfillmentMode(record.FulfillmentMode),
		"production_status": orderdomain.NormalizeProductionStatus(record.ProductionStatus),
	}
}
