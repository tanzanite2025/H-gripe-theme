package service

import (
	"context"
	"testing"
	"time"

	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestResolveTrackingCarrierUsesCarrierServiceMappingFirst(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)

	seedTrackingCarrierMapping(t, db, provider.ID, "carrier", &carrier.ID, nil, "DHL")
	serviceMapping := seedTrackingCarrierMapping(t, db, provider.ID, "carrier_service", nil, &carrierService.ID, "DHL-EXP-US")

	resolution, err := shippingService.ResolveTrackingCarrier(TrackingCarrierResolutionInput{
		ProviderID:       provider.ID,
		CarrierServiceID: &carrierService.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "DHL-EXP-US", resolution.ProviderCarrierCode)
	assert.Equal(t, serviceMapping.ID, resolution.Mapping.ID)
	assert.Equal(t, carrier.ID, resolution.Carrier.ID)
	assert.Equal(t, carrierService.ID, resolution.CarrierService.ID)
}

func TestResolveTrackingCarrierFallsBackToCarrierMapping(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)

	carrierMapping := seedTrackingCarrierMapping(t, db, provider.ID, "carrier", &carrier.ID, nil, "DHL")

	resolution, err := shippingService.ResolveTrackingCarrier(TrackingCarrierResolutionInput{
		ProviderID:       provider.ID,
		CarrierServiceID: &carrierService.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "DHL", resolution.ProviderCarrierCode)
	assert.Equal(t, carrierMapping.ID, resolution.Mapping.ID)
	assert.Equal(t, carrier.ID, resolution.Carrier.ID)
	assert.Equal(t, carrierService.ID, resolution.CarrierService.ID)
}

func TestSyncTrackingPersistsEventsFromProvider(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)

	result, err := shippingService.SyncTracking(context.Background(), TrackingSyncInput{
		OrderID:             88,
		ProviderID:          provider.ID,
		TrackingNumber:      "MOCK123456",
		ProviderCarrierCode: "DHL",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Events, 2)
	assert.Equal(t, "MOCK123456", result.TrackingNumber)
	assert.Equal(t, "DHL", result.Carrier)
	require.NotNil(t, result.Shipment)
	assert.Equal(t, "synced", result.Shipment.SyncStatus)
	assert.Equal(t, 2, result.Shipment.EventCount)
	assert.NotNil(t, result.Shipment.LastEventAt)
	assert.NotNil(t, result.Shipment.LastSyncedAt)

	events, err := shippingService.GetTrackingEventsByOrderID(88)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "MOCK123456", events[0].TrackingNumber)
	assert.Equal(t, "DHL", events[0].ProviderCarrierCode)
}

func TestSyncTrackingAutoRegistersProviderBeforeSync(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode:   "mock",
		ProviderName:   "Mock Provider",
		Enabled:        true,
		AutoRegister:   true,
		PollingEnabled: true,
	}
	require.NoError(t, db.Create(&provider).Error)

	result, err := shippingService.SyncTracking(context.Background(), TrackingSyncInput{
		OrderID:             89,
		ProviderID:          provider.ID,
		TrackingNumber:      "MOCK123489",
		ProviderCarrierCode: "DHL",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shipment)
	assert.Equal(t, "registered", result.Shipment.RegistrationStatus)
	assert.Equal(t, "synced", result.Shipment.SyncStatus)
	assert.NotNil(t, result.Shipment.NextSyncAt)
}

func TestUpsertTrackingShipmentPreservesOperationalStateWhenSourceUnchanged(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)

	eventTime := time.Now().Add(-time.Hour)
	syncedAt := time.Now().Add(-time.Minute)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             90,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-UP-SAME",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          1,
		LastEventAt:         &eventTime,
		LastSyncedAt:        &syncedAt,
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             90,
		TrackingNumber:      "TRACK-UP-SAME",
		ProviderCarrierCode: "DHL",
		Status:              "In transit",
		Description:         "Already synced",
		EventTime:           eventTime,
	}).Error)

	shipment, err := shippingService.UpsertTrackingShipment(TrackingShipmentInput{
		OrderID:             90,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-UP-SAME",
		ProviderCarrierCode: "DHL",
	})

	require.NoError(t, err)
	require.NotNil(t, shipment)
	assert.Equal(t, "registered", shipment.RegistrationStatus)
	assert.Equal(t, "synced", shipment.SyncStatus)
	assert.Equal(t, 1, shipment.EventCount)
	assert.NotNil(t, shipment.LastSyncedAt)

	events, err := shippingService.GetTrackingEventsByOrderID(90)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "Already synced", events[0].Description)
}

func TestUpsertTrackingShipmentResetsOperationalStateWhenSourceChanges(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)

	eventTime := time.Now().Add(-time.Hour)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             91,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-UP-OLD",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          1,
		LastEventAt:         &eventTime,
		LastSyncedAt:        &eventTime,
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             91,
		TrackingNumber:      "TRACK-UP-OLD",
		ProviderCarrierCode: "DHL",
		Status:              "Delivered",
		Description:         "Old tracking event",
		EventTime:           eventTime,
	}).Error)

	shipment, err := shippingService.UpsertTrackingShipment(TrackingShipmentInput{
		OrderID:             91,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-UP-NEW",
		ProviderCarrierCode: "DHL",
	})

	require.NoError(t, err)
	require.NotNil(t, shipment)
	assert.Equal(t, "TRACK-UP-NEW", shipment.TrackingNumber)
	assert.Equal(t, "pending", shipment.RegistrationStatus)
	assert.Equal(t, "pending", shipment.SyncStatus)
	assert.Equal(t, 0, shipment.EventCount)
	assert.Nil(t, shipment.LastSyncedAt)

	events, err := shippingService.GetTrackingEventsByOrderID(91)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSyncDueTrackingShipmentsProcessesPendingShipments(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             99,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "MOCK123499",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "pending",
		SyncStatus:          "pending",
		Enabled:             true,
	}).Error)

	result, err := shippingService.SyncDueTrackingShipments(context.Background(), 10)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Matched)
	assert.Equal(t, 1, result.Synced)
	assert.Equal(t, 0, result.Failed)

	shipment, err := shippingService.GetTrackingShipmentByOrderID(99)
	require.NoError(t, err)
	assert.Equal(t, "synced", shipment.SyncStatus)
	assert.Equal(t, 2, shipment.EventCount)
}

func TestListTrackingShipmentsAppliesOperationalFilters(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider, carrier, carrierService := seedTrackingProviderCarrierAndService(t, db)
	otherProvider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Provider",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&otherProvider).Error)

	past := time.Now().Add(-time.Minute)
	future := time.Now().Add(time.Hour)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             201,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-FILTER-201",
		ProviderCarrierCode: "190271",
		CarrierID:           &carrier.ID,
		CarrierServiceID:    &carrierService.ID,
		RegistrationStatus:  "failed",
		SyncStatus:          "failed",
		NextSyncAt:          &past,
		LastError:           "17TRACK api key missing",
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             202,
		TrackingProviderID:  otherProvider.ID,
		TrackingNumber:      "TRACK-FILTER-202",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		NextSyncAt:          &future,
		Enabled:             true,
	}).Error)

	enabled := true
	shipments, err := shippingService.ListTrackingShipments(TrackingShipmentListFilter{
		SyncStatus:         "failed",
		RegistrationStatus: "failed",
		ProviderID:         provider.ID,
		CarrierID:          carrier.ID,
		CarrierServiceID:   carrierService.ID,
		Enabled:            &enabled,
		DueOnly:            true,
		Keyword:            "api key",
		Limit:              10,
	})

	require.NoError(t, err)
	require.Len(t, shipments, 1)
	assert.Equal(t, uint(201), shipments[0].OrderID)
	assert.Equal(t, "TRACK-FILTER-201", shipments[0].TrackingNumber)
}

func TestTrackingPollingStateCapturesLatestRun(t *testing.T) {
	_, shippingService := newTestShippingTrackingService(t)
	startedAt := time.Now().Add(-2 * time.Second)

	shippingService.ConfigureTrackingPolling(true, 5*time.Minute, 20)
	shippingService.MarkTrackingPollingStarted(startedAt)
	shippingService.MarkTrackingPollingFinished(startedAt, &TrackingShipmentSyncBatchResult{
		Matched: 3,
		Synced:  2,
		Failed:  1,
		Errors: []TrackingShipmentSyncFailure{
			{OrderID: 301, TrackingNumber: "TRACK-301", Error: "provider timeout"},
		},
	}, nil)

	state := shippingService.TrackingPollingState()
	assert.True(t, state.Enabled)
	assert.False(t, state.Running)
	assert.Equal(t, int((5 * time.Minute).Seconds()), state.IntervalSeconds)
	assert.Equal(t, 20, state.BatchLimit)
	assert.Equal(t, 3, state.LastMatched)
	assert.Equal(t, 2, state.LastSynced)
	assert.Equal(t, 1, state.LastFailed)
	assert.NotNil(t, state.LastStartedAt)
	assert.NotNil(t, state.LastFinishedAt)
	require.Len(t, state.LastErrors, 1)
	assert.Equal(t, "provider timeout", state.LastErrors[0].Error)
}

func TestTrackingWebhookStateCapturesLatestWebhookWithoutPayload(t *testing.T) {
	_, shippingService := newTestShippingTrackingService(t)
	receivedAt := time.Now().Add(-time.Second)
	finishedAt := time.Now()

	shippingService.RecordTrackingWebhookRun(TrackingWebhookRunState{
		LastReceivedAt:       &receivedAt,
		LastFinishedAt:       &finishedAt,
		LastDurationMs:       12,
		LastProviderCode:     "17TRACK",
		LastProviderID:       7,
		LastTrackingNumber:   "TRACK-WEBHOOK-STATE",
		LastCarrierCode:      "190271",
		LastOrderID:          401,
		LastEventCount:       3,
		LastHTTPStatus:       200,
		LastAccepted:         true,
		LastSignatureChecked: true,
		LastSignatureValid:   true,
	})

	state := shippingService.TrackingWebhookState()
	require.NotNil(t, state.LastReceivedAt)
	require.NotNil(t, state.LastFinishedAt)
	assert.Equal(t, "17TRACK", state.LastProviderCode)
	assert.Equal(t, uint(7), state.LastProviderID)
	assert.Equal(t, "TRACK-WEBHOOK-STATE", state.LastTrackingNumber)
	assert.Equal(t, "190271", state.LastCarrierCode)
	assert.Equal(t, uint(401), state.LastOrderID)
	assert.Equal(t, 3, state.LastEventCount)
	assert.Equal(t, 200, state.LastHTTPStatus)
	assert.True(t, state.LastAccepted)
	assert.True(t, state.LastSignatureChecked)
	assert.True(t, state.LastSignatureValid)
	assert.Empty(t, state.LastError)
}

func TestApplyTrackingWebhookReplacesEventsForExistingShipment(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "17TRACK",
		ProviderName: "17TRACK",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             100,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-100",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "pending",
		SyncStatus:          "pending",
		Enabled:             true,
	}).Error)

	eventTime := time.Now().Add(-time.Hour).UTC()
	result, err := shippingService.ApplyTrackingWebhook(TrackingWebhookInput{
		ProviderID:          provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-100",
		ProviderCarrierCode: "DHL",
		Events: []TrackingWebhookEventInput{
			{
				Status:      "Delivered",
				Location:    "Los Angeles, US",
				Description: "Delivered to recipient",
				EventTime:   eventTime,
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shipment)
	assert.Equal(t, "synced", result.Shipment.SyncStatus)
	assert.Equal(t, 1, result.Shipment.EventCount)
	require.Len(t, result.Events, 1)
	assert.Equal(t, "Delivered", result.Events[0].Status)

	events, err := shippingService.GetTrackingEventsByOrderID(100)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "Delivered", events[0].Status)
	assert.Equal(t, "Los Angeles, US", events[0].Location)
}

func TestApplyTrackingWebhookFallsBackToUniqueTrackingNumberWhenCarrierCodeDiffers(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "17TRACK",
		ProviderName: "17TRACK",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             101,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-FALLBACK",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		Enabled:             true,
	}).Error)

	eventTime := time.Now().Add(-time.Hour).UTC()
	result, err := shippingService.ApplyTrackingWebhook(TrackingWebhookInput{
		ProviderID:          provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-FALLBACK",
		ProviderCarrierCode: "190271",
		Events: []TrackingWebhookEventInput{
			{
				Status:      "InTransit",
				Location:    "Shenzhen, CN",
				Description: "Carrier code differs but tracking number is unique",
				EventTime:   eventTime,
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shipment)
	assert.Equal(t, uint(101), result.Shipment.OrderID)
	assert.Equal(t, "DHL", result.Shipment.ProviderCarrierCode)

	events, err := shippingService.GetTrackingEventsByOrderID(101)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "DHL", events[0].ProviderCarrierCode)
	assert.Equal(t, "Carrier code differs but tracking number is unique", events[0].Description)
}

func TestApplyTrackingWebhookRejectsAmbiguousTrackingNumberFallback(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "17TRACK",
		ProviderName: "17TRACK",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             102,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-DUP",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             103,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-DUP",
		ProviderCarrierCode: "UPS",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		Enabled:             true,
	}).Error)

	result, err := shippingService.ApplyTrackingWebhook(TrackingWebhookInput{
		ProviderID:          provider.ID,
		TrackingNumber:      "TRACK-WEBHOOK-DUP",
		ProviderCarrierCode: "190271",
		Events: []TrackingWebhookEventInput{
			{Status: "InTransit", Description: "Ambiguous duplicate"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "multiple tracking shipments")
}

func newTestShippingTrackingService(t *testing.T) (*gorm.DB, *ShippingService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingCarrierMapping{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	))

	shippingRepo := repository.NewShippingRepository(db)
	return db, NewShippingService(shippingRepo)
}

func seedTrackingProviderCarrierAndService(t *testing.T, db *gorm.DB) (shippingdomain.TrackingProviderConfig, shippingdomain.Carrier, shippingdomain.CarrierService) {
	t.Helper()

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode:          "17track",
		ProviderName:          "17TRACK",
		Environment:           "production",
		BaseURL:               "https://api.17track.net",
		APIKey:                "test-api-key",
		Enabled:               true,
		RequestTimeoutSeconds: 15,
	}
	require.NoError(t, db.Create(&provider).Error)

	carrier := shippingdomain.Carrier{
		Name:    "DHL",
		Code:    "DHL",
		Enabled: true,
	}
	require.NoError(t, db.Create(&carrier).Error)

	carrierService := shippingdomain.CarrierService{
		CarrierID:            carrier.ID,
		ServiceCode:          "DHL-EXP-US",
		ServiceName:          "DHL Express US",
		BillingMode:          "actual_weight",
		VolumetricDivisor:    6000,
		Enabled:              true,
		MinChargeWeightGrams: 1,
	}
	require.NoError(t, db.Create(&carrierService).Error)

	return provider, carrier, carrierService
}

func seedTrackingCarrierMapping(t *testing.T, db *gorm.DB, providerID uint, scope string, carrierID *uint, carrierServiceID *uint, providerCarrierCode string) shippingdomain.TrackingCarrierMapping {
	t.Helper()

	mapping := shippingdomain.TrackingCarrierMapping{
		ProviderID:          providerID,
		Scope:               scope,
		CarrierID:           carrierID,
		CarrierServiceID:    carrierServiceID,
		ProviderCarrierCode: providerCarrierCode,
		Enabled:             true,
	}
	require.NoError(t, db.Create(&mapping).Error)
	return mapping
}
