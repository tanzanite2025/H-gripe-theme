package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentRiskFailOpenOutboxPublisherDeduplicatesIncidentBucket(t *testing.T) {
	db := newPaymentRiskFailOpenAlertTestDB(t)
	publisher := NewPaymentRiskFailOpenOutboxPublisher(repository.NewOutboxRepository(db))
	now := time.Date(2026, time.September, 2, 3, 0, 0, 0, time.UTC)
	input := PaymentRiskFailOpenAlertInput{
		Provider:       "Stripe",
		Component:      "antifraud",
		Operation:      "payment_3ds_policy_evaluate",
		Reason:         "payment_risk_unavailable",
		FallbackAction: "step_up_3ds_any",
		OccurredAt:     now,
		Err:            errors.New("redis unavailable"),
	}

	require.NoError(t, publisher.PublishPaymentRiskFailOpenAlert(context.Background(), input))
	require.NoError(t, publisher.PublishPaymentRiskFailOpenAlert(context.Background(), input))

	var events []outbox.Event
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	require.Len(t, events, 1)
	require.Equal(t, outbox.EventTypePaymentRiskFailOpen, events[0].EventType)
	require.Equal(t, "payment.risk_fail_open:stripe:antifraud:payment_3ds_policy_evaluate:202609020300", events[0].EventKey)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
	require.Equal(t, "stripe", payload["provider"])
	require.Equal(t, "critical", payload["severity"])
	require.Equal(t, true, payload["fail_open"])
	require.Equal(t, true, payload["redis_unavailable"])
	require.Equal(t, "step_up_3ds_any", payload["fallback_action"])
	require.NotContains(t, string(events[0].Payload), "order_id")
	require.NotContains(t, string(events[0].Payload), "user_id")

	input.OccurredAt = now.Add(6 * time.Minute)
	require.NoError(t, publisher.PublishPaymentRiskFailOpenAlert(context.Background(), input))
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	require.Len(t, events, 2)
}

func newPaymentRiskFailOpenAlertTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:payment-risk-fail-open-alert?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	return db
}
