package service

import (
	"context"
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func newTemplateReadinessHandler(t *testing.T) (*CanonicalDomainEventOutboxHandler, *TransactionalNotificationTemplateService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))
	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	templateService := NewTransactionalNotificationTemplateService(repository.NewNotificationTemplateRepository(db))
	return NewCanonicalDomainEventOutboxHandler(templateService), templateService
}

func TestCanonicalDomainEventOutboxHandlerChecksPersistedTemplateReadiness(t *testing.T) {
	handler, templateService := newTemplateReadinessHandler(t)
	require.NoError(t, func() error {
		definition, err := LookupTransactionalNotificationTemplate(NotificationTemplateOrderDelivered)
		if err != nil {
			return err
		}
		_, err = templateService.Save(SaveTransactionalNotificationTemplateInput{
			Code:              definition.Code,
			Locale:            "en",
			Category:          notification.TemplateCategoryOrder,
			Name:              "Order delivered",
			SubjectTemplate:   "Order {{order_number}} delivered",
			BodyText:          "Order {{order_number}} delivered at {{delivered_at}}.",
			AllowedVariables:  definition.AllowedVariables,
			RequiredVariables: definition.RequiredVariables,
			IsEnabled:         true,
			Version:           1,
		})
		return err
	}())

	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{SchemaVersion: 1, IdempotencyKey: "delivery_fact:7:TRK-7"},
		OrderID:                      7,
		OrderNumber:                  "ORD-7",
		NewShippingStatus:            "delivered",
	})
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeOrderDelivered,
		Payload:   datatypes.JSON(payload),
	}))
}

func TestCanonicalDomainEventOutboxHandlerRejectsMissingPersistedTemplate(t *testing.T) {
	handler, _ := newTemplateReadinessHandler(t)
	payload, err := json.Marshal(outbox.OrderDeliveredPayload{
		CanonicalDomainEventMetadata: outbox.CanonicalDomainEventMetadata{SchemaVersion: 1, IdempotencyKey: "delivery_fact:8:TRK-8"},
		OrderID:                      8,
		OrderNumber:                  "ORD-8",
		NewShippingStatus:            "delivered",
	})
	require.NoError(t, err)
	err = handler.Handle(context.Background(), outbox.Event{EventType: outbox.EventTypeOrderDelivered, Payload: payload})
	assert.Error(t, err)
}
