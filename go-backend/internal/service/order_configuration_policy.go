package service

import (
	"encoding/json"
	"fmt"

	"commerce-platform/internal/domain/order"
)

func orderConfigurationPolicies(record *order.Order) (hasNeverCancel, hasNonReturnable bool, err error) {
	if record == nil {
		return false, false, nil
	}
	for _, item := range record.Items {
		if len(item.ConfigurationSnapshotData) == 0 || string(item.ConfigurationSnapshotData) == "{}" {
			continue
		}
		var configuration ProductConfigurationSnapshot
		if err := json.Unmarshal(item.ConfigurationSnapshotData, &configuration); err != nil {
			return false, false, fmt.Errorf("parse configuration snapshot for order item %d: %w", item.ID, err)
		}
		if configuration.CancellationPolicy == ProductCustomOptionCancellationNever {
			hasNeverCancel = true
		}
		if configuration.ReturnPolicy == ProductCustomOptionReturnNotAllowed {
			hasNonReturnable = true
		}
	}
	return hasNeverCancel, hasNonReturnable, nil
}
