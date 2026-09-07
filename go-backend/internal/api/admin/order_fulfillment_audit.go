package admin

import (
	"strings"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/service"
)

const adminAuditResourceOrderFulfillment = "order_fulfillment"

func orderFulfillmentAuditValue(record *orderdomain.Order) map[string]interface{} {
	if record == nil {
		return nil
	}
	return map[string]interface{}{
		"order_number":       strings.TrimSpace(record.OrderNumber),
		"status":             strings.TrimSpace(record.Status),
		"payment_status":     strings.TrimSpace(record.PaymentStatus),
		"shipping_status":    strings.TrimSpace(record.ShippingStatus),
		"fulfillment_mode":   orderdomain.NormalizeFulfillmentMode(record.FulfillmentMode),
		"production_status":  orderdomain.NormalizeProductionStatus(record.ProductionStatus),
		"signature_required": record.SignatureRequired,
	}
}

func orderFulfillmentAuditChanges(
	request orderFulfillmentRequest,
	result *service.OrderFulfillmentResult,
) map[string]interface{} {
	changes := map[string]interface{}{
		"tracking_number":                     strings.TrimSpace(request.TrackingNumber),
		"tracking_provider_id":                request.TrackingProviderID,
		"carrier_id":                          auditUintPointerValue(request.CarrierID),
		"carrier_service_id":                  auditUintPointerValue(request.CarrierServiceID),
		"signature_confirmed":                 request.SignatureConfirmed,
		"tracking_registration_error_present": false,
	}
	if result != nil {
		changes["tracking_registration_error_present"] =
			strings.TrimSpace(result.TrackingRegistrationError) != ""
	}
	return changes
}

func auditUintPointerValue(value *uint) interface{} {
	if value == nil {
		return nil
	}
	return *value
}
