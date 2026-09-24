package admin

import (
	"strings"

	orderdomain "commerce-platform/internal/domain/order"
)

const adminAuditResourceOrderTracking = "order_tracking"

func orderTrackingAuditValue(record *orderdomain.Order) map[string]interface{} {
	if record == nil {
		return nil
	}

	return map[string]interface{}{
		"order_number":       strings.TrimSpace(record.OrderNumber),
		"status":             strings.TrimSpace(record.Status),
		"shipping_status":    strings.TrimSpace(record.ShippingStatus),
		"shipped_at_present": record.ShippedAt != nil && !record.ShippedAt.IsZero(),
	}
}

func orderTrackingAuditChanges(request trackingInfoRequest) map[string]interface{} {
	return map[string]interface{}{
		"tracking_number":      strings.TrimSpace(request.TrackingNumber),
		"tracking_provider_id": request.TrackingProviderID,
		"carrier_id":           auditUintPointerValue(request.CarrierID),
		"carrier_service_id":   auditUintPointerValue(request.CarrierServiceID),
		"operation":            "tracking_package_upsert",
	}
}
