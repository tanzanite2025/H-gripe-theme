package service

import (
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	orderdomain "commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shippingAuditRecorder struct {
	logs []audit.AuditLog
	err  error
}

func (r *shippingAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return r.err
}

func TestApplyTrackingWebhookRecordsFirstDeliveryAuditOnly(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	auditRecorder := &shippingAuditRecorder{}
	shippingService.ConfigureAuditRecorder(auditRecorder)

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode:  "17TRACK",
		ProviderName:  "17TRACK",
		APIKey:        "provider-api-key-must-not-be-audit-data",
		WebhookSecret: "provider-webhook-secret-must-not-be-audit-data",
		Enabled:       true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&orderdomain.Order{
		ID:             601,
		OrderNumber:    "ORDER-DELIVERY-AUDIT-601",
		ShippingStatus: "shipped",
		Currency:       "USD",
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             601,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-DELIVERY-AUDIT-601",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "pending",
		Enabled:             true,
	}).Error)

	input := TrackingWebhookInput{
		ProviderID:          provider.ID,
		TrackingNumber:      "TRACK-DELIVERY-AUDIT-601",
		ProviderCarrierCode: "DHL",
		Events: []TrackingWebhookEventInput{{
			Status:                 "Delivered",
			Location:               "private delivery location",
			Description:            "private delivery description",
			RecipientSignatureName: "private recipient name",
			ProofOfDeliveryURL:     "https://carrier.example.test/private-pod.pdf",
			EventTime:              time.Date(2026, 9, 5, 7, 0, 0, 0, time.UTC),
		}},
	}

	_, err := shippingService.ApplyTrackingWebhook(input)
	require.NoError(t, err)
	_, err = shippingService.ApplyTrackingWebhook(input)
	require.NoError(t, err)

	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	assert.Equal(t, "execute", log.Action)
	assert.Equal(t, "order_delivery", log.Resource)
	assert.Equal(t, uint(601), log.ResourceID)
	assert.Equal(t, "success", log.Status)
	assert.Contains(t, log.Changes, `"source":"tracking_webhook"`)
	assert.Contains(t, log.Changes, `"tracking_status":"Delivered"`)
	assert.Contains(t, log.OldValue, `"shipping_status":"shipped"`)
	assert.Contains(t, log.NewValue, `"shipping_status":"delivered"`)
	assert.NotContains(t, log.Changes, "provider-api-key-must-not-be-audit-data")
	assert.NotContains(t, log.Changes, "provider-webhook-secret-must-not-be-audit-data")
	assert.NotContains(t, log.Changes, "private delivery location")
	assert.NotContains(t, log.Changes, "private-pod.pdf")
}

func TestTrackingDeliveryAuditFailureDoesNotBlockBusinessState(t *testing.T) {
	db, shippingService := newTestShippingTrackingService(t)
	shippingService.ConfigureAuditRecorder(&shippingAuditRecorder{
		err: errors.New("audit storage unavailable"),
	})

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "17TRACK",
		ProviderName: "17TRACK",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	require.NoError(t, db.Create(&orderdomain.Order{
		ID:             602,
		OrderNumber:    "ORDER-DELIVERY-AUDIT-602",
		ShippingStatus: "shipped",
		Currency:       "USD",
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             602,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-DELIVERY-AUDIT-602",
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "pending",
		Enabled:             true,
	}).Error)

	_, err := shippingService.ApplyTrackingWebhook(TrackingWebhookInput{
		ProviderID:          provider.ID,
		TrackingNumber:      "TRACK-DELIVERY-AUDIT-602",
		ProviderCarrierCode: "DHL",
		Events: []TrackingWebhookEventInput{{
			Status:    "Delivered",
			EventTime: time.Now().UTC(),
		}},
	})
	require.NoError(t, err)

	var stored orderdomain.Order
	require.NoError(t, db.First(&stored, 602).Error)
	assert.Equal(t, "delivered", stored.ShippingStatus)
}
