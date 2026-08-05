package order

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tanzanite/internal/domain/audit"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateOrderRejectsPausedProviderBeforeCartLookup(t *testing.T) {
	db := newOrderPaymentProtectionTestDB(t)
	protection := newOrderPaymentProtectionService(db)
	_, err := protection.CreateControl(service.CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "provider",
		ScopeValue: "stripe",
		Reason:     "temporary card acquiring incident",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString(validCreateOrderBody("card", "US")))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("user_id", uint(42))

	handler := NewHandler(nil, nil)
	handler.ConfigurePaymentProtection(protection)
	handler.CreateOrder(context)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.JSONEq(t, `{"code":"payment_paused","message":"Payment is temporarily unavailable for this order","details":{"action":"pause_payment"}}`, recorder.Body.String())
}

func TestAuthorizeOrderPaymentStartUsesShippingCountry(t *testing.T) {
	db := newOrderPaymentProtectionTestDB(t)
	protection := newOrderPaymentProtectionService(db)
	_, err := protection.CreateControl(service.CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "country",
		ScopeValue: "US",
		Reason:     "temporary regional acquiring incident",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)

	handler := NewHandler(nil, nil)
	handler.ConfigurePaymentProtection(protection)
	allowed := handler.authorizeOrderPaymentStart(context, CreateOrderRequest{
		PaymentMethod: "paypal",
		ShippingAddress: AddressRequest{
			Country: "US",
		},
	})

	require.False(t, allowed)
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAuthorizeOrderPaymentStartAllowsNonMatchingProvider(t *testing.T) {
	db := newOrderPaymentProtectionTestDB(t)
	protection := newOrderPaymentProtectionService(db)
	_, err := protection.CreateControl(service.CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "provider",
		ScopeValue: "paypal",
		Reason:     "temporary paypal incident",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)

	handler := NewHandler(nil, nil)
	handler.ConfigurePaymentProtection(protection)
	allowed := handler.authorizeOrderPaymentStart(context, CreateOrderRequest{
		PaymentMethod: "card",
		ShippingAddress: AddressRequest{
			Country: "US",
		},
	})

	require.True(t, allowed)
	require.Empty(t, recorder.Body.String())
}

func validCreateOrderBody(paymentMethod string, country string) string {
	return `{
		"items": [{"product_id": 1, "quantity": 1}],
		"shipping_address": {
			"first_name": "Test",
			"last_name": "Buyer",
			"address1": "1 Test Street",
			"city": "Austin",
			"postal_code": "78701",
			"country": "` + country + `",
			"phone": "+15555550123",
			"email": "buyer@example.com"
		},
		"billing_address": {
			"first_name": "Test",
			"last_name": "Buyer",
			"address1": "1 Test Street",
			"city": "Austin",
			"postal_code": "78701",
			"country": "` + country + `",
			"phone": "+15555550123",
			"email": "buyer@example.com"
		},
		"payment_method": "` + paymentMethod + `",
		"shipping_method": "standard"
	}`
}

func newOrderPaymentProtectionService(db *gorm.DB) *service.PaymentProtectionService {
	return service.NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
}

func newOrderPaymentProtectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
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
		&paymentdomain.PaymentProtectionControl{},
		&audit.AuditLog{},
	))
	return db
}
