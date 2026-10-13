package shipping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPublicShippingHTTPResponsesExcludeProviderCostsAndPreserveCustomerPricing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shippingdomain.Carrier{}, &shippingdomain.CarrierService{},
		&shippingdomain.ShippingRouteGroup{}, &shippingdomain.ShippingRouteGroupCarrierService{},
		&shippingdomain.ShippingTemplate{}, &shippingdomain.ShippingRule{}))

	carrier := shippingdomain.Carrier{Name: "Independent provider", Code: "INDEPENDENT", Enabled: true,
		Contact: "private provider contact", Email: "provider@example.test"}
	require.NoError(t, db.Create(&carrier).Error)
	route := shippingdomain.CarrierService{
		CarrierID: carrier.ID, ServiceCode: "ROUTE-US", ServiceName: "US route", Countries: `["US"]`, Enabled: true,
		Currency: "CNY", BillingMode: "volumetric_weight", FirstWeightGrams: 9900,
		AdditionalWeightGrams: 9900, MinChargeWeightGrams: 99000, VolumetricDivisor: 1,
		FuelSurchargePercentDecimal: "95", RemoteSurchargeMinor: 99999,
		RemotePostalCodes: `["999*"]`, EtaMinDays: 90, EtaMaxDays: 99, Description: "private provider cost notes",
	}
	require.NoError(t, db.Create(&route).Error)
	shippingRepository := repository.NewShippingRepository(db)
	aggregate := shippingdomain.ShippingRouteGroup{Name: "US routes", Enabled: true}
	require.NoError(t, shippingRepository.CreateShippingRouteGroup(&aggregate, []uint{route.ID}))
	template := shippingdomain.ShippingTemplate{
		Name: "Buyer standard delivery", Type: "weight", Currency: "USD", Enabled: true, RouteGroupID: &aggregate.ID,
		CustomerBillingMode: "actual_weight", CustomerFirstWeightGrams: 500, CustomerAdditionalWeightGrams: 100,
		CustomerFuelSurchargePercentDecimal: "2", CustomerRemoteSurchargeMinor: 300,
		CustomerRemotePostalCodes: `["123*"]`, CustomerEtaMinDays: 3, CustomerEtaMaxDays: 7,
	}
	require.NoError(t, shippingRepository.CreateTemplateWithRules(&template, []shippingdomain.ShippingRule{
		{Region: "US", Currency: "USD", FeeMinor: 1200, AdditionalMinor: 100},
	}))
	handler := NewHandler(service.NewShippingService(shippingRepository))
	router := gin.New()
	router.GET("/templates", handler.ListTemplates)
	router.GET("/templates/:id", handler.GetTemplate)
	router.GET("/carrier-services", handler.ListCarrierServices)
	router.GET("/carrier-services/:id", handler.GetCarrierService)

	for _, testCase := range []struct {
		name       string
		path       string
		isList     bool
		isTemplate bool
	}{
		{"template list", "/templates", true, true},
		{"template detail", fmt.Sprintf("/templates/%d", template.ID), false, true},
		{"route list", "/carrier-services", true, false},
		{"route detail", fmt.Sprintf("/carrier-services/%d", route.ID), false, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, testCase.path, nil))
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			var envelope map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			payload := envelope["data"].(map[string]interface{})
			if testCase.isList {
				entries := payload["data"].([]interface{})
				require.Len(t, entries, 1)
				payload = entries[0].(map[string]interface{})
			}
			if testCase.isTemplate {
				assert.Equal(t, "USD", payload["currency"])
				assert.Equal(t, "actual_weight", payload["customer_billing_mode"])
				assert.Equal(t, float64(500), payload["customer_first_weight_grams"])
				assert.Equal(t, "2", payload["customer_fuel_surcharge_percent_decimal"])
				assert.Equal(t, float64(300), payload["customer_remote_surcharge_minor"])
				assert.Equal(t, float64(7), payload["customer_eta_max_days"])
				rules := payload["rules"].([]interface{})
				require.Len(t, rules, 1)
				assert.Equal(t, float64(1200), rules[0].(map[string]interface{})["fee_minor"])
				aggregatePayload := payload["route_group"].(map[string]interface{})
				assert.Equal(t, `["US"]`, aggregatePayload["countries"])
				members := aggregatePayload["carrier_services"].([]interface{})
				require.Len(t, members, 1)
				payload = members[0].(map[string]interface{})
			}
			assert.Equal(t, "ROUTE-US", payload["service_code"])
			assert.Equal(t, `["US"]`, payload["countries"])
			for _, providerField := range []string{
				"currency", "billing_mode", "first_weight_grams", "additional_weight_grams",
				"min_charge_weight_grams", "volumetric_divisor", "fuel_surcharge_percent_decimal",
				"remote_surcharge_minor", "remote_postal_codes", "eta_min_days", "eta_max_days", "description",
			} {
				assert.NotContains(t, payload, providerField)
			}
			assert.NotContains(t, recorder.Body.String(), "private provider")
			assert.NotContains(t, recorder.Body.String(), "provider@example.test")
		})
	}
	// Public response construction must not alter the provider's persisted facts.
	var storedRoute shippingdomain.CarrierService
	require.NoError(t, db.First(&storedRoute, route.ID).Error)
	assert.Equal(t, int64(99999), storedRoute.RemoteSurchargeMinor)
	assert.Equal(t, "95", storedRoute.FuelSurchargePercentDecimal)
}
