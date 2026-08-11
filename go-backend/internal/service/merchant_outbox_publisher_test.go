package service

import (
	"encoding/json"
	"strings"
	"testing"

	outboxdomain "commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestMerchantOutboxPublisherEnqueuesTypedProductEvents(t *testing.T) {
	db, _ := newTestOutboxService(t)
	publisher := NewMerchantOutboxPublisher(repository.NewOutboxRepository(db))

	require.NoError(t, publisher.EnqueueProductUpsert(21, "product_price_changed"))
	require.NoError(t, publisher.EnqueueProductWithdraw(21, "product_deleted"))
	require.NoError(t, publisher.EnqueueOfferRevalidate(34, "merchant_fields_changed"))

	var events []outboxdomain.Event
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	require.Len(t, events, 3)

	require.Equal(t, outboxdomain.EventTypeMerchantProductUpsert, events[0].EventType)
	require.Equal(t, outboxdomain.AggregateTypeProduct, events[0].AggregateType)
	require.Equal(t, "21", events[0].AggregateID)
	require.Contains(t, events[0].EventKey, outboxdomain.EventTypeMerchantProductUpsert+":21:")
	assertMerchantProductPayload(t, events[0], 21, "product_price_changed")

	require.Equal(t, outboxdomain.EventTypeMerchantProductWithdraw, events[1].EventType)
	require.Equal(t, outboxdomain.AggregateTypeProduct, events[1].AggregateType)
	require.Equal(t, "21", events[1].AggregateID)
	assertMerchantProductPayload(t, events[1], 21, "product_deleted")

	require.Equal(t, outboxdomain.EventTypeMerchantOfferRevalidate, events[2].EventType)
	require.Equal(t, outboxdomain.AggregateTypeMerchantOffer, events[2].AggregateType)
	require.Equal(t, "34", events[2].AggregateID)

	var revalidate outboxdomain.MerchantOfferRevalidatePayload
	require.NoError(t, json.Unmarshal(events[2].Payload, &revalidate))
	require.Equal(t, uint(34), revalidate.OfferID)
	require.Equal(t, "merchant_fields_changed", revalidate.Reason)
}

func TestMerchantOutboxPublisherIgnoresEmptyIdentifiers(t *testing.T) {
	db, _ := newTestOutboxService(t)
	publisher := NewMerchantOutboxPublisher(repository.NewOutboxRepository(db))

	require.NoError(t, publisher.EnqueueProductUpsert(0, "invalid"))
	require.NoError(t, publisher.EnqueueProductWithdraw(0, "invalid"))
	require.NoError(t, publisher.EnqueueOfferRevalidate(0, "invalid"))

	var count int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Count(&count).Error)
	require.Zero(t, count)
}

func assertMerchantProductPayload(t *testing.T, event outboxdomain.Event, productID uint, reason string) {
	t.Helper()

	var payload outboxdomain.MerchantProductSyncPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, productID, payload.ProductID)
	require.Equal(t, reason, payload.Reason)
	require.True(t, strings.Contains(string(event.Payload), `"product_id"`))
}
