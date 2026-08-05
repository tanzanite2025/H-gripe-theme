package payment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tanzanite/internal/domain/audit"
	"tanzanite/internal/domain/currency"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestListPaymentMethodsMarksPausedProviderUnavailable(t *testing.T) {
	db := newPaymentMethodAvailabilityTestDB(t)
	require.NoError(t, db.Create(&paymentdomain.PaymentMethod{Name: "Card", Code: "card", Enabled: true}).Error)
	require.NoError(t, db.Create(&paymentdomain.PaymentMethod{Name: "PayPal", Code: "paypal", Enabled: true}).Error)

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
		ScopeValue: "stripe",
		Reason:     "temporary card acquiring incident",
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
	}, service.PaymentProtectionActor{UserID: 1})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/methods?country=US", nil)

	handler := &Handler{
		paymentService: service.NewPaymentService(nil, repository.NewPaymentRepository(db)),
		protection:     protection,
	}
	handler.ListPaymentMethods(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Data struct {
			Data []paymentMethodResponse `json:"data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Len(t, payload.Data.Data, 2)

	byCode := map[string]paymentMethodResponse{}
	for _, item := range payload.Data.Data {
		byCode[item.Code] = item
	}
	require.False(t, byCode["card"].Available)
	require.Equal(t, "stripe", byCode["card"].Provider)
	require.Equal(t, "temporarily_unavailable", byCode["card"].UnavailableReason)
	require.True(t, byCode["paypal"].Available)
	require.Equal(t, "paypal", byCode["paypal"].Provider)
}

func TestListPaymentMethodsUsesCountryQueryForAvailability(t *testing.T) {
	db := newPaymentMethodAvailabilityTestDB(t)
	require.NoError(t, db.Create(&paymentdomain.PaymentMethod{Name: "Card", Code: "card", Enabled: true}).Error)

	protection := service.NewPaymentProtectionService(
		repository.NewPaymentProtectionRepository(db),
		config.PaymentProtectionConfig{
			Enabled:                 true,
			MaxControlDurationHours: 24,
		},
	)
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
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/methods?country=US", nil)

	handler := &Handler{
		paymentService: service.NewPaymentService(nil, repository.NewPaymentRepository(db)),
		protection:     protection,
	}
	handler.ListPaymentMethods(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Data struct {
			Data []paymentMethodResponse `json:"data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Len(t, payload.Data.Data, 1)
	require.False(t, payload.Data.Data[0].Available)
	require.Equal(t, "temporarily_unavailable", payload.Data.Data[0].UnavailableReason)
}

func TestListPaymentMethodsMarksUnsupportedDefaultCurrencyUnavailable(t *testing.T) {
	db := newPaymentMethodAvailabilityTestDB(t)
	require.NoError(t, db.Create(&paymentdomain.PaymentMethod{Name: "Card", Code: "card", Enabled: true}).Error)
	require.NoError(t, db.Create(&paymentdomain.PaymentMethod{Name: "WeChat Pay", Code: "wechat_pay", Enabled: true}).Error)

	currencyPolicy := service.NewCurrencyPolicyService(repository.NewSettingRepository(db))
	_, err := currencyPolicy.UpdatePolicy(currency.Policy{
		AccountingCurrency:   "USD",
		DefaultOrderCurrency: "USD",
		AcceptedCurrencies:   []string{"USD"},
	})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/methods", nil)

	handler := &Handler{
		paymentService: service.NewPaymentService(nil, repository.NewPaymentRepository(db)),
		currencyPolicy: currencyPolicy,
	}
	handler.ListPaymentMethods(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Data struct {
			Data []paymentMethodResponse `json:"data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Len(t, payload.Data.Data, 2)

	byCode := map[string]paymentMethodResponse{}
	for _, item := range payload.Data.Data {
		byCode[item.Code] = item
	}
	require.True(t, byCode["card"].Available)
	require.False(t, byCode["wechat_pay"].Available)
	require.Equal(t, "currency_not_supported", byCode["wechat_pay"].UnavailableReason)
}

func newPaymentMethodAvailabilityTestDB(t *testing.T) *gorm.DB {
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
		&paymentdomain.PaymentMethod{},
		&paymentdomain.PaymentProtectionControl{},
		&audit.AuditLog{},
		&setting.Setting{},
	))
	return db
}
