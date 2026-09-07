package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/order"
)

var ErrOrderFulfillmentTrackingConflict = errors.New(
	"order has already been fulfilled with different tracking information",
)

// sameOrderFulfillmentTracking accepts only an exact retry of the current
// canonical fulfillment facts for this order.
func sameOrderFulfillmentTracking(
	existing *order.Order,
	resolved *resolvedOrderTrackingUpdate,
) bool {
	if existing == nil || resolved == nil || resolved.trackingInfo.TrackingProviderID == nil {
		return false
	}
	if existing.TrackingProviderID == nil ||
		*existing.TrackingProviderID != *resolved.trackingInfo.TrackingProviderID {
		return false
	}
	if strings.TrimSpace(existing.TrackingNumber) == "" ||
		strings.TrimSpace(existing.TrackingNumber) != strings.TrimSpace(resolved.trackingInfo.TrackingNumber) {
		return false
	}
	if !sameFulfillmentID(existing.CarrierID, resolved.trackingInfo.CarrierID) {
		return false
	}
	if !sameFulfillmentID(existing.CarrierServiceID, resolved.trackingInfo.CarrierServiceID) {
		return false
	}
	if !sameFulfillmentID(existing.TrackingCarrierMappingID, resolved.trackingInfo.TrackingCarrierMappingID) {
		return false
	}
	if strings.TrimSpace(existing.ProviderCarrierCode) !=
		strings.TrimSpace(resolved.trackingInfo.ProviderCarrierCode) {
		return false
	}
	return true
}

func sameFulfillmentID(stored *uint, requested *uint) bool {
	if stored == nil || requested == nil {
		return stored == nil && requested == nil
	}
	return *stored == *requested
}
