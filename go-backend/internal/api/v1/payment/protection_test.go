package payment

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAuthorizePaymentStartBlocksPausedScope(t *testing.T) {
	db := newPaymentProtectionHandlerTestDB(t)
	protection := service.NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
	_, err := protection.CreateControl(service.CreatePaymentProtectionControlInput{
		Action:    "pause_payment",
		ScopeType: "global",
		Reason:    "temporary payment incident",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/stripe/payment-intents", nil)

	handler := &Handler{protection: protection}
	allowed := handler.authorizePaymentStart(context, paymentStartProtectionInput{
		Provider: string(pgateway.GatewayStripe),
		Order: &orderdomain.Order{
			PaymentMethod: "stripe",
			BillingAddress: orderdomain.Address{
				Country: "US",
			},
		},
	})

	require.False(t, allowed)
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.JSONEq(t, `{"code":"payment_paused","message":"Payment is temporarily unavailable for this order","details":{"action":"pause_payment"}}`, recorder.Body.String())
}

func TestAuthorizePaymentStartUsesProviderAsPaymentMethodFallback(t *testing.T) {
	db := newPaymentProtectionHandlerTestDB(t)
	protection := service.NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
	_, err := protection.CreateControl(service.CreatePaymentProtectionControlInput{
		Action:     "pause_payment",
		ScopeType:  "payment_method",
		ScopeValue: "stripe",
		Reason:     "temporary card payment incident",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/stripe/payment-intents", nil)

	handler := &Handler{protection: protection}
	allowed := handler.authorizePaymentStart(context, paymentStartProtectionInput{
		Provider: string(pgateway.GatewayStripe),
		Order: &orderdomain.Order{
			BillingAddress: orderdomain.Address{
				Country: "US",
			},
		},
	})

	require.False(t, allowed)
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAuthorizePaymentStartAllowsNonMatchingProviderScope(t *testing.T) {
	db := newPaymentProtectionHandlerTestDB(t)
	protection := service.NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
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
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/stripe/payment-intents", nil)

	handler := &Handler{protection: protection}
	allowed := handler.authorizePaymentStart(context, paymentStartProtectionInput{
		Provider: string(pgateway.GatewayStripe),
		Order: &orderdomain.Order{
			PaymentMethod: "stripe",
			BillingAddress: orderdomain.Address{
				Country: "US",
			},
		},
	})

	require.True(t, allowed)
	require.Empty(t, recorder.Body.String())
}

func newPaymentProtectionHandlerTestDB(t *testing.T) *gorm.DB {
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
