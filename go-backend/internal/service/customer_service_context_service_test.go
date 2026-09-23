package service

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerCartItemsUsesCurrencyMinorUnits(t *testing.T) {
	items, err := customerCartItems([]product.CartItem{{
		ID:         7,
		Quantity:   2,
		PriceMinor: 10000,
		Currency:   "JPY",
	}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "JPY", items[0].Currency)
	assert.Equal(t, "20000", items[0].LineTotal)
}

func TestCustomerCartItemsRejectsInvalidMoneyInsteadOfReturningZero(t *testing.T) {
	items, err := customerCartItems([]product.CartItem{{
		ID:         8,
		Quantity:   1,
		PriceMinor: 1000,
		Currency:   "XXX",
	}})
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "price")
}

func TestCustomerBrowsingItemsEnrichesPublicProductProjection(t *testing.T) {
	lastViewedAt := time.Date(2026, 9, 18, 8, 0, 0, 0, time.UTC)
	items := customerBrowsingItems([]user.BrowsingHistory{{
		ProductID:    7,
		ViewCount:    3,
		LastViewedAt: lastViewedAt,
	}}, []product.Product{{
		ID:         7,
		Name:       "Road Bike",
		SKU:        "BIKE-LEGACY",
		Currency:   "USD",
		PriceMinor: 12999,
		Media: []product.ProductMedia{{
			MediaType: "image",
			URL:       "https://cdn.example.test/bike.webp",
			IsVisible: true,
		}},
	}})

	require.Len(t, items, 1)
	assert.Equal(t, "Road Bike", items[0].Name)
	assert.Equal(t, "BIKE-LEGACY", items[0].SKU)
	assert.Equal(t, "https://cdn.example.test/bike.webp", items[0].Thumbnail)
	assert.Equal(t, "USD", items[0].Currency)
	assert.Equal(t, "129.99", items[0].Price)
	assert.Equal(t, 3, items[0].ViewCount)
}

func TestCustomerServiceReplyMetricsPairsLatestCustomerTurn(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(1 * time.Minute)},
		{CreatedAt: start.Add(2 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(5 * time.Minute)},
		{CreatedAt: start.Add(10 * time.Minute)},
		{IsInternal: true, CreatedAt: start.Add(11 * time.Minute)},
		{IsStaff: true, CreatedAt: end},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 180.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 1, unansweredTurns)
}

func TestCustomerServiceReplyMetricsIncludesPendingTurnBeforeWindow(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	messages := []ticket.TicketMessage{
		{CreatedAt: start.Add(-30 * time.Minute)},
		{IsStaff: true, CreatedAt: start.Add(30 * time.Minute)},
	}

	totalSeconds, replyCount, unansweredTurns := customerServiceReplyMetrics(messages, start, end)

	assert.Equal(t, 3600.0, totalSeconds)
	assert.Equal(t, 1, replyCount)
	assert.Equal(t, 0, unansweredTurns)
}

func TestCustomerServiceContextProductImageUsesCanonicalPublicURL(t *testing.T) {
	resolver := NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := &product.Product{
		Media: []product.ProductMedia{{
			MediaType: "image",
			IsVisible: true,
			URL:       "http://media.internal:8080/uploads/products/wheel.jpg",
		}},
	}

	require.Equal(t, "https://shop.example.test/uploads/products/wheel.jpg", firstProductImage(item, resolver))
	require.Equal(t, "http://media.internal:8080/uploads/products/wheel.jpg", item.Media[0].URL)
}

func TestCustomerServiceContextAppliesVisitorTimezoneBeforeAccountTimezone(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	userID := uint(42)
	require.NoError(t, db.Create(&visitor.Profile{
		UserID:     &userID,
		Timezone:   "Europe/Berlin",
		LastSeenAt: time.Now().UTC(),
	}).Error)
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "visitor-hash",
		UserID:                     &userID,
		Timezone:                   "America/Los_Angeles",
		LastSeenAt:                 time.Now().UTC(),
	}).Error)

	contextService := &CustomerServiceContextService{
		visitorProfileService: NewVisitorProfileService(repository.NewVisitorProfileRepository(db)),
	}
	context := &CustomerServiceContext{
		Contact: CustomerServiceContextContact{TimezoneSource: "not_captured"},
	}

	contextService.applyVisitorTimezone(context, "visitor-hash", userID)

	assert.Equal(t, "America/Los_Angeles", context.Contact.Timezone)
	assert.Equal(t, "visitor_profile", context.Contact.TimezoneSource)
}

func TestCustomerServiceContextUsesFreshestDuplicateVisitorTimezone(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	oldSeen := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	newSeen := oldSeen.Add(24 * time.Hour)
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "duplicate-visitor-hash",
		Timezone:                   "",
		LastSeenAt:                 oldSeen,
		UpdatedAt:                  oldSeen,
	}).Error)
	require.NoError(t, db.Create(&visitor.Profile{
		CustomerServiceVisitorHash: "duplicate-visitor-hash",
		Timezone:                   "Asia/Tokyo",
		LastSeenAt:                 newSeen,
		UpdatedAt:                  newSeen,
	}).Error)

	contextService := &CustomerServiceContextService{
		visitorProfileService: NewVisitorProfileService(repository.NewVisitorProfileRepository(db)),
	}
	context := &CustomerServiceContext{
		Contact: CustomerServiceContextContact{TimezoneSource: "not_captured"},
	}

	contextService.applyVisitorTimezone(context, "duplicate-visitor-hash", 0)

	assert.Equal(t, "Asia/Tokyo", context.Contact.Timezone)
	assert.Equal(t, "visitor_profile", context.Contact.TimezoneSource)
}

func TestCustomerServiceContextFallsBackToAccountTimezoneWhenVisitorProfileMissing(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&visitor.Profile{}))

	userID := uint(84)
	require.NoError(t, db.Create(&visitor.Profile{
		UserID:     &userID,
		Timezone:   "Asia/Tokyo",
		LastSeenAt: time.Now().UTC(),
	}).Error)

	contextService := &CustomerServiceContextService{
		visitorProfileService: NewVisitorProfileService(repository.NewVisitorProfileRepository(db)),
	}
	context := &CustomerServiceContext{
		Contact: CustomerServiceContextContact{TimezoneSource: "not_captured"},
	}

	contextService.applyVisitorTimezone(context, "missing-visitor-profile", userID)

	assert.Equal(t, "Asia/Tokyo", context.Contact.Timezone)
	assert.Equal(t, "visitor_profile", context.Contact.TimezoneSource)
}

func TestCustomerShippingAddressContextIsMinimized(t *testing.T) {
	result := customerShippingAddressContext([]order.Order{{
		ID:          9,
		OrderNumber: "ORD-9",
		ShippingAddress: order.Address{
			FirstName: "Ada", LastName: "Lovelace", Address1: "1 Analytical Engine Way",
			Address2: "Unit 2", City: "London", State: "LDN", PostalCode: "NW1",
			Country: "GB", Phone: "+44 1234 567890", Email: "ada@example.test",
		},
	}})

	require.True(t, result.Available)
	assert.Equal(t, "Ada Lovelace", result.RecipientName)
	assert.Equal(t, "1 Analytical Engine Way Unit 2", result.AddressLine)
	assert.True(t, result.PhonePresent)
	payload, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "+44 1234 567890")
	assert.NotContains(t, string(payload), "ada@example.test")
}

func TestCustomerOperationalFactsContextProjectsRMAAndDisputeSummary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&order.Order{}, &order.OrderItem{},
		&aftersales.AfterSalesCase{}, &aftersales.AfterSalesCaseItem{},
		&payment.StripeDispute{}, &payment.PayPalDispute{},
	))

	userID := uint(73)
	orderRecord := order.Order{
		OrderNumber: "ORD-73", UserID: userID, Status: "paid", PaymentStatus: "paid",
		ShippingStatus: "processing", TotalAmountMinor: 12950, Currency: "USD",
		ShippingAddress: order.Address{FirstName: "Grace", LastName: "Hopper", Address1: "1 Navy St", City: "Arlington", Country: "US"},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	require.NoError(t, db.Create(&aftersales.AfterSalesCase{
		OrderID: orderRecord.ID, Type: aftersales.TypeRefundOnly, Status: aftersales.StatusReviewing,
		Reason: "Damaged item", Items: []aftersales.AfterSalesCaseItem{{OrderID: orderRecord.ID, ProductID: 5, ProductName: "Compiler", Quantity: 1}},
	}).Error)
	evidenceDue := time.Now().UTC().Add(24 * time.Hour)
	require.NoError(t, db.Create(&payment.StripeDispute{
		StripeDisputeID: "dp_secret", StripeChargeID: "ch_secret", PaymentIntentID: "pi_secret",
		OrderID: &orderRecord.ID, AmountMinor: 12950, Currency: "USD", Reason: "fraudulent", Status: "needs_response", EvidenceDueAt: &evidenceDue,
	}).Error)

	service := &CustomerServiceContextService{
		orderRepo:      repository.NewOrderRepository(db),
		afterSalesRepo: repository.NewAfterSalesCaseRepository(db),
		paymentRepo:    repository.NewPaymentRepository(db),
	}
	orders, address, rma, disputes := service.customerOperationalFactsContext(userID)
	require.True(t, orders.Available)
	require.True(t, address.Available)
	require.True(t, rma.Available)
	require.Len(t, rma.Items, 1)
	require.True(t, disputes.Available)
	require.Len(t, disputes.Items, 1)

	payload, err := json.Marshal(disputes)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "dp_secret")
	assert.NotContains(t, string(payload), "ch_secret")
	assert.NotContains(t, string(payload), "pi_secret")
}

func TestCustomerOperationalFactsContextKeepsBlockErrorsIsolated(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}, &order.OrderItem{}))
	userID := uint(91)
	require.NoError(t, db.Create(&order.Order{OrderNumber: "ORD-91", UserID: userID, Status: "paid", PaymentStatus: "paid", Currency: "USD"}).Error)

	service := &CustomerServiceContextService{
		orderRepo:      repository.NewOrderRepository(db),
		afterSalesRepo: repository.NewAfterSalesCaseRepository(db),
		paymentRepo:    repository.NewPaymentRepository(db),
	}
	orders, _, rma, disputes := service.customerOperationalFactsContext(userID)
	assert.Equal(t, "available", orders.Status)
	assert.Equal(t, "error", rma.Status)
	assert.Equal(t, "error", disputes.Status)
	assert.NotEmpty(t, rma.Reason)
	assert.NotEmpty(t, disputes.Reason)
}
