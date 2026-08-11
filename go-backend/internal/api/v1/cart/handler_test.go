package cart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetCartSummaryDoesNotCreatePassiveVisitorState(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	router := gin.New()
	router.GET("/cart/summary", handler.GetCartSummary)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/cart/summary", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.NotContains(t, strings.Join(recorder.Header().Values("Set-Cookie"), ";"), "session_id=")
	assertTableCount(t, db, &product.Cart{}, 0)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func TestAddToCartCreatesMeaningfulVisitorProfile(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	productRecord := seedPurchasableProduct(t, db)

	router := gin.New()
	trustedEdge, err := middleware.NewTrustedEdgeMetadata([]string{"10.0.0.0/8"})
	require.NoError(t, err)
	router.Use(trustedEdge)
	router.POST("/cart/add", handler.AddToCart)

	recorder := httptest.NewRecorder()
	body := []byte(`{"product_id":` + strconv.FormatUint(uint64(productRecord.ID), 10) + `,"quantity":1}`)
	request := httptest.NewRequest(http.MethodPost, "/cart/add", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("CF-IPCountry", "US")
	request.Header.Set("CF-Connecting-IP", "203.0.113.10")
	request.Header.Set("User-Agent", "Cart Handler Test")
	request.RemoteAddr = "10.1.0.8:1234"
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	sessionCookie := responseCookie(recorder, "session_id")
	require.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
	assertTableCount(t, db, &product.Cart{}, 1)
	assertTableCount(t, db, &visitor.Profile{}, 1)

	var profile visitor.Profile
	require.NoError(t, db.First(&profile).Error)
	assert.Equal(t, sessionCookie.Value, profile.CartSessionID)
	assert.Equal(t, "en-us", profile.Locale)
	assert.Equal(t, "US", profile.CountryCode)
	assert.NotEmpty(t, profile.IPHash)
	assert.NotEmpty(t, profile.UserAgentHash)
}

func TestGetCartSummaryForExistingCartDoesNotIncreaseVisitorQuality(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	productRecord := seedPurchasableProduct(t, db)

	router := gin.New()
	router.POST("/cart/add", handler.AddToCart)
	router.GET("/cart/summary", handler.GetCartSummary)

	addRecorder := httptest.NewRecorder()
	body := []byte(`{"product_id":` + strconv.FormatUint(uint64(productRecord.ID), 10) + `,"quantity":1}`)
	addRequest := httptest.NewRequest(http.MethodPost, "/cart/add", bytes.NewReader(body))
	addRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(addRecorder, addRequest)

	require.Equal(t, http.StatusOK, addRecorder.Code)
	sessionCookie := responseCookie(addRecorder, "session_id")
	require.NotNil(t, sessionCookie)

	var before visitor.Profile
	require.NoError(t, db.First(&before).Error)
	require.Equal(t, service.VisitorProfileQualityCartAction, before.ProfileQualityScore)
	require.Equal(t, service.VisitorProfileActionCart, before.LastMeaningfulAction)
	require.NotNil(t, before.LastMeaningfulSeenAt)
	lastMeaningfulSeenAt := *before.LastMeaningfulSeenAt

	summaryRecorder := httptest.NewRecorder()
	summaryRequest := httptest.NewRequest(http.MethodGet, "/cart/summary", nil)
	summaryRequest.AddCookie(sessionCookie)
	router.ServeHTTP(summaryRecorder, summaryRequest)

	require.Equal(t, http.StatusOK, summaryRecorder.Code)
	var after visitor.Profile
	require.NoError(t, db.First(&after, before.ID).Error)
	assert.Equal(t, before.ProfileQualityScore, after.ProfileQualityScore)
	assert.Equal(t, before.LastMeaningfulAction, after.LastMeaningfulAction)
	require.NotNil(t, after.LastMeaningfulSeenAt)
	assert.Equal(t, lastMeaningfulSeenAt, *after.LastMeaningfulSeenAt)
}

func TestGetCartSummaryOmitsExactInventory(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	productRecord := seedPurchasableProduct(t, db)
	cartRecord := seedEmptyCart(t, db, "inventory-safe-summary")

	var variant product.ProductVariant
	require.NoError(t, db.Where("product_id = ?", productRecord.ID).First(&variant).Error)
	require.NoError(t, db.Create(&product.CartItem{
		CartID:    cartRecord.ID,
		ProductID: productRecord.ID,
		VariantID: &variant.ID,
		Quantity:  1,
		Price:     100,
	}).Error)

	router := gin.New()
	router.GET("/cart/summary", handler.GetCartSummary)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/cart/summary", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: cartRecord.SessionID})
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	data := response["data"].(map[string]any)
	items := data["items"].([]any)
	require.Len(t, items, 1)
	item := items[0].(map[string]any)
	publicProduct := item["product"].(map[string]any)
	publicVariant := item["variant"].(map[string]any)

	assert.NotContains(t, publicProduct, "stock")
	assert.NotContains(t, publicVariant, "stock")
	for _, field := range []string{
		"product_type_id",
		"shipping_template_id",
		"sku",
		"weight_grams",
	} {
		assert.NotContains(t, string(recorder.Body.Bytes()), `"`+field+`"`)
	}
	assert.Equal(t, "in_stock", publicProduct["availability"])
	assert.Equal(t, "in_stock", publicVariant["availability"])
}

func TestAddToCartRejectsInvalidProductWithoutPassiveVisitorState(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)

	router := gin.New()
	router.POST("/cart/add", handler.AddToCart)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/cart/add", bytes.NewReader([]byte(`{"product_id":999999,"quantity":1}`)))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, responseCookie(recorder, "session_id"))
	assertTableCount(t, db, &product.Cart{}, 0)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func TestSyncCartWithEmptyPayloadDoesNotCreatePassiveVisitorState(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)

	router := gin.New()
	router.POST("/cart/sync", handler.SyncCart)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/cart/sync", bytes.NewReader([]byte(`[]`)))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, responseCookie(recorder, "session_id"))
	assertTableCount(t, db, &product.Cart{}, 0)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func TestSyncCartWithInvalidItemsDoesNotCreatePassiveVisitorState(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)

	router := gin.New()
	router.POST("/cart/sync", handler.SyncCart)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/cart/sync", bytes.NewReader([]byte(`[{"product_id":999999,"quantity":1}]`)))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, responseCookie(recorder, "session_id"))
	assertTableCount(t, db, &product.Cart{}, 0)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func TestSyncCartWithItemsCreatesMeaningfulVisitorProfile(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	productRecord := seedPurchasableProduct(t, db)

	router := gin.New()
	router.POST("/cart/sync", handler.SyncCart)

	recorder := httptest.NewRecorder()
	body := []byte(`[{` +
		`"product_id":` + strconv.FormatUint(uint64(productRecord.ID), 10) + `,` +
		`"quantity":2` +
		`}]`)
	request := httptest.NewRequest(http.MethodPost, "/cart/sync", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("CF-IPCountry", "US")
	request.Header.Set("CF-Connecting-IP", "203.0.113.11")
	request.Header.Set("User-Agent", "Cart Sync Handler Test")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	sessionCookie := responseCookie(recorder, "session_id")
	require.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
	assertTableCount(t, db, &product.Cart{}, 1)
	assertTableCount(t, db, &product.CartItem{}, 1)
	assertTableCount(t, db, &visitor.Profile{}, 1)

	var profile visitor.Profile
	require.NoError(t, db.First(&profile).Error)
	assert.Equal(t, sessionCookie.Value, profile.CartSessionID)
	assert.Equal(t, "en-us", profile.Locale)
}

func TestRemoveFromEmptyCartDoesNotCreateVisitorProfile(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	seedEmptyCart(t, db, "empty-cart-session")

	router := gin.New()
	router.DELETE("/cart/items/:id", handler.RemoveFromCart)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/cart/items/999999", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: "empty-cart-session"})
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assertTableCount(t, db, &product.Cart{}, 1)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func TestClearEmptyCartDoesNotCreateVisitorProfile(t *testing.T) {
	db, handler := newCartHandlerTestEnv(t)
	seedEmptyCart(t, db, "empty-cart-session")

	router := gin.New()
	router.POST("/cart/clear", handler.ClearCart)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/cart/clear", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: "empty-cart-session"})
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assertTableCount(t, db, &product.Cart{}, 1)
	assertTableCount(t, db, &visitor.Profile{}, 0)
}

func seedPurchasableProduct(t *testing.T, db *gorm.DB) product.Product {
	t.Helper()

	productRecord := product.Product{
		SKU:   "CART-HANDLER-RIM",
		Name:  "Cart Handler Rim",
		Slug:  "cart-handler-rim",
		Price: 100,
		Stock: 10,
	}
	require.NoError(t, db.Create(&productRecord).Error)

	variant := product.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "CART-HANDLER-RIM-DEFAULT",
		OptionValues: "{}",
		Price:        100,
		Stock:        10,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)
	return productRecord
}

func seedEmptyCart(t *testing.T, db *gorm.DB, sessionID string) product.Cart {
	t.Helper()

	cartRecord := product.Cart{SessionID: sessionID}
	require.NoError(t, db.Create(&cartRecord).Error)
	return cartRecord
}

func responseCookie(recorder *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func assertTableCount(t *testing.T, db *gorm.DB, model any, expected int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Count(&count).Error)
	assert.EqualValues(t, expected, count)
}

func newCartHandlerTestEnv(t *testing.T) (*gorm.DB, *Handler) {
	t.Helper()

	gin.SetMode(gin.TestMode)
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

	require.NoError(t, db.AutoMigrate(
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.Cart{},
		&product.CartItem{},
		&visitor.Profile{},
	))

	cartService := service.NewCartService(repository.NewCartRepository(db), repository.NewProductRepository(db))
	visitorProfileService := service.NewVisitorProfileService(repository.NewVisitorProfileRepository(db))
	return db, NewHandler(cartService, Options{
		VisitorProfileService: visitorProfileService,
		VisitorSecret:         "handler-test-secret",
	})
}
