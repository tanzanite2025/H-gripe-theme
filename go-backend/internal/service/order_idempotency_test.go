package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/order"
	attributionpkg "commerce-platform/internal/pkg/attribution"

	"github.com/stretchr/testify/require"
)

func TestCreateOrderWithIdempotencyReplaysDurableOrder(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)
	options := OrderCreationOptions{
		IdempotencyKey:         "checkout-key-1",
		IdempotencyRequestHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	first, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		options,
	)
	require.NoError(t, err)

	second, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		options,
	)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)

	var orderCount int64
	require.NoError(t, db.Model(&order.Order{}).Where("user_id = ?", 42).Count(&orderCount).Error)
	require.Equal(t, int64(1), orderCount)

	var idempotencyRecord order.OrderIdempotency
	require.NoError(t, db.Where("user_id = ? AND idempotency_key = ?", 42, options.IdempotencyKey).First(&idempotencyRecord).Error)
	require.NotNil(t, idempotencyRecord.OrderID)
	require.Equal(t, first.ID, *idempotencyRecord.OrderID)
}

func TestCreateOrderWithIdempotencyRejectsDifferentPayloadHash(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 50, 5)
	base := OrderCreationOptions{
		IdempotencyKey:         "checkout-key-2",
		IdempotencyRequestHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	_, err := orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		base,
	)
	require.NoError(t, err)

	conflicting := base
	conflicting.IdempotencyRequestHash = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	_, err = orderService.CreateOrderWithAttributionAndOptions(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		testAddress(),
		testAddress(),
		"card",
		"standard",
		"",
		0,
		attributionpkg.Context{},
		conflicting,
	)
	require.ErrorIs(t, err, ErrOrderIdempotencyConflict)
}
