package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	coupondomain "commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	orderdomain "commerce-platform/internal/domain/order"
	outboxdomain "commerce-platform/internal/domain/outbox"
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

func TestReferralBindingProvisionsLockedRefereeCoupon(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "bind-coupon-referrer@example.test", "bind-coupon-referrer")
	referee := createReferralTestUser(t, db, "bind-coupon-referee@example.test", "bind-coupon-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitFixedCoupon
	config.RefereeBenefitValue = 3000
	require.NoError(t, db.Save(config).Error)
	dashboard, err := referralService.Dashboard(referrer.ID)
	require.NoError(t, err)
	token, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	record, err := referralService.BindFromToken(referee.ID, token, "203.0.113.13")
	require.NoError(t, err)

	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusLocked, reward.Status)
	assert.Equal(t, loyalty.ReferralRewardTypeCoupon, reward.RewardType)
	require.NotNil(t, reward.CouponID)
	var issued coupondomain.Coupon
	require.NoError(t, db.First(&issued, *reward.CouponID).Error)
	assert.Equal(t, "fixed", issued.Type)
	assert.InDelta(t, 30, issued.Value, 0.001)
	checkout := &CheckoutService{}
	autoCode, err := checkout.lockedRefereeCouponCode(checkoutRepositories{
		couponRepo:          repository.NewCouponRepository(db),
		referralRepo:        repository.NewReferralRepository(db),
		referralProgramRepo: repository.NewReferralProgramRepository(db),
	}, referee.ID)
	require.NoError(t, err)
	assert.Equal(t, issued.Code, autoCode)
}

func TestPublishAdminReferralProgramConfigCreatesNewVersion(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, false)
	operator := createReferralTestUser(t, db, "operator@example.test", "operator")

	config, err := referralService.PublishAdminProgramConfig(ReferralProgramConfigInput{
		Enabled:                 true,
		Currency:                "usd",
		MinOrderAmountMinor:     25000,
		ReferrerRewardPoints:    1250,
		RefereeBenefitType:      loyalty.ReferralBenefitPercentCoupon,
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
		Currency:                "USD",
		ReferrerRewardPoints:    1000,
		RefereeBenefitType:      loyalty.ReferralBenefitNone,
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
		OrderNumber:    "REF-PAID-1",
		UserID:         formerCustomer.ID,
		Status:         "refunded",
		PaymentStatus:  "refunded",
		SubtotalAmount: 250,
		TotalAmount:    250,
		Currency:       "USD",
		PaidAt:         &paidAt,
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
		OrderNumber:    "REF-LIFECYCLE-1",
		UserID:         referee.ID,
		Status:         "processing",
		PaymentStatus:  "paid",
		SubtotalAmount: 250,
		TotalAmount:    250,
		Currency:       "USD",
		PaidAt:         &paidAt,
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
	assert.Zero(t, rewardCount)
}

func TestDeliveredTrackingUpdatePersistsTimestampAndReferralEventOnce(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	account := createReferralTestUser(t, db, "delivery@example.test", "delivery")
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-DELIVERY-1", UserID: account.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", ShippingStatus: "shipped",
	}
	require.NoError(t, db.Create(orderRecord).Error)

	shippingService := NewShippingService(nil)
	shippingService.ConfigureOrderRepository(repository.NewOrderRepository(db))
	shippingService.ConfigureTxManager(referralService.txManager)
	deliveredAt := time.Date(2026, 9, 20, 9, 30, 0, 0, time.UTC)
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

func TestReferralLifecycleScanExpiresPendingAndAdvancesDelivery(t *testing.T) {
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt,
		ShippingStatus: "delivered", DeliveredAt: referralPtrTime(paidAt.AddDate(0, 0, 2)),
	}
	require.NoError(t, db.Create(orderRecord).Error)
	require.NoError(t, db.Model(&loyalty.ReferralRecord{}).Where("id = ?", pending.ID).Updates(map[string]any{
		"order_id": orderRecord.ID, "status": loyalty.ReferralStatusOrdered, "ordered_at": paidAt,
	}).Error)

	expiredReferee := createReferralTestUser(t, db, "scan-expired@example.test", "scan-expired")
	expiredToken, _, err := referralService.CreateAttributionToken(dashboard.ReferralCode, "link")
	require.NoError(t, err)
	expired, err := referralService.BindFromToken(expiredReferee.ID, expiredToken, "203.0.113.31")
	require.NoError(t, err)
	require.NoError(t, db.Model(&loyalty.ReferralRecord{}).Where("id = ?", expired.ID).Update("expires_at", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)).Error)

	result, err := referralService.ScanLifecycle(context.Background(), time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), 100)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Expired)
	assert.Equal(t, 1, result.AdvancedToVesting)

	var savedPending, savedExpired loyalty.ReferralRecord
	require.NoError(t, db.First(&savedPending, pending.ID).Error)
	require.NoError(t, db.First(&savedExpired, expired.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusSettled, savedPending.Status)
	assert.Equal(t, orderRecord.DeliveredAt.UTC(), savedPending.DeliveredAt.UTC())
	assert.Equal(t, orderRecord.DeliveredAt.UTC().AddDate(0, 0, 30), savedPending.VestingUntil.UTC())
	assert.Equal(t, loyalty.ReferralStatusExpired, savedExpired.Status)

	var rewards int64
	require.NoError(t, db.Model(&loyalty.ReferralReward{}).Count(&rewards).Error)
	assert.Equal(t, int64(1), rewards)
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: referralPtrTime(shippedAt.Add(-24 * time.Hour)),
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &priorPaidAt, ShippingAddress: sharedAddress,
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt, ShippingAddress: sharedAddress,
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &priorPaidAt, ShippingAddress: sharedAddress,
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
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt, ShippingAddress: sharedAddress,
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
	assert.Equal(t, int64(0), rewardCount)
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
	orderRecord := &orderdomain.Order{OrderNumber: "REF-CAP-SECOND", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt}
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

func TestReferralOrderPaidReleasesRefereePointsIdempotently(t *testing.T) {
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
	orderRecord := &orderdomain.Order{OrderNumber: "REF-PAID-POINTS", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	event := outboxdomain.Event{EventKey: "referral.order_paid:points-test", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload}
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))

	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardTypePoints, reward.RewardType)
	assert.Equal(t, 75, reward.PointsAmount)
	assert.Equal(t, loyalty.ReferralRewardStatusReleased, reward.Status)
	var transactionCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", "referral_referee", record.ID).Count(&transactionCount).Error)
	assert.Equal(t, int64(1), transactionCount)
	var balance loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", referee.ID).First(&balance).Error)
	assert.Equal(t, 75, balance.AvailablePoints)
}

func TestReferralOrderInvalidatedReversesReleasedRefereePoints(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "invalidated-points-referrer@example.test", "invalidated-points-referrer")
	referee := createReferralTestUser(t, db, "invalidated-points-referee@example.test", "invalidated-points-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitPoints
	config.RefereeBenefitValue = 75
	require.NoError(t, db.Save(config).Error)
	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "INVPOINT234", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusPending,
		RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(record).Error)
	paidAt := time.Now().UTC()
	orderRecord := &orderdomain.Order{OrderNumber: "REF-INVALIDATED-POINTS", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	paidPayload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), outboxdomain.Event{
		EventKey: "referral.order_paid:invalidated-points", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: paidPayload,
	}))

	// The referee spends part of the welcome reward on this order. The payment
	// refund flow returns the order's 50 points before the referral invalidation
	// event is handled, so the two independent ledgers can both settle without
	// creating a negative balance.
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
	assert.Equal(t, loyalty.ReferralRewardStatusReversed, reward.Status)
	var balance loyalty.UserLoyalty
	require.NoError(t, db.Where("user_id = ?", referee.ID).First(&balance).Error)
	assert.Equal(t, 0, balance.AvailablePoints)
	var reversalCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", referralRefereeReversalSource, record.ID).Count(&reversalCount).Error)
	assert.Equal(t, int64(1), reversalCount)
}

func TestReferralOrderPaidCreatesRefereeCouponBenefits(t *testing.T) {
	tests := []struct {
		name        string
		benefit     string
		value       int64
		couponType  string
		couponValue float64
	}{
		{name: "percent", benefit: loyalty.ReferralBenefitPercentCoupon, value: 500, couponType: "percentage", couponValue: 5},
		{name: "fixed", benefit: loyalty.ReferralBenefitFixedCoupon, value: 3000, couponType: "fixed", couponValue: 30},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			referralService, db := newReferralServiceFixture(t, true)
			referrer := createReferralTestUser(t, db, "paid-coupon-referrer-"+test.name+"@example.test", "paid-coupon-referrer-"+test.name)
			referee := createReferralTestUser(t, db, "paid-coupon-referee-"+test.name+"@example.test", "paid-coupon-referee-"+test.name)
			config := &loyalty.ReferralProgramConfig{}
			require.NoError(t, db.First(config).Error)
			config.RefereeBenefitType = test.benefit
			config.RefereeBenefitValue = test.value
			require.NoError(t, db.Save(config).Error)
			identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "PAID" + strings.ToUpper(test.name[:1]) + "234", IsActive: true}
			require.NoError(t, db.Create(identity).Error)
			record := &loyalty.ReferralRecord{
				ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
				ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusPending,
				RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
			}
			require.NoError(t, db.Create(record).Error)
			paidAt := time.Now().UTC()
			orderRecord := &orderdomain.Order{OrderNumber: "REF-PAID-COUPON-" + test.name, UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt}
			require.NoError(t, db.Create(orderRecord).Error)
			payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
			require.NoError(t, err)
			event := outboxdomain.Event{EventKey: "referral.order_paid:coupon-" + test.name, EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload}
			require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))
			require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))

			var reward loyalty.ReferralReward
			require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&reward).Error)
			assert.Equal(t, loyalty.ReferralRewardTypeCoupon, reward.RewardType)
			assert.Equal(t, loyalty.ReferralRewardStatusReleased, reward.Status)
			require.NotNil(t, reward.CouponID)
			var issued coupondomain.Coupon
			require.NoError(t, db.First(&issued, *reward.CouponID).Error)
			assert.Equal(t, test.couponType, issued.Type)
			assert.InDelta(t, test.couponValue, issued.Value, 0.001)
			assert.Equal(t, 1, issued.UsageLimit)
		})
	}
}

func TestReferralOrderPaidForfeitsUnusedRefereeCoupon(t *testing.T) {
	referralService, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "unused-coupon-referrer@example.test", "unused-coupon-referrer")
	referee := createReferralTestUser(t, db, "unused-coupon-referee@example.test", "unused-coupon-referee")
	config := &loyalty.ReferralProgramConfig{}
	require.NoError(t, db.First(config).Error)
	config.RefereeBenefitType = loyalty.ReferralBenefitFixedCoupon
	config.RefereeBenefitValue = 3000
	require.NoError(t, db.Save(config).Error)

	identity := &loyalty.ReferralIdentity{UserID: referrer.ID, ReferralCode: "UNUSED234", IsActive: true}
	require.NoError(t, db.Create(identity).Error)
	record := &loyalty.ReferralRecord{
		ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID,
		ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", Currency: "USD", Status: loyalty.ReferralStatusPending,
		RecordVersion: 1, ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(record).Error)
	require.NoError(t, referralService.provisionLockedRefereeCouponInTx(repository.TxRepositories{
		Referral: referralService.repo,
		Coupon:   repository.NewCouponRepository(db),
	}, record, config, time.Now().UTC()))

	var lockedReward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ? AND recipient_role = ?", record.ID, loyalty.ReferralRecipientReferee).First(&lockedReward).Error)
	require.NotNil(t, lockedReward.CouponID)
	var referralCoupon coupondomain.Coupon
	require.NoError(t, db.First(&referralCoupon, *lockedReward.CouponID).Error)

	paidAt := time.Now().UTC()
	orderRecord := &orderdomain.Order{
		OrderNumber: "REF-PAID-UNUSED-COUPON", UserID: referee.ID, Status: "processing", PaymentStatus: "paid",
		SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", CouponCode: "PUBLIC-SAVE10", PaidAt: &paidAt,
	}
	require.NoError(t, db.Create(orderRecord).Error)
	payload, err := json.Marshal(outboxdomain.ReferralOrderPaidPayload{OrderID: orderRecord.ID, UserID: referee.ID, AmountMinor: 25000, Currency: "USD", PaidAt: paidAt})
	require.NoError(t, err)
	event := outboxdomain.Event{EventKey: "referral.order_paid:unused-coupon", EventType: outboxdomain.EventTypeReferralOrderPaid, Payload: payload}
	require.NoError(t, referralService.HandleOrderPaidOutbox(context.Background(), event))

	var savedRecord loyalty.ReferralRecord
	require.NoError(t, db.First(&savedRecord, record.ID).Error)
	assert.Equal(t, loyalty.ReferralStatusOrdered, savedRecord.Status)
	var savedReward loyalty.ReferralReward
	require.NoError(t, db.First(&savedReward, lockedReward.ID).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusForfeited, savedReward.Status)
	var savedCoupon coupondomain.Coupon
	require.NoError(t, db.First(&savedCoupon, referralCoupon.ID).Error)
	assert.False(t, savedCoupon.Enabled)
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
	orderRecord := &orderdomain.Order{OrderNumber: "REF-REVERSE-1", UserID: referee.ID, Status: "processing", PaymentStatus: "paid", SubtotalAmount: 250, TotalAmount: 250, Currency: "USD", PaidAt: &paidAt}
	require.NoError(t, db.Create(orderRecord).Error)
	record := &loyalty.ReferralRecord{ReferralIdentityID: identity.ID, ProgramConfigID: config.ID, ReferrerID: referrer.ID, RefereeID: &referee.ID, ReferralCodeSnapshot: identity.ReferralCode, AttributionSource: "link", OrderID: &orderRecord.ID, Currency: "USD", OrderAmountMinor: 25000, Status: loyalty.ReferralStatusVesting, RecordVersion: 1, ExpiresAt: paidAt.AddDate(0, 0, 30), VestingUntil: referralPtrTime(paidAt.AddDate(0, 0, 30))}
	require.NoError(t, db.Create(record).Error)
	_, err := referralService.SettleReferral(record.ID, "vesting complete", nil)
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
	var reward loyalty.ReferralReward
	require.NoError(t, db.Where("referral_record_id = ?", record.ID).First(&reward).Error)
	assert.Equal(t, loyalty.ReferralRewardStatusReversed, reward.Status)
	var reversalCount int64
	require.NoError(t, db.Model(&loyalty.LoyaltyTransaction{}).Where("source = ? AND source_id = ?", "referral_reversal", record.ID).Count(&reversalCount).Error)
	assert.Equal(t, int64(1), reversalCount)
}

func referralProgramConfigForServiceTest(enabled bool) *loyalty.ReferralProgramConfig {
	return &loyalty.ReferralProgramConfig{
		Version:                 1,
		Status:                  "active",
		Enabled:                 enabled,
		Currency:                "USD",
		MinOrderAmountMinor:     20000,
		ReferrerRewardPoints:    1000,
		RefereeBenefitType:      loyalty.ReferralBenefitNone,
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
