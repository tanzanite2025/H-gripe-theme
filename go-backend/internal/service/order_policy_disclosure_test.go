package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	refundcancellationdomain "commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestCreateOrderCapturesHistoricalRefundCancellationPolicyDisclosure(t *testing.T) {
	db, orderService := newTestOrderService(t)
	settings := repository.NewSettingRepository(db)
	policyService := NewRefundCancellationPolicyService(settings)
	orderService.ConfigureRefundCancellationPolicy(policyService)

	policy := refundcancellationdomain.DefaultPolicy()
	policy.Title = "Refund policy version one"
	payload, err := json.Marshal(policy)
	require.NoError(t, err)
	require.NoError(t, settings.Set(&setting.Setting{
		Key:         refundcancellationdomain.Key,
		Value:       string(payload),
		Type:        "json",
		Locale:      "en",
		Group:       refundcancellationdomain.Group,
		IsPublic:    true,
		Description: "test policy",
	}))

	productRecord := seedProduct(t, db, 50, 5)
	created, err := orderService.CreateOrderWithAttributionAndOptions(
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
		OrderCreationOptions{
			PolicyLocale:                 "en",
			PolicyURL:                    "/policies/refund-cancellation",
			PolicyDisclosureAcknowledged: true,
			PolicySource:                 "checkout_test",
		},
	)
	require.NoError(t, err)

	var disclosure order.PolicyDisclosure
	require.NoError(t, db.Where("order_id = ?", created.ID).First(&disclosure).Error)
	require.Equal(t, refundcancellationdomain.Key, disclosure.PolicyKey)
	require.Equal(t, "en", disclosure.Locale)
	require.Equal(t, "/policies/refund-cancellation", disclosure.PolicyURL)
	require.Equal(t, "checkout_test", disclosure.Source)
	require.NotEmpty(t, disclosure.PolicyHash)
	require.NotNil(t, disclosure.ConsentedAt)

	var savedPolicy refundcancellationdomain.Policy
	require.NoError(t, json.Unmarshal([]byte(disclosure.PolicyJSON), &savedPolicy))
	require.Equal(t, "Refund policy version one", savedPolicy.Title)

	updatedPolicy := policy
	updatedPolicy.Title = "Refund policy version two"
	updatedPayload, err := json.Marshal(updatedPolicy)
	require.NoError(t, err)
	require.NoError(t, settings.Set(&setting.Setting{
		Key:         refundcancellationdomain.Key,
		Value:       string(updatedPayload),
		Type:        "json",
		Locale:      "en",
		Group:       refundcancellationdomain.Group,
		IsPublic:    true,
		Description: "test policy",
	}))

	var unchanged order.PolicyDisclosure
	require.NoError(t, db.First(&unchanged, disclosure.ID).Error)
	require.Equal(t, "Refund policy version one", savedPolicy.Title)
	require.Equal(t, disclosure.PolicyHash, unchanged.PolicyHash)
	require.WithinDuration(t, disclosure.DisclosedAt, unchanged.DisclosedAt, time.Second)
}
