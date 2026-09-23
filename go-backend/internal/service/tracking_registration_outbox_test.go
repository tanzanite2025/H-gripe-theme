package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTrackingRegistrationOutboxKeyIsBoundedAndChangesWithSource(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))

	trackingNumber := strings.Repeat("T", 120)
	input := TrackingShipmentInput{
		OrderID:                  42,
		TrackingProviderID:       7,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      "carrier-a",
		TrackingCarrierMappingID: uintPtr(11),
	}
	repo := repository.NewOutboxRepository(db)
	require.NoError(t, enqueueTrackingShipmentRegistrationOutboxEvent(repo, input, time.Unix(100, 0).UTC()))

	var first outboxdomain.Event
	require.NoError(t, db.First(&first).Error)
	require.LessOrEqual(t, len(first.EventKey), 160)
	var payload outboxdomain.TrackingShipmentRegistrationPayload
	require.NoError(t, json.Unmarshal(first.Payload, &payload))
	require.Equal(t, trackingRegistrationPayloadVersion, payload.Version)
	require.Equal(t, trackingRegistrationSourceFingerprint(input), payload.SourceFingerprint)

	input.ProviderCarrierCode = "carrier-b"
	require.NoError(t, enqueueTrackingShipmentRegistrationOutboxEvent(repo, input, time.Unix(101, 0).UTC()))
	var count int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Count(&count).Error)
	require.Equal(t, int64(2), count)
}

func TestTrackingRegistrationOutboxSkipsStaleShipmentSource(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&shippingdomain.TrackingShipment{}))

	current := shippingdomain.TrackingShipment{
		OrderID:             99,
		TrackingProviderID:  3,
		TrackingNumber:      "STALE-TRACKING",
		ProviderCarrierCode: "carrier-current",
		RegistrationStatus:  "pending",
		SyncStatus:          "pending",
		Enabled:             true,
	}
	require.NoError(t, db.Create(&current).Error)
	service := NewShippingService(repository.NewShippingRepository(db))
	oldInput := TrackingShipmentInput{
		OrderID:             current.OrderID,
		TrackingProviderID:  current.TrackingProviderID,
		TrackingNumber:      current.TrackingNumber,
		ProviderCarrierCode: "carrier-old",
	}
	payload, err := json.Marshal(outboxdomain.TrackingShipmentRegistrationPayload{
		Version:             trackingRegistrationPayloadVersion,
		OrderID:             oldInput.OrderID,
		TrackingProviderID:  oldInput.TrackingProviderID,
		TrackingNumber:      oldInput.TrackingNumber,
		ProviderCarrierCode: oldInput.ProviderCarrierCode,
		SourceFingerprint:   trackingRegistrationSourceFingerprint(oldInput),
		RequestedAt:         time.Now().UTC(),
	})
	require.NoError(t, err)

	handler := NewTrackingShipmentRegistrationOutboxHandler(service)
	require.NoError(t, handler.Handle(context.Background(), outboxdomain.Event{
		EventType: outboxdomain.EventTypeTrackingShipmentRegistration,
		Payload:   payload,
	}))
	var saved shippingdomain.TrackingShipment
	require.NoError(t, db.First(&saved, current.ID).Error)
	require.Equal(t, "pending", saved.RegistrationStatus)
}
