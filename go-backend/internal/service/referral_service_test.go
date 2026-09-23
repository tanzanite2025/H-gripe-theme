package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	coupondomain "commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	orderdomain "commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestReferralBindingCreatesPermanentPendingAttribution(t *testing.T) {
	service, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "referrer@example.test", "referrer")
	referee := createReferralTestUser(t, db, "referee@example.test", "referee")

	dashboard, err := service.Dashboard(referrer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, dashboard.ReferralCode)
	require.Contains(t, dashboard.ShareURL, "/r/")

	token, maxAge, err := service.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	assert.Equal(t, 30*24*60*60, maxAge)

	record, err := service.BindFromToken(referee.ID, token, "203.0.113.10")
	require.NoError(t, err)
	assert.Equal(t, loyalty.ReferralStatusPending, record.Status)
	assert.Equal(t, referrer.ID, record.ReferrerID)
	assert.Equal(t, referee.ID, *record.RefereeID)
	assert.Len(t, record.ClientIPHash, 64)
	assert.NotContains(t, record.ClientIPHash, "203.0.113.10")

	var transitionCount int64
	require.NoError(t, db.Model(&loyalty.ReferralTransition{}).
		Where("referral_record_id = ?", record.ID).
		Count(&transitionCount).Error)
	assert.Equal(t, int64(1), transitionCount)

	// Replaying the same signed attribution is an idempotent success.
	replayed, err := service.BindFromToken(referee.ID, token, "203.0.113.10")
	require.NoError(t, err)
	assert.Equal(t, record.ID, replayed.ID)
	require.NoError(t, db.Model(&loyalty.ReferralTransition{}).
		Where("referral_record_id = ?", record.ID).
		Count(&transitionCount).Error)
	assert.Equal(t, int64(1), transitionCount)
}

func TestReferralDashboardUsesStorefrontURLForShareURL(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referralService = NewReferralService(referralService.txManager, referralService.repo, referralService.programRepo,
		referralService.userRepo, "http://api.test", "test-referral-service-secret", "https://shop.test")
	referrer := createReferralTestUser(t, db, "share-url-referrer@example.test", "share-url-referrer")

	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, dashboard.ReferralCode)
	assert.Equal(t, "https://shop.test/r/"+dashboard.ReferralCode, dashboard.ShareURL)
}

func TestReferralBindingReleasesRefereePointsAtRegistration(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "bind-points-referrer@example.test", "bind-points-referrer")
	referee := createReferralTestUser(t, db, "bind-points-referee@example.test", "bind-points-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitPoints
	config.RefereeBenefitValue = 75
	require.NoError(t, db.Save(config).Error)
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromToken(referee.ID, token, "203.0.113.14")
	require.NoError(t, err)

	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusReleased, reward.Status)
	assert.Equal(t, loyalty.ReferralRewardTypePoints, reward.RewardType)
	assert.Equal(t, 75, reward.PointsAmount)
	var balance loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", referee.ID).First(&balance).Error)
	assert.Equal(t, 75, balance.AvailablePoints)
	var transactionCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", "referral_referee", record.ID).Count(&transactionCount).Error)
	assert.Equal(t, int64(1), transactionCount)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	_, err = loyaltyRepo.AdjustUserPoints(referee.ID, 25, "earn", "order", 7001, "Order completion points")
	require.NoError(t, err)
	_, err = loyaltyRepo.AdjustUserPoints(referee.ID, -80, "spend", "order", 7002, "Spent aggregate points")
	require.NoError(t, err)
	require.NoError(t, db.Where("user_id = ?", referee.ID).First(&balance).Error)
	assert.Equal(t, 20, balance.AvailablePoints)
}

func TestPublishAdminReferralProgramConfigCreatesNewVersion(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, false)
	operator := createReferralTestUser(t, db, "operator@example.test", "operator")

	config, err := referralService.PublishAdminProgramConfig(ReferralProgramConfigInput{
		Enabled:                 true,
		MinOrderAmountMinor:     25000,
		ReferrerRewardPoints:    1250,
		RefereeBenefitType:      loyalty.ReferralBenefitPoints,
		RefereeBenefitValue:     500,
		VestingPeriodDays:       30,
		UndeliveredFallbackDays: 45,
		AttributionTTLDays:      30,
		MonthlyCapPerReferrer:   10,
		AntiFraudMode:           loyalty.ReferralFraudModeMonitor,
		CreatedBy:               &operator.ID,
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, config.Version)
	assert.True(t, config.Enabled)
	assert.Equal(t, "USD", config.Currency)

	active, err := referralService.GetAdminProgramConfig()
	require.NoError(t, err)
	assert.Equal(t, config.ID, active.ID)
	assert.Equal(t, 1250, active.ReferrerRewardPoints)
}

func TestPublishAdminReferralProgramConfigRejectsInvalidWindows(t *testing.T) {
	referralService, _ := newReferralServiceFixture(t, false)
	_, err := referralService.PublishAdminProgramConfig(ReferralProgramConfigInput{
		ReferrerRewardPoints:    1000,
		RefereeBenefitType:      loyalty.ReferralBenefitPoints,
		VestingPeriodDays:       45,
		UndeliveredFallbackDays: 30,
		AttributionTTLDays:      30,
		MonthlyCapPerReferrer:   10,
		AntiFraudMode:           loyalty.ReferralFraudModeMonitor,
	}, 1)
	assert.ErrorIs(t, err, ErrInvalidReferralProgramConfig)
}

func TestReferralBindingRejectsSelfAndFormerPurchaser(t *testing.T) {
	service, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "referrer-2@example.test", "referrer2")
	formerCustomer := createReferralTestUser(t, db, "former@example.test", "former")
	dashboard, err := service.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := service.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)

	_, err = service.BindFromToken(referrer.ID, token, "203.0.113.11")
	assert.ErrorIs(t, err, ErrSelfReferralForbidden)

	paidAt := time.Now().UTC().Add(-24 * time.Hour)
	require.NoError(t, db.Create(&orderdomain.Order{
		OrderNumber:         "REF-PAID-1",
		UserID:              formerCustomer.ID,
		Status:              "refunded",
		PaymentStatus:       "refunded",
		SubtotalAmountMinor: 25000,
		TotalAmountMinor:    25000,
		Currency:            "USD",
		PaidAt:              &paidAt,
	}).Error)
	_, err = service.BindFromToken(formerCustomer.ID, token, "203.0.113.12")
	assert.ErrorIs(t, err, ErrRefereeNotEligible)
}

func TestDisabledReferralProgramDoesNotAllocateIdentity(t *testing.T) {
	service, db := newReferralServiceFixture(t, false)
	account := createReferralTestUser(t, db, "disabled@example.test", "disabled")

	dashboard, err := service.Dashboard(account.ID)
	require.NoError(t, err)
	assert.False(t, dashboard.Enabled)
	assert.Empty(t, dashboard.ReferralCode)

	var identityCount int64
	require.NoError(t, db.Model(&loyalty.ReferralIdentity{}).Count(&identityCount).Error)
	assert.Zero(t, identityCount)
}

func TestReferralValidationTreatsMalformedCodeAsUnavailable(t *testing.T) {
	service, _ := newReferralServiceFixture(t, true)

	_, err := service.ValidateCode("bad!")
	assert.ErrorIs(t, err, ErrReferralCodeNotFound)
}

func TestReferralLifecycleOutboxRunsInShadowMode(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "lifecycle-referrer@example.test", "lifecycle-referrer")
	referee := createReferralTestUser(t, db, "lifecycle-referee@example.test", "lifecycle-referee")
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	bound, err := referralService.BindFromToken(referee.ID, token, "203.0.113.20")
	require.NoError(t, err)

	paidAt := time.Date(2026, 9, 14, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{
		OrderNumber:         "REF-LIFECYCLE-1",
		UserID:              referee.ID,
		Status:              "processing",
		PaymentStatus:       "paid",
		SubtotalAmountMinor: 25000,
		TotalAmountMinor:    25000,
		Currency:            "USD",
		PaidAt:              &paidAt,
	}
	require.NoError(t, db.Create(orderRecord).Error)

	paidPayload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{
		OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt,
	})
	require.NoError(t, err)
	paidEvent := outboxdomain.Event{
		EventKey: "referral.order_paid:test-1", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: paidPayload,
	}
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), paidEvent))
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), paidEvent))

	var record loyalty.ReferralRecord
	require.NoError(t, db.First(&record, bound.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusOrdered, record.Status)
	require.NotNil(t, record.OrderID)
	assert.Equal(t, orderRecord.ID, *record.OrderID)
	assert.Equal(t, int64(25000), record.OrderAmountMinor)

	deliveredAt := paidAt.Add(7 * 24 * time.Hour)
	deliveredPayload, err := json.Marshal(outboxdomain.ReferralOrderDeliveredPayload{
		OrderID: orderRecord.ID, DeliveredAt: deliveredAt, Source: "tracking_webhook",
	})
	require.NoError(t, err)
	deliveredEvent := outboxdomain.Event{
		EventKey: "referral.order_delivered:test-1", EventType: outboxdomain.EventTypeReferralOrderDelivered, Payload: deliveredPayload,
	}
	require.NoError(t, referralService.HandleOrderDeliveredOutbox(context.Background(), deliveredEvent))
	require.NoError(t, referralService.HandleOrderDeliveredOutbox(context.Background(), deliveredEvent))
	require.NoError(t, db.First(&record, bound.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusVesting, record.Status)
	require.NotNil(t, record.DeliveredAt)
	require.NotNil(t, record.VestingUntil)
	assert.Equal(t, deliveredAt, record.DeliveredAt.UTC())
	assert.Equal(t, deliveredAt.AddDate(0, 0, 30), record.VestingUntil.UTC())

	invalidatedAt := deliveredAt.Add(24 * time.Hour)
	invalidatedPayload, err := json.Marshal(outboxdomain.ReferralOrderInvalidatedPayload{
		OrderID: orderRecord.ID, OccurredAt: invalidatedAt, Reason: "order fully refunded", Source: "refund_webhook", Reference: "refund_1",
	})
	require.NoError(t, err)
	invalidatedEvent := outboxdomain.Event{
		EventKey: "referral.order_invalidated:test-1", EventType: outboxdomain.EventTypeReferralOrderInvalidated, Payload: invalidatedPayload,
	}
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), invalidatedEvent))
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), invalidatedEvent))
	require.NoError(t, db.First(&record, bound.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusRevoked, record.Status)
	assert.Equal(t, "order fully refunded", record.RevokeReason)

	var transitionCount int64
	require.NoError(t, db.Model(&loyalty.ReferralTransition{}).Where("referral_record_id = ?", bound.ID).Count(&transitionCount).Error)
	assert.Equal(t, int64(4), transitionCount)
	var rewardCount int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).Count(&rewardCount).Error)
	// Registration points are credited at binding and remain ordinary account
	// points even when the order-side referral record is later revoked.
	assert.Equal(t, int64(1), rewardCount)
}

func TestDeliveredTrackingUpdatePersistsTimestampAndReferralEventOnce(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	account := createReferralTestUser(t, db, "delivery@example.test", "delivery")
	now := time.Now().UTC()
	createdAt := now.Add(-72 * time.Hour)
	shippedAt := now.Add(-48 * time.Hour)
	deliveredAt := now.Add(-24 * time.Hour)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-DELIVERY-1", UserID: account.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", ShippingStatus: "shipped",
		CreatedAt: createdAt, ShippedAt: &shippedAt,
	}
	require.NoError(t, db.Create(orderRecord).Error)

	shippingService := NewShippingService(nil)
	shippingService.ConfigureOrderRepository(repository.NewOrderRepository(db))
	shippingService.ConfigureTxManager(referralService.txManager)
	events := []shippingdomain.TrackingEvent{{OrderID: orderRecord.ID, Status: "Delivered", EventTime: deliveredAt}}
	require.NoError(t, shippingService.updateOrderShippingStatusIfDelivered(
		orderRecord.ID, "TRACK-1", "carrier", "Delivered", 4, events, "tracking_webhook",
	))
	require.NoError(t, shippingService.updateOrderShippingStatusIfDelivered(
		orderRecord.ID, "TRACK-1", "carrier", "Delivered", 4, events, "tracking_webhook",
	))

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "delivered", savedOrder.ShippingStatus)
	require.NotNil(t, savedOrder.DeliveredAt)
	assert.Equal(t, deliveredAt, savedOrder.DeliveredAt.UTC())

	var eventCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).
		Where("event_type = ? AND aggregate_id = ?", outboxdomain.EventTypeReferralOrderDelivered, fmt.Sprint(orderRecord.ID)).
		Count(&eventCount).Error)
	assert.Equal(t, int64(1), eventCount)
}

func TestDeliveredTrackingTimestampSanitizesFutureAndInvertedTimes(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	account := createReferralTestUser(t, db, "delivery-boundary@example.test", "delivery-boundary")
	now := time.Now().UTC()
	createdAt := now.Add(-72 * time.Hour)
	shippedAt := now.Add(-48 * time.Hour)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-DELIVERY-BOUNDARY", UserID: account.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", ShippingStatus: "shipped",
		CreatedAt: createdAt, ShippedAt: &shippedAt,
	}
	require.NoError(t, db.Create(orderRecord).Error)

	shippingService := NewShippingService(nil)
	shippingService.ConfigureOrderRepository(repository.NewOrderRepository(db))
	shippingService.ConfigureTxManager(referralService.txManager)

	futureAt := time.Now().UTC().Add(4 * 24 * time.Hour)
	require.NoError(t, shippingService.updateOrderShippingStatusIfDelivered(
		orderRecord.ID, "TRACK-FUTURE", "carrier", "Delivered", 4,
		[]shippingdomain.TrackingEvent{{OrderID: orderRecord.ID, Status: "Delivered", EventTime: futureAt}},
		"tracking_webhook",
	))

	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.NotNil(t, savedOrder.DeliveredAt)
	assert.False(t, savedOrder.DeliveredAt.After(time.Now().UTC().Add(time.Hour)))

	// A second order verifies the lower bound independently because delivery is
	// persisted only once for each order.
	secondOrder := &orderdomain.Order{
		OrderNumber: "REF-DELIVERY-INVERTED", UserID: account.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", ShippingStatus: "shipped",
		CreatedAt: createdAt, ShippedAt: &shippedAt,
	}
	require.NoError(t, db.Create(secondOrder).Error)
	require.NoError(t, shippingService.updateOrderShippingStatusIfDelivered(
		secondOrder.ID, "TRACK-INVERTED", "carrier", "Delivered", 4,
		[]shippingdomain.TrackingEvent{{OrderID: secondOrder.ID, Status: "Delivered", EventTime: createdAt.Add(-24 * time.Hour)}},
		"tracking_webhook",
	))

	savedOrder = orderdomain.Order{}
	require.NoError(t, db.First(&savedOrder, secondOrder.ID).Error)
	require.NotNil(t, savedOrder.DeliveredAt)
	assert.Equal(t, shippedAt.UTC(), savedOrder.DeliveredAt.UTC())
}

func TestReferralLifecycleScanKeepsBoundPendingPermanentAndAdvancesDelivery(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "scan-referrer@example.test", "scan-referrer")
	referee := createReferralTestUser(t, db, "scan-referee@example.test", "scan-referee")
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	pending, err := referralService.BindFromToken(referee.ID, token, "203.0.113.30")
	require.NoError(t, err)

	paidAt := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-SCAN-DELIVERED", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt,
		ShippingStatus: "delivered", DeliveredAt: referralPtrTime(paidAt.AddDate(0, 0, 2)),
	}
	require.NoError(t, db.Create(orderRecord).Error)
	require.NoError(t, db.Model(&loyalty.ReferralRecord{}).Where("id = ?", pending.ID).Updates(map[string]any{
		"order_id": orderRecord.ID, "status": loyalty.ReferralStatusOrdered, "ordered_at": paidAt,
	}).Error)

	permanentReferee := createReferralTestUser(t, db, "scan-permanent@example.test", "scan-permanent")
	permanentToken, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	permanent, err := referralService.BindFromToken(permanentReferee.ID, permanentToken, "203.0.113.31")
	require.NoError(t, err)
	// The persisted timestamp is retained as an audit snapshot of the signed
	// cookie, but it must not expire a relationship after account binding.
	require.NoError(t, db.Model(&loyalty.ReferralRecord{}).Where("id = ?", permanent.ID).Update("expires_at", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)).Error)

	result, err := referralService.ScanLifecycle(context.Background(), time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), 100)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Expired)
	assert.Equal(t, 1, result.AdvancedToVesting)

	var savedPending, savedPermanent loyalty.ReferralRecord
	require.NoError(t, db.First(&savedPending, pending.ID).Error)
	require.NoError(t, db.First(&savedPermanent, permanent.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusSettled, savedPending.Status)
	assert.Equal(t, orderRecord.DeliveredAt.UTC(), savedPending.DeliveredAt.UTC())
	assert.Equal(t, orderRecord.DeliveredAt.UTC().AddDate(0, 0, 30), savedPending.VestingUntil.UTC())
	assert.Equal(t, loyalty.ReferralStatusPending, savedPermanent.Status)

	var rewards int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).Count(&rewards).Error)
	assert.Equal(t, int64(3), rewards)
}

func TestReferralLifecycleScanUsesUndeliveredFallbackWithoutFabricatingDelivery(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "fallback-referrer@example.test", "fallback-referrer")
	referee := createReferralTestUser(t, db, "fallback-referee@example.test", "fallback-referee")
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromToken(referee.ID, token, "203.0.113.32")
	require.NoError(t, err)

	shippedAt := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-SCAN-FALLBACK", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: referralPtrTime(shippedAt.Add(-24 * time.Hour)),
		ShippingStatus: "shipped", ShippedAt: &shippedAt,
	}
	require.NoError(t, db.Create(orderRecord).Error)
	require.NoError(t, db.Model(&loyalty.ReferralRecord{}).Where("id = ?", record.ID).Updates(map[string]any{
		"order_id": orderRecord.ID, "status": loyalty.ReferralStatusOrdered, "ordered_at": shippedAt.Add(-24 * time.Hour),
	}).Error)

	now := shippedAt.AddDate(0, 0, 45)
	result, err := referralService.ScanLifecycle(context.Background(), now, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, result.AdvancedToVesting)

	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusSettled, saved.Status)
	assert.Nil(t, saved.DeliveredAt)
	assert.Equal(t, now, saved.VestingUntil.UTC())
	assert.GreaterOrEqual(t, result.MaturedVestingCandidates, 1)
}

func TestReferralBindContextStoresOnlySensitiveHashes(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "hash-referrer@example.test", "hash-referrer")
	referee := createReferralTestUser(t, db, "hash-referee@example.test", "hash-referee")
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromTokenWithContext(referee.ID, token, ReferralBindContext{
		ClientIP: "203.0.113.40", DeviceFingerprint: "device-fingerprint-raw", RefereeEmail: referee.Email,
	})
	require.NoError(t, err)
	assert.Len(t, record.RefereeEmailHash, 64)
	assert.Len(t, record.DeviceFingerprintHash, 64)
	assert.NotContains(t, record.RefereeEmailHash, referee.Email)
	assert.NotContains(t, record.DeviceFingerprintHash, "device-fingerprint-raw")
}

func TestReferralPaymentRiskSignalsMonitorModeDoesNotBlockOrder(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "risk-referrer@example.test", "risk-referrer")
	referee := createReferralTestUser(t, db, "risk-referee@example.test", "risk-referee")
	priorPaidAt := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	sharedAddress := orderdomain.Address{
		Address1: "1 Shared Way", City: "Portland", State: "OR", PostalCode: "97201", Country: "US", Phone: "+1 (503) 555-0100",
	}
	require.NoError(t, db.Create(&orderdomain.Order{
		OrderNumber: "REF-RISK-REFERRER", UserID: referrer.ID, Status: "completed", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &priorPaidAt, ShippingAddress: sharedAddress,
	}).Error)

	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromTokenWithContext(referee.ID, token, ReferralBindContext{
		ClientIP: "203.0.113.41", DeviceFingerprint: "risk-device", RefereeEmail: referee.Email,
	})
	require.NoError(t, err)

	paidAt := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-RISK-REFEREE", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt, ShippingAddress: sharedAddress,
	}
	require.NoError(t, db.Create(orderRecord).Error)

	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{
		OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt,
	})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{
		EventKey: "referral.order_paid:risk-test", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload,
	}))

	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusOrdered, saved.Status)
	var flags []map[string]any
	require.NoError(t, json.Unmarshal(saved.RiskFlags, &flags))
	assert.Len(t, flags, 2)
	assert.Contains(t, string(saved.RiskFlags), "shipping_address_match")
	assert.Contains(t, string(saved.RiskFlags), "shipping_phone_match")
}

func TestReferralPaymentRiskSignalsStrictModeRevokesOrder(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "strict-risk-referrer@example.test", "strict-risk-referrer")
	referee := createReferralTestUser(t, db, "strict-risk-referee@example.test", "strict-risk-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.AntiFraudMode = loyalty.ReferralFraudModeStrict
	require.NoError(t, db.Save(config).Error)
	sharedAddress := orderdomain.Address{
		Address1: "9 Collision Road", City: "Portland", State: "OR", PostalCode: "97201", Country: "US", Phone: "+1 (503) 555-0199",
	}
	priorPaidAt := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&orderdomain.Order{
		OrderNumber: "REF-STRICT-REFERRER", UserID: referrer.ID, Status: "completed", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &priorPaidAt, ShippingAddress: sharedAddress,
	}).Error)
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromToken(referee.ID, token, "203.0.113.42")
	require.NoError(t, err)

	paidAt := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-STRICT-REFEREE", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt, ShippingAddress: sharedAddress,
	}
	require.NoError(t, db.Create(orderRecord).Error)
	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{
		EventKey: "referral.order_paid:strict-risk-test", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload,
	}))

	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusRevoked, saved.Status)
	assert.Contains(t, saved.RevokeReason, "strict anti-fraud")
	var rewardCount int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).Where("referral_record_id = ?", record.ID).Count(&rewardCount).Error)
	assert.Equal(t, int64(1), rewardCount)
}

func TestReferralOrderPaidEnforcesMonthlyCap(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "cap-referrer@example.test", "cap-referrer")
	referee := createReferralTestUser(t, db, "cap-referee@example.test", "cap-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.MonthlyCapPerReferrer = 1
	require.NoError(t, db.Save(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "CAP23456", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	firstOrderedAt := time.Date(2026, 9, 2, 8, 0, 0, 0, time.UTC)
	firstReferee := createReferralTestUser(t, db, "cap-first@example.test", "cap-first")
	firstRecord := &loyalty.ReferralRecord{ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &firstReferee.ID, ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusOrdered, RecordVersion: 1, ExpiresAt: firstOrderedAt.Add(24 * time.Hour), OrderedAt: &firstOrderedAt}
	require.NoError(t, db.Create(firstRecord).Error)
	record := &loyalty.ReferralRecord{ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID, ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusPending, RecordVersion: 1, ExpiresAt: firstOrderedAt.AddDate(0, 0, 30)}
	require.NoError(t, db.Create(record).Error)
	paidAt := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{OrderNumber: "REF-CAP-SECOND", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{EventKey: "referral.order_paid:cap-test", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload}))
	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusRevoked, saved.Status)
	assert.Contains(t, saved.RevokeReason, "monthly referral cap")
}

func referralPtrTime(value time.Time) *time.Time {
	return &value
}

func newReferralServiceFixture(t *testing.T, enabled bool) (*ReferralService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:referral-service-%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&user.User{},
		&orderdomain.Order{},
		&paymentdomain.Transaction{},
		&coupondomain.Coupon{},
		&coupondomain.CouponUsage{},
		&loyalty.ReferralProgramConfig{},
		&loyalty.ReferralIdentity{},
		&loyalty.ReferralRecord{},
		&loyalty.ReferralReward{},
		&loyalty.ReferralTransition{},
		&loyalty.UserLoyalty{},
		&loyalty.LoyaltyTransaction{},
		&outboxdomain.Event{},
	))

	config := referralProgramConfigForServiceTest(enabled)
	require.NoError(t, db.Create(config).Error)
	orderRepo := repository.NewOrderRepository(db)
	referralRepo := repository.NewReferralRepository(db)
	referralProgramRepo := repository.NewReferralProgramRepository(db)
	txManager := repository.NewTxManager(
		db,
		orderRepo,
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigureReferralRepositories(referralRepo, referralProgramRepo)
	txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	return NewReferralService(
		txManager,
		referralRepo,
		referralProgramRepo,
		repository.NewUserRepository(db),
		"https://shop.example.test",
		"test-referral-service-secret",
	), db
}

func TestSettleReferralIsIdempotentAndReleasesPoints(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "settle-referrer@example.test", "settle-referrer")
	referee := createReferralTestUser(t, db, "settle-referee@example.test", "settle-referee")
	config := referralProgramConfigForServiceTest(true)
	require.NoError(t, db.First(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "SETTLE234", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusVesting,
		RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour), VestingUntil: referralPtrTime(time.Now().UTC().Add(-time.Hour)),
	}
	require.NoError(t, db.Create(record).Error)

	first, err := referralService.SettleReferral(record.ID, "manual verification", &referrer.ID)
	require.NoError(t, err)
	require.NotNil(t, first.Reward)
	assert.Equal(t, loyalty.ReferralStatusSettled, first.Record.Status)

	second, err := referralService.SettleReferral(record.ID, "retry", &referrer.ID)
	require.NoError(t, err)
	assert.Equal(t, loyalty.ReferralStatusSettled, second.Record.Status)
	var rewardCount, transactionCount int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).Where("referral_record_id = ?", record.ID).Count(&rewardCount).Error)
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", "referral", record.ID).Count(&transactionCount).Error)
	assert.Equal(t, int64(1), rewardCount)
	assert.Equal(t, int64(1), transactionCount)
}

func TestReferralOrderPaidDoesNotGrantRegistrationPoints(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "paid-points-referrer@example.test", "paid-points-referrer")
	referee := createReferralTestUser(t, db, "paid-points-referee@example.test", "paid-points-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitPoints
	config.RefereeBenefitValue = 75
	require.NoError(t, db.Save(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "PAIDPTS2", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusPending,
		RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(record).Error)
	paidAt := time.Now().UTC()
	orderRecord := &orderdomain.Order{OrderNumber: "REF-PAID-POINTS", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	event := outboxdomain.Event{EventKey: "referral.order_paid:points-test", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload}
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))

	// Registration points are issued only by BindFromToken in the account
	// binding transaction. A payment event must never create a missing reward
	// row or credit points as a side effect (including when retried).
	var rewardCount, transactionCount int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).
		Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).
		Count(&rewardCount).Error)
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).
		Where("source = ? AND source_id = ?", "referral_referee", record.ID).
		Count(&transactionCount).Error)
	assert.Equal(t, int64(0), rewardCount)
	assert.Equal(t, int64(0), transactionCount)
	var balance loyalty.UserLoyalty
	assert.ErrorIs(t, db.Where("user_id = ?", referee.ID).First(&balance).Error, gorm.ErrRecordNotFound)

}

func TestReferralOrderInvalidatedDoesNotChangeRegistrationPoints(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "invalidated-points-referrer@example.test", "invalidated-points-referrer")
	referee := createReferralTestUser(t, db, "invalidated-points-referee@example.test", "invalidated-points-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitPoints
	config.RefereeBenefitValue = 75
	require.NoError(t, db.Save(config).Error)
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromToken(referee.ID, token, "203.0.113.41")
	require.NoError(t, err)
	paidAt := time.Now().UTC()
	orderRecord := &orderdomain.Order{OrderNumber: "REF-INVALIDATED-POINTS", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	paidPayload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{
		EventKey: "referral.order_paid:invalidated-points", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: paidPayload,
	}))

	// The referee spends aggregate account points on this order. The payment
	// refund flow returns the order's 50 points before the referral invalidation
	// event is handled. The registration reward remains in that same aggregate
	// balance; the source labels do not create separate wallets.
	orderRecord.PointsUsed = 50
	require.NoError(t, db.Save(orderRecord).Error)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	_, err = loyaltyRepo.AdjustUserPoints(referee.ID, -50, "spend", "order", orderRecord.ID, "Spent referral points on order")
	require.NoError(t, err)
	_, err = loyaltyRepo.AdjustUserPoints(referee.ID, 50, "refund", "refund_loyalty_points_return", 9001, "Returned points used on refunded order")
	require.NoError(t, err)

	invalidatedAt := paidAt.Add(time.Hour)
	invalidatedPayload, err := json.Marshal(outboxdomain.ReferralOrderInvalidatedPayload{OrderID: orderRecord.ID, OccurredAt: invalidatedAt, Reason: "full refund", Source: "refund_webhook", Reference: "refund-points-1"})
	require.NoError(t, err)
	event := outboxdomain.Event{EventKey: "referral.order_invalidated:invalidated-points", EventType: outboxdomain.EventTypeReferralOrderInvalidated, Payload: invalidatedPayload}
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), event))
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), event))

	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusReleased, reward.Status)
	var balance loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", referee.ID).First(&balance).Error)
	assert.Equal(t, 75, balance.AvailablePoints)
}

func TestReferralBindingRejectsReferrerEmailAndStrictIPVelocity(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "strict-email@example.test", "strict-email-referrer")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.AntiFraudMode = loyalty.ReferralFraudModeStrict
	require.NoError(t, db.Save(config).Error)

	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	emailMatch := createReferralTestUser(t, db, "STRICT-EMAIL@example.test", "strict-email-referee")
	_, err = referralService.BindFromTokenWithContext(emailMatch.ID, token, ReferralBindContext{RefereeEmail: emailMatch.Email, ClientIP: "203.0.113.99"})
	assert.ErrorIs(t, err, ErrSelfReferralForbidden)

	for index := 0; index < 4; index++ {
		referee := createReferralTestUser(t, db, fmt.Sprintf("strict-ip-%d@example.test", index), fmt.Sprintf("strict-ip-%d", index))
		freshToken, _, tokenErr := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
		require.NoError(t, tokenErr)
		_, bindErr := referralService.BindFromTokenWithContext(referee.ID, freshToken, ReferralBindContext{
			RefereeEmail: referee.Email,
			ClientIP:     "203.0.113.99",
		})
		if index < 3 {
			require.NoError(t, bindErr)
		} else {
			assert.ErrorIs(t, bindErr, ErrRefereeNotEligible)
		}
	}
}

func TestReferralBindingRejectsStrictDeviceCollisionWithReferrerHistory(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	historicalReferrer := createReferralTestUser(t, db, "historical-referrer@example.test", "historical-referrer")
	referrer := createReferralTestUser(t, db, "device-referrer@example.test", "device-referrer")
	referee := createReferralTestUser(t, db, "device-referee@example.test", "device-referee")

	historicalDashboard, err := referralService.Dashboard(historicalReferrer.ID)
	require.NoError(t, err)
	historicalToken, _, err := referralService.CreateAttributionToken(historicalDashboard.ReferralCode, "link")
	require.NoError(t, err)
	_, err = referralService.BindFromTokenWithContext(referrer.ID, historicalToken, ReferralBindContext{
		RefereeEmail:      referee.Email,
		DeviceFingerprint: "shared-device",
		ClientIP:          "198.51.100.10",
	})
	require.NoError(t, err)

	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.AntiFraudMode = loyalty.ReferralFraudModeStrict
	require.NoError(t, db.Save(config).Error)
	referrerDashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	mainToken, _, err := referralService.CreateAttributionToken(referrerDashboard.ReferralCode, "link")
	require.NoError(t, err)
	newReferee := createReferralTestUser(t, db, "device-new-referee@example.test", "device-new-referee")
	_, err = referralService.BindFromTokenWithContext(newReferee.ID, mainToken, ReferralBindContext{
		RefereeEmail:      newReferee.Email,
		DeviceFingerprint: "shared-device",
		ClientIP:          "198.51.100.11",
	})
	assert.ErrorIs(t, err, ErrRefereeNotEligible)
}

func TestReferralPaymentFingerprintCollisionRevokesInStrictMode(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "payment-referrer@example.test", "payment-referrer")
	referee := createReferralTestUser(t, db, "payment-referee@example.test", "payment-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.AntiFraudMode = loyalty.ReferralFraudModeStrict
	require.NoError(t, db.Save(config).Error)

	paidAt := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	referrerOrder := &orderdomain.Order{
		OrderNumber: "REF-FP-REFERRER", UserID: referrer.ID, Status: "completed", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt,
	}
	require.NoError(t, db.Create(referrerOrder).Error)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID: referrerOrder.ID, TransactionID: "txn_referrer_fp", PaymentMethod: "stripe",
		AmountMinor: 25000, Currency: "USD", Status: "completed",
		GatewayResponse: `{"payment_method_details":{"card":{"fingerprint":"fp-shared"}}}`,
		CompletedAt:     &paidAt,
	}).Error)

	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromTokenWithContext(referee.ID, token, ReferralBindContext{RefereeEmail: referee.Email, ClientIP: "192.0.2.10"})
	require.NoError(t, err)

	refereeOrder := &orderdomain.Order{
		OrderNumber: "REF-FP-REFEREE", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt,
	}
	require.NoError(t, db.Create(refereeOrder).Error)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID: refereeOrder.ID, TransactionID: "txn_referee_fp", PaymentMethod: "stripe",
		AmountMinor: 25000, Currency: "USD", Status: "completed",
		GatewayResponse: `{"payment_method_details":{"card":{"fingerprint":"fp-shared"}}}`,
		CompletedAt:     &paidAt,
	}).Error)

	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{
		OrderID: refereeOrder.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt,
	})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{
		EventKey: "referral.order_paid:payment-fingerprint", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload,
	}))

	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusRevoked, saved.Status)
	assert.Contains(t, string(saved.RiskFlags), "payment_fingerprint_match")
	assert.NotEmpty(t, saved.PaymentFingerprintHash)
}

func TestRevokeReferralForfeitsLockedReward(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "revoke-referrer@example.test", "revoke-referrer")
	referee := createReferralTestUser(t, db, "revoke-referee@example.test", "revoke-referee")
	config := referralProgramConfigForServiceTest(true)
	require.NoError(t, db.First(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "REVOKE234", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusVesting,
		RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour), VestingUntil: referralPtrTime(time.Now().UTC().Add(time.Hour)),
	}
	require.NoError(t, db.Create(record).Error)
	reward := &loyalty.ReferralReward{ReferralRecordID: record.ID, ProgramConfigID: config.ID, RecipientUserID: referrer.ID, RecipientRole: loyalty.ReferralRecipientReferrer, RewardType: loyalty.ReferralRewardTypePoints, PointsAmount: config.ReferrerRewardPoints, IdempotencyKey: fmt.Sprintf("referral:%d:referrer:points:v1", record.ID), Status: loyalty.ReferralRewardStatusLocked, RuleSnapshot: []byte(`{"version":1}`)}
	require.NoError(t, db.Create(reward).Error)

	result, err := referralService.RevokeReferral(record.ID, "address collision", &referrer.ID)
	require.NoError(t, err)
	assert.Equal(t, loyalty.ReferralStatusRevoked, result.Record.Status)
	var stored loyalty.ReferralReward
	require.NoError(t, db.First(&stored, reward.ID).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusForfeited, stored.Status)
}

func TestSettledReferralRefundReversesRewardIdempotently(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "reverse-referrer@example.test", "reverse-referrer")
	referee := createReferralTestUser(t, db, "reverse-referee@example.test", "reverse-referee")
	config := referralProgramConfigForServiceTest(true)
	require.NoError(t, db.First(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "REVERSE234", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	paidAt := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	orderRecord := &orderdomain.Order{OrderNumber: "REF-REVERSE-1", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmountMinor: 25000, TotalAmountMinor: 25000, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	record := &loyalty.ReferralRecord{ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID, ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", OrderID: &orderRecord.ID, Currency: "USD", OrderAmountMinor: 25000, Status: loyalty.ReferralStatusVesting, RecordVersion: 1, ExpiresAt: paidAt.AddDate(0, 0, 30), VestingUntil: referralPtrTime(paidAt.AddDate(0, 0, 30))}
	require.NoError(t, db.Create(record).Error)
	_, err := referralService.SettleReferral(record.ID, "vesting complete", nil)
	require.NoError(t, err)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	_, err = loyaltyRepo.AdjustUserPoints(referrer.ID, -config.ReferrerRewardPoints, "spend", "redemption", record.ID, "Redeemed all referral points")
	require.NoError(t, err)
	invalidatedAt := paidAt.AddDate(0, 0, 35)
	payload, err := json.Marshal(outboxdomain.ReferralOrderInvalidatedPayload{OrderID: orderRecord.ID, OccurredAt: invalidatedAt, Reason: "full refund", Source: "refund_webhook", Reference: "refund-reverse-1"})
	require.NoError(t, err)
	event := outboxdomain.Event{EventKey: "referral.reverse:test-1", EventType: outboxdomain.EventTypeReferralOrderInvalidated, Payload: payload}
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), event))
	require.NoError(t, referralService.HandleOrderInvalidatedOutbox(context.Background(), event))
	var saved loyalty.ReferralRecord
	require.NoError(t, db.First(&saved, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusReversed, saved.Status)
	var referrerBalance loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", referrer.ID).First(&referrerBalance).Error)
	assert.Zero(t, referrerBalance.AvailablePoints)
	assert.Equal(t, config.ReferrerRewardPoints, referrerBalance.DebtPoints)
	_, err = loyaltyRepo.AdjustUserPoints(referrer.ID, -1, "spend", "redemption", record.ID+1, "Attempted redemption while in debt")
	assert.ErrorIs(t, err, repository.ErrInsufficientPoints)
	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ?", record.ID).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusReversed, reward.Status)
	var reversalCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", "referral_reversal", record.ID).Count(&reversalCount).Error)
	assert.Equal(t, int64(1), reversalCount)
	var reversal loyalty.LoyaltyTransaction
	require.NoError(t, db.Where("source = ? AND source_id = ?", "referral_reversal", record.ID).First(&reversal).Error)
	assert.Equal(t, config.ReferrerRewardPoints, reversal.DebtBalance)

	partialRepayment := config.ReferrerRewardPoints / 2
	_, err = loyaltyRepo.AdjustUserPoints(referrer.ID, partialRepayment, "earn", "order", orderRecord.ID+10, "Future points repay referral debt")
	require.NoError(t, err)
	require.NoError(t, db.Where("user_id = ?", referrer.ID).First(&referrerBalance).Error)
	assert.Equal(t, config.ReferrerRewardPoints-partialRepayment, referrerBalance.DebtPoints)
	assert.Zero(t, referrerBalance.AvailablePoints)
	_, err = loyaltyRepo.AdjustUserPoints(referrer.ID, partialRepayment+25, "earn", "order", orderRecord.ID+11, "Future points finish repaying referral debt")
	require.NoError(t, err)
	require.NoError(t, db.Where("user_id = ?", referrer.ID).First(&referrerBalance).Error)
	assert.Zero(t, referrerBalance.DebtPoints)
	assert.Equal(t, 25, referrerBalance.AvailablePoints)
}

func referralProgramConfigForServiceTest(enabled bool) *loyalty.ReferralProgramConfig {
	return &loyalty.ReferralProgramConfig{
		Version:                 1,
		Status:                  "active",
		Enabled:                 enabled,
		Currency:                "USD",
		MinOrderAmountMinor:     20000,
		ReferrerRewardPoints:    1000,
		RefereeBenefitType:      loyalty.ReferralBenefitPoints,
		RefereeBenefitValue:     50,
		VestingPeriodDays:       30,
		UndeliveredFallbackDays: 45,
		AttributionTTLDays:      30,
		MonthlyCapPerReferrer:   10,
		AntiFraudMode:           loyalty.ReferralFraudModeMonitor,
	}
}

func createReferralTestUser(t *testing.T, db *gorm.DB, email, username string) *user.User {
	t.Helper()
	account := &user.User{Email: email, Username: username, Password: "not-used", Status: "active", Role: "user"}
	require.NoError(t, db.Create(account).Error)
	return account
}
