package service

import (
	"testing"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"github.com/stretchr/testify/require"
)

func TestShipmentRecordWarrantyStatus(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	record := &shippingdomain.ShipmentRecord{
		WarrantyExpires: now.Add(45*24*time.Hour + 3*time.Hour),
		Status:          "active",
	}

	status, days := ShipmentRecordWarrantyStatus(record, now)
	require.Equal(t, "valid", status)
	require.Equal(t, 45, days)

	record.WarrantyExpires = now.Add(-3*24*time.Hour - time.Hour)
	status, days = ShipmentRecordWarrantyStatus(record, now)
	require.Equal(t, "expired", status)
	require.Equal(t, 3, days)

	record.Status = "cancelled"
	status, days = ShipmentRecordWarrantyStatus(record, now)
	require.Equal(t, "expired", status)
	require.Equal(t, 3, days)
}
