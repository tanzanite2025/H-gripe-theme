package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenGatewayCreatesOrderWithSignedOfficialPayload(t *testing.T) {
	const userID = "create-user"
	const apiToken = "create-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := newYanwenCreateOrderTestServer(t, userID, apiToken, `{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE123","orderNumber":"ORDER-1","yanwenOrderNumber":"YW-1"}}`)
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewYanwenGatewayClient()
	client.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	client.now = func() time.Time { return fixedTime }

	result, err := client.CreateYanwenOrder(context.Background(), yanwenGatewayCredentials{
		environment: "fat",
		endpoint:    yanwenFATEndpoint,
		userID:      userID,
		apiToken:    apiToken,
	}, YanwenCreateOrderRequest{
		ChannelID:   "481",
		OrderSource: "tanzanite-theme",
		OrderNumber: "ORDER-1",
		CompanyCode: "WH01",
		ReceiverInfo: YanwenReceiverInfo{
			Name: "Rider Example", Country: "US", Address: "1 Main Street",
		},
		ParcelInfo: YanwenParcelInfo{
			HasBattery: 0, Currency: "USD", TotalQuantity: 1, TotalWeight: 100,
			ProductList: []YanwenProductDeclare{{
				GoodsNameChinese: "辐条", GoodsNameEnglish: "Spoke", Price: json.Number("12.00"), PriceExport: json.Number("12.00"), Quantity: 1, Weight: 100,
			}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "YE123", result.WaybillNumber)
	require.Equal(t, "YW-1", result.YanwenOrderNumber)
}

func TestYanwenGatewayGetsOrderDetailsWithSignedOfficialPayload(t *testing.T) {
	const userID = "get-user"
	const apiToken = "get-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenGetOrderMethod, request.URL.Query().Get("method"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"waybillNumber":"YE-GET"}`, string(body))
		raw := userID + string(body) + "json" + yanwenGetOrderMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-GET","orderNumber":"ORDER-GET","referenceNumber":"REF-GET","yanwenOrderNumber":"YW-GET","status":"3","isPrint":"1"}}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	client.now = func() time.Time { return fixedTime }
	result, err := client.GetYanwenOrderDetails(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: userID, apiToken: apiToken,
	}, "YE-GET")
	require.NoError(t, err)
	require.Equal(t, "YE-GET", result.WaybillNumber)
	require.Equal(t, "ORDER-GET", result.OrderNumber)
	require.Equal(t, "REF-GET", result.ReferenceNumber)
	require.Equal(t, 3, result.OfficialStatus)
	require.True(t, result.IsPrinted)
	require.NotEmpty(t, result.RawResponse)
}

func TestYanwenGatewayGetsBatchOrderDetailsWithSignedOfficialPayload(t *testing.T) {
	const userID = "batch-get-user"
	const apiToken = "batch-get-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenGetOrderListMethod, request.URL.Query().Get("method"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"listNumber":["YE-BATCH-1","ORDER-BATCH-2"]}`, string(body))
		raw := userID + string(body) + "json" + yanwenGetOrderListMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"waybillNumber":"YE-BATCH-1","orderNumber":"ORDER-BATCH-1","referenceNumber":"REF-1","yanwenOrderNumber":"YW-1","status":"3","isPrint":"1"},{"waybillNumber":"YE-BATCH-2","orderNumber":"ORDER-BATCH-2","referenceNumber":"REF-2","yanwenOrderNumber":"YW-2","status":5,"isPrint":0}]}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	client.now = func() time.Time { return fixedTime }
	results, err := client.GetYanwenOrderDetailsBatch(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: userID, apiToken: apiToken,
	}, []string{"YE-BATCH-1", "ORDER-BATCH-2"})
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "YE-BATCH-1", results[0].WaybillNumber)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusInTransit, results[0].OfficialStatus)
	require.True(t, results[0].IsPrinted)
	require.Equal(t, "YE-BATCH-2", results[1].WaybillNumber)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusCancelled, results[1].OfficialStatus)
	require.NotEmpty(t, results[0].RawResponse)
}

func TestYanwenGatewayRejectsInvalidBatchOrderDetailRequestsAndResponses(t *testing.T) {
	client := NewYanwenGatewayClient()
	_, err := client.GetYanwenOrderDetailsBatch(context.Background(), yanwenGatewayCredentials{}, nil)
	require.ErrorContains(t, err, "at least one")
	identifiers := make([]string, yanwenGetOrderListMaximumIdentifiers+1)
	for index := range identifiers {
		identifiers[index] = "YE-" + strconv.Itoa(index)
	}
	_, err = client.GetYanwenOrderDetailsBatch(context.Background(), yanwenGatewayCredentials{}, identifiers)
	require.ErrorContains(t, err, "at most 50")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"waybillNumber":"YE-BATCH","orderNumber":"ORDER-BATCH","status":"14","isPrint":"0"}]}`))
	}))
	defer server.Close()
	client.httpClient = server.Client()
	_, err = client.GetYanwenOrderDetailsBatch(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: "user", apiToken: "token",
	}, []string{"YE-BATCH"})
	require.ErrorContains(t, err, "unknown status")
}

func TestYanwenGatewayCancelsOneOrderWithOptionalNoteAndAcceptsNullData(t *testing.T) {
	const userID = "cancel-user"
	const apiToken = "cancel-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenCancelOrderMethod, request.URL.Query().Get("method"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"waybillNumber":"YE-CANCEL","note":"duplicate order"}`, string(body))
		raw := userID + string(body) + "json" + yanwenCancelOrderMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":null}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	client.now = func() time.Time { return fixedTime }
	result, err := client.CancelYanwenOrder(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: userID, apiToken: apiToken,
	}, "YE-CANCEL", "duplicate order")
	require.NoError(t, err)
	require.Equal(t, "YE-CANCEL", result.WaybillNumber)
	require.NotEmpty(t, result.RawResponse)
}

func TestYanwenGatewayRejectsMissingWaybillAndOfficialCancellationErrors(t *testing.T) {
	client := NewYanwenGatewayClient()
	_, err := client.CancelYanwenOrder(context.Background(), yanwenGatewayCredentials{}, "", "")
	require.ErrorContains(t, err, "requires waybillNumber")

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"success":false,"code":"CANCEL_400","message":"waybill cannot be cancelled","data":null}`))
	}))
	defer server.Close()
	client.httpClient = server.Client()
	_, err = client.CancelYanwenOrder(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: "user", apiToken: "token",
	}, "YE-CANCEL", "")
	require.ErrorContains(t, err, "waybill cannot be cancelled")
}

func TestYanwenGatewayRejectsUnknownOrMissingOfficialWaybillFields(t *testing.T) {
	responses := []string{
		`{"success":true,"code":"0","data":{"waybillNumber":"YE-GET","orderNumber":"ORDER-GET","isPrint":"1"}}`,
		`{"success":true,"code":"0","data":{"waybillNumber":"YE-GET","orderNumber":"ORDER-GET","status":"14","isPrint":"1"}}`,
	}
	for _, responseBody := range responses {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(responseBody))
		}))
		client := NewYanwenGatewayClient()
		client.httpClient = server.Client()
		_, err := client.GetYanwenOrderDetails(context.Background(), yanwenGatewayCredentials{
			environment: "fat", endpoint: server.URL, userID: "user", apiToken: "token",
		}, "YE-GET")
		require.Error(t, err)
		server.Close()
	}
}

func TestYanwenGatewayGetsOneOfficialPDFLabelWithSignedPayload(t *testing.T) {
	const userID = "label-user"
	const apiToken = "label-token"
	fixedTime := time.UnixMilli(1790000000123)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenGetOrderLabelMethod, request.URL.Query().Get("method"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"waybillNumber":"YE-LABEL"}`, string(body))
		raw := userID + string(body) + "json" + yanwenGetOrderLabelMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-LABEL","isSuccess":true,"errorMsg":"","base64String":"JVBERi0xLjQ="}}`))
	}))
	defer server.Close()

	client := NewYanwenGatewayClient()
	client.httpClient = server.Client()
	client.now = func() time.Time { return fixedTime }
	result, err := client.GetYanwenOrderLabel(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: server.URL, userID: userID, apiToken: apiToken,
	}, "YE-LABEL")
	require.NoError(t, err)
	require.Equal(t, "YE-LABEL", result.WaybillNumber)
	require.True(t, result.IsSuccessful)
	require.Equal(t, "JVBERi0xLjQ=", result.Base64String)
	require.NotEmpty(t, result.RawResponse)
}

func TestYanwenGatewayRejectsFailedOrInvalidOfficialPDFLabelResponses(t *testing.T) {
	responses := []string{
		`{"success":true,"code":"0","data":{"waybillNumber":"YE-LABEL","isSuccess":false,"errorMsg":"label unavailable","base64String":""}}`,
		`{"success":true,"code":"0","data":{"waybillNumber":"YE-LABEL","isSuccess":true,"base64String":"not-base64"}}`,
		`{"success":true,"code":"0","data":{"waybillNumber":"OTHER","isSuccess":true,"base64String":"JVBERi0xLjQ="}}`,
	}
	for _, responseBody := range responses {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(responseBody))
		}))
		client := NewYanwenGatewayClient()
		client.httpClient = server.Client()
		_, err := client.GetYanwenOrderLabel(context.Background(), yanwenGatewayCredentials{
			environment: "fat", endpoint: server.URL, userID: "user", apiToken: "token",
		}, "YE-LABEL")
		require.Error(t, err)
		server.Close()
	}
}

func TestYanwenAPIServiceCreatesAndPersistsRealWaybillIdempotently(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&order.Order{}, &order.OrderItem{},
		&shipping.YanwenAPIConfig{}, &shipping.YanwenProductCatalogEntry{}, &shipping.YanwenCountryCatalogEntry{},
		&shipping.YanwenWarehouseCatalogEntry{}, &shipping.YanwenPublishedChannel{}, &shipping.YanwenWaybill{},
	))

	const userID = "waybill-user"
	const apiToken = "waybill-token"
	createRequestCount := 0
	rejectAddressPreflight := false
	server := httpServerFunc(t, func(writer http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		switch request.URL.Query().Get("method") {
		case yanwenVerifyUnitedStatesAddressMethod:
			require.JSONEq(t, `{"receiverInfo":{"address":"1 Main Street","zipCode":"78701","city":"Austin","state":"TX"}}`, string(body))
			if rejectAddressPreflight {
				_, _ = writer.Write([]byte(`{"success":false,"code":"ADDRESS_INVALID","message":"address rejected","data":null}`))
				return
			}
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"address ok","data":{"receiverInfo":{"address":"1 Main Street","city":"Austin","state":"TX","zipCode4":"7870","zipCode5":"78701"}}}`))
		case yanwenCreateOrderMethod:
			createRequestCount++
			require.Equal(t, "481", payload["channelId"])
			require.Equal(t, "WH01", payload["companyCode"])
			receiverInfo := payload["receiverInfo"].(map[string]any)
			require.Equal(t, "US", receiverInfo["country"])
			require.Equal(t, "US-TAX-123", receiverInfo["taxNumber"])
			parcelInfo := payload["parcelInfo"].(map[string]any)
			require.Equal(t, float64(0), parcelInfo["hasBattery"])
			require.Equal(t, "IOSS-123", parcelInfo["ioss"])
			productList := parcelInfo["productList"].([]any)
			require.Len(t, productList, 1)
			declaredProduct := productList[0].(map[string]any)
			require.Equal(t, "Spoke", declaredProduct["goodsNameCh"])
			require.Equal(t, "Bicycle spoke", declaredProduct["goodsNameEn"])
			importCustomsInfo := payload["importCustomsInfo"].(map[string]any)
			require.Equal(t, "DE123456789000", importCustomsInfo["eori"])
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-SAVED","orderNumber":"ORDER-1","yanwenOrderNumber":"YW-SAVED"}}`))
		default:
			t.Fatalf("unexpected Yanwen method %q", request.URL.Query().Get("method"))
		}
	})
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	productRepo := repository.NewYanwenProductCatalogRepository(db)
	countryRepo := repository.NewYanwenCountryCatalogRepository(db)
	warehouseRepo := repository.NewYanwenWarehouseCatalogRepository(db)
	publishedChannelRepo := repository.NewYanwenPublishedChannelRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, productRepo)
	apiService.ConfigureYanwenCountryCatalogRepository(countryRepo)
	apiService.ConfigureYanwenWarehouseCatalogRepository(warehouseRepo)
	apiService.ConfigureYanwenWaybillRepositories(orderRepo, waybillRepo)
	apiService.ConfigureYanwenPublishedChannelRepository(publishedChannelRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: userID, APIToken: apiToken, Enabled: true}))
	now := time.Now().UTC()
	_, err = productRepo.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{{ProductID: "481", NameChinese: "普货", NameEnglish: "Tracked"}}, now)
	require.NoError(t, err)
	_, err = countryRepo.UpsertYanwenCountryCatalogEntries("fat", []shipping.YanwenCountryCatalogEntry{{CountryID: "840", CountryCode: "US", NameEnglish: "United States"}}, now)
	require.NoError(t, err)
	_, err = warehouseRepo.UpsertYanwenWarehouseCatalogEntries("fat", []shipping.YanwenWarehouseCatalogEntry{{WarehouseCode: "WH01", Name: "华东仓"}}, now)
	require.NoError(t, err)
	channel := shipping.YanwenPublishedChannel{Environment: "fat", ProductCode: "481", DisplayName: "燕文普货", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000}
	require.NoError(t, publishedChannelRepo.CreateYanwenPublishedChannel(&channel))
	productionChannel := shipping.YanwenPublishedChannel{
		Environment: "production", ProductCode: "481", DisplayName: "燕文生产普货", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, publishedChannelRepo.CreateYanwenPublishedChannel(&productionChannel))
	declared := int64(1200)
	variantID := uint(1)
	orderRecord := &order.Order{
		OrderNumber: "ORDER-1", Status: "processing", PaymentStatus: "paid", Currency: "USD", PaymentCurrency: "USD", PaymentAmountMinor: 1200,
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "US", Phone: "123", Email: "rider@example.com"},
		Items:           []order.OrderItem{{ProductID: 1, VariantID: &variantID, ProductName: "Spoke", CustomsDescription: "Bicycle spoke", Quantity: 1, Currency: "USD", PriceMinor: 1200, SubtotalMinor: 1200, TotalMinor: 1200, WeightGrams: 100, DeclaredValueMinor: &declared, DeclaredValueConfirmed: true}},
	}
	require.NoError(t, orderRepo.Create(orderRecord))

	hasBattery := false
	_, err = apiService.CreateYanwenWaybill(context.Background(), YanwenCreateWaybillInput{Environment: "production", OrderID: orderRecord.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery})
	require.ErrorContains(t, err, "belongs to the fat environment")
	require.Zero(t, createRequestCount, "a channel from another environment must be rejected before an official request")

	created, err := apiService.CreateYanwenWaybill(context.Background(), YanwenCreateWaybillInput{
		Environment:       "fat",
		OrderID:           orderRecord.ID,
		ChannelID:         channel.ID,
		WarehouseCode:     "WH01",
		HasBattery:        &hasBattery,
		ReceiverTaxNumber: "US-TAX-123",
		IOSS:              "IOSS-123",
		EORI:              "DE123456789000",
	})
	require.NoError(t, err)
	require.Equal(t, "YE-SAVED", created.WaybillNumber)
	require.Equal(t, "Bicycle spoke", created.DeclaredDescription)
	require.Equal(t, 1, createRequestCount)
	second, err := apiService.CreateYanwenWaybill(context.Background(), YanwenCreateWaybillInput{Environment: "fat", OrderID: orderRecord.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery})
	require.NoError(t, err)
	require.Equal(t, created.ID, second.ID)
	require.Equal(t, 1, createRequestCount, "a persisted successful waybill must prevent a duplicate official create call")

	rejectedOrder := &order.Order{
		OrderNumber: "ORDER-2", Status: "processing", PaymentStatus: "paid", Currency: "USD", PaymentCurrency: "USD", PaymentAmountMinor: 1200,
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "US", Phone: "123", Email: "rider@example.com"},
		Items:           []order.OrderItem{{ProductID: 1, VariantID: &variantID, ProductName: "Spoke", CustomsDescription: "Bicycle spoke", Quantity: 1, Currency: "USD", PriceMinor: 1200, SubtotalMinor: 1200, TotalMinor: 1200, WeightGrams: 100, DeclaredValueMinor: &declared, DeclaredValueConfirmed: true}},
	}
	require.NoError(t, orderRepo.Create(rejectedOrder))
	rejectAddressPreflight = true
	_, err = apiService.CreateYanwenWaybill(context.Background(), YanwenCreateWaybillInput{
		Environment: "fat", OrderID: rejectedOrder.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery,
	})
	require.ErrorContains(t, err, "verify Yanwen US address before waybill creation")
	require.Equal(t, 1, createRequestCount, "an official customs rejection must block express.order.create")
}

func TestYanwenWaybillCustomsPreflightBlocksOfficialUnitedStatesRejection(t *testing.T) {
	requestMethods := make([]string, 0, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestMethods = append(requestMethods, request.URL.Query().Get("method"))
		_, _ = writer.Write([]byte(`{"success":false,"code":"ADDRESS_INVALID","message":"address rejected","data":null}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewYanwenGatewayClient()
	client.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	service := &YanwenAPIService{gateway: client}

	err = service.verifyYanwenWaybillCustomsPreflight(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: yanwenFATEndpoint, userID: "user", apiToken: "token",
	}, "US", YanwenCreateOrderRequest{ReceiverInfo: YanwenReceiverInfo{
		Name: "Rider Example", Address: "1 Main Street", City: "Austin", State: "TX", ZipCode: "78701",
	}})
	require.ErrorContains(t, err, "verify Yanwen US address before waybill creation")
	require.ErrorContains(t, err, "address rejected")
	require.Equal(t, []string{yanwenVerifyUnitedStatesAddressMethod}, requestMethods)
}

func TestYanwenWaybillCustomsPreflightRequiresKoreaPCCCBeforeOfficialCall(t *testing.T) {
	service := &YanwenAPIService{gateway: NewYanwenGatewayClient()}
	err := service.verifyYanwenWaybillCustomsPreflight(context.Background(), yanwenGatewayCredentials{
		environment: "fat", endpoint: yanwenFATEndpoint, userID: "user", apiToken: "token",
	}, "KR", YanwenCreateOrderRequest{ReceiverInfo: YanwenReceiverInfo{
		Name: "Korean Recipient", Phone: "010-1234-5678", ZipCode: "04524",
	}})
	require.ErrorContains(t, err, "Korea PCCC tax number is required")
}

func TestYanwenAPIServiceSyncsOfficialWaybillDetailsAndPersistsTheLatestResponse(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	const userID = "sync-user"
	const apiToken = "sync-token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenGetOrderMethod, request.URL.Query().Get("method"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-SYNC","orderNumber":"ORDER-SYNC","referenceNumber":"REF-SYNC","yanwenOrderNumber":"YW-SYNC","status":"4","isPrint":"1"}}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: userID, APIToken: apiToken, Enabled: true}))
	waybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 11, OrderNumber: "ORDER-SYNC", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-SYNC", YanwenOrderNumber: "YW-CREATE",
		Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(waybill))

	synced, err := apiService.SyncYanwenWaybillOfficialDetails(context.Background(), waybill.ID)
	require.NoError(t, err)
	require.Equal(t, "REF-SYNC", synced.ReferenceNumber)
	require.Equal(t, "YW-SYNC", synced.YanwenOrderNumber)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusDelivered, synced.OfficialStatus)
	require.True(t, synced.IsPrinted)
	require.NotNil(t, synced.LastOfficialSyncedAt)
	require.JSONEq(t, `{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-SYNC","orderNumber":"ORDER-SYNC","referenceNumber":"REF-SYNC","yanwenOrderNumber":"YW-SYNC","status":"4","isPrint":"1"}}`, string(synced.ResponseData))

	stored, err := waybillRepo.FindYanwenWaybillByID(waybill.ID)
	require.NoError(t, err)
	require.Equal(t, "REF-SYNC", stored.ReferenceNumber)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusDelivered, stored.OfficialStatus)
	require.True(t, stored.IsPrinted)
	require.NotNil(t, stored.LastOfficialSyncedAt)
}

func TestYanwenAPIServiceBatchSyncsAndAtomicallyPersistsOfficialWaybillDetails(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	const userID = "batch-sync-user"
	const apiToken = "batch-sync-token"
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		require.Equal(t, yanwenGetOrderListMethod, request.URL.Query().Get("method"))
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		require.JSONEq(t, `{"listNumber":["YE-BATCH-SERVICE-1","YE-BATCH-SERVICE-2"]}`, string(body))
		raw := userID + string(body) + "json" + yanwenGetOrderListMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"waybillNumber":"YE-BATCH-SERVICE-2","orderNumber":"ORDER-BATCH-SERVICE-2","referenceNumber":"REF-BATCH-2","yanwenOrderNumber":"YW-BATCH-2","status":"5","isPrint":"0"},{"waybillNumber":"YE-BATCH-SERVICE-1","orderNumber":"ORDER-BATCH-SERVICE-1","referenceNumber":"REF-BATCH-1","yanwenOrderNumber":"YW-BATCH-1","status":"3","isPrint":"1"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: userID, APIToken: apiToken, Enabled: true}))

	firstWaybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 31, OrderNumber: "ORDER-BATCH-SERVICE-1", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-BATCH-SERVICE-1", Status: shipping.YanwenWaybillStatusCreated,
	}
	secondWaybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 32, OrderNumber: "ORDER-BATCH-SERVICE-2", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-BATCH-SERVICE-2", Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(firstWaybill))
	require.NoError(t, waybillRepo.CreateYanwenWaybill(secondWaybill))

	updated, err := apiService.SyncYanwenWaybillsOfficialDetails(context.Background(), []uint{firstWaybill.ID, secondWaybill.ID})
	require.NoError(t, err)
	require.Len(t, updated, 2)
	require.Equal(t, firstWaybill.ID, updated[0].ID)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusInTransit, updated[0].OfficialStatus)
	require.True(t, updated[0].IsPrinted)
	require.Equal(t, secondWaybill.ID, updated[1].ID)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusCancelled, updated[1].OfficialStatus)
	require.Equal(t, 1, requestCount, "selected waybills should be queried in one official request")

	storedFirst, err := waybillRepo.FindYanwenWaybillByID(firstWaybill.ID)
	require.NoError(t, err)
	require.Equal(t, "REF-BATCH-1", storedFirst.ReferenceNumber)
	require.Equal(t, "YW-BATCH-1", storedFirst.YanwenOrderNumber)
	require.Equal(t, shipping.YanwenWaybillStatusCreated, storedFirst.Status)
	require.NotNil(t, storedFirst.LastOfficialSyncedAt)
	storedSecond, err := waybillRepo.FindYanwenWaybillByID(secondWaybill.ID)
	require.NoError(t, err)
	require.Equal(t, "REF-BATCH-2", storedSecond.ReferenceNumber)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusCancelled, storedSecond.OfficialStatus)
}

func TestYanwenAPIServiceRejectsMismatchedBatchOfficialDetailsWithoutLocalWrites(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":[{"waybillNumber":"YE-BATCH-MISMATCH","orderNumber":"WRONG-ORDER","referenceNumber":"REF-SHOULD-NOT-SAVE","yanwenOrderNumber":"YW-MISMATCH","status":"5","isPrint":"1"}]}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: "batch-mismatch-user", APIToken: "batch-mismatch-token", Enabled: true}))
	waybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 33, OrderNumber: "ORDER-BATCH-MISMATCH", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-BATCH-MISMATCH", Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(waybill))

	_, err = apiService.SyncYanwenWaybillsOfficialDetails(context.Background(), []uint{waybill.ID})
	require.ErrorContains(t, err, "does not match local")
	stored, err := waybillRepo.FindYanwenWaybillByID(waybill.ID)
	require.NoError(t, err)
	require.Empty(t, stored.ReferenceNumber)
	require.Equal(t, 0, stored.OfficialStatus)
	require.Nil(t, stored.LastOfficialSyncedAt)
}

func TestYanwenAPIServiceReturnsOfficialPDFLabelForOnePersistedWaybill(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	const userID = "label-service-user"
	const apiToken = "label-service-token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, yanwenGetOrderLabelMethod, request.URL.Query().Get("method"))
		_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-LABEL-SERVICE","isSuccess":true,"errorMsg":"","base64String":"JVBERi0xLjQ="}}`))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: userID, APIToken: apiToken, Enabled: true}))
	waybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 12, OrderNumber: "ORDER-LABEL", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-LABEL-SERVICE", YanwenOrderNumber: "YW-LABEL",
		Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(waybill))

	label, err := apiService.DownloadYanwenWaybillLabel(context.Background(), waybill.ID)
	require.NoError(t, err)
	require.Equal(t, waybill.ID, label.WaybillID)
	require.Equal(t, "YE-LABEL-SERVICE", label.WaybillNumber)
	require.Equal(t, "yanwen-YE-LABEL-SERVICE.pdf", label.FileName)
	require.Equal(t, "application/pdf", label.ContentType)
	require.Equal(t, "JVBERi0xLjQ=", label.Base64String)
}

func TestYanwenAPIServicePackagesBatchOfficialPDFLabelsWithManifest(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload struct {
			WaybillNumber string `json:"waybillNumber"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		if payload.WaybillNumber == "YE-BATCH-LABEL-FAILED" {
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-BATCH-LABEL-FAILED","isSuccess":false,"errorMsg":"label unavailable","base64String":""}}`))
			return
		}
		_, _ = writer.Write([]byte(fmt.Sprintf(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":%q,"isSuccess":true,"errorMsg":"","base64String":"JVBERi0xLjQ="}}`, payload.WaybillNumber)))
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: "batch-label-user", APIToken: "batch-label-token", Enabled: true}))

	createWaybill := func(environment, number string, orderID uint) *shipping.YanwenWaybill {
		waybill := &shipping.YanwenWaybill{
			Environment: environment, OrderID: orderID, OrderNumber: "ORDER-BATCH-LABEL", ProductCode: "481", ChannelName: "燕文普货",
			WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
			TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: number, Status: shipping.YanwenWaybillStatusCreated,
		}
		require.NoError(t, waybillRepo.CreateYanwenWaybill(waybill))
		return waybill
	}
	okWaybill := createWaybill("fat", "YE-BATCH-LABEL-OK", 51)
	failedWaybill := createWaybill("fat", "YE-BATCH-LABEL-FAILED", 52)
	productionWaybill := createWaybill("production", "YE-BATCH-LABEL-PRODUCTION", 53)

	archive, err := apiService.DownloadYanwenWaybillLabelsArchive(context.Background(), []uint{okWaybill.ID, failedWaybill.ID})
	require.NoError(t, err)
	require.Equal(t, "yanwen-waybill-labels.zip", archive.FileName)
	require.Equal(t, "application/zip", archive.ContentType)
	require.Equal(t, 1, archive.Summary.Succeeded)
	require.Equal(t, 1, archive.Summary.Failed)
	zipReader, err := zip.NewReader(bytes.NewReader(archive.Data), int64(len(archive.Data)))
	require.NoError(t, err)
	require.Len(t, zipReader.File, 2)
	manifestFound := false
	pdfFound := false
	for _, file := range zipReader.File {
		fileReader, openErr := file.Open()
		require.NoError(t, openErr)
		content, readErr := io.ReadAll(fileReader)
		_ = fileReader.Close()
		require.NoError(t, readErr)
		switch file.Name {
		case "yanwen-label-manifest.json":
			manifestFound = true
			require.Contains(t, string(content), "YE-BATCH-LABEL-FAILED")
			require.Contains(t, string(content), "label unavailable")
		default:
			pdfFound = true
			require.True(t, bytes.HasPrefix(content, []byte("%PDF")))
		}
	}
	require.True(t, manifestFound)
	require.True(t, pdfFound)

	_, err = apiService.DownloadYanwenWaybillLabelsArchive(context.Background(), []uint{okWaybill.ID, productionWaybill.ID})
	require.ErrorContains(t, err, "one environment")
}

func TestYanwenAPIServiceCancelsWaybillsAndOnlyReportsOfficialCancellationStatus(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	const userID = "cancel-service-user"
	const apiToken = "cancel-service-token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		method := request.URL.Query().Get("method")
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload struct {
			WaybillNumber string `json:"waybillNumber"`
			Note          string `json:"note"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		switch method {
		case yanwenCancelOrderMethod:
			require.NotEmpty(t, payload.WaybillNumber)
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":null}`))
		case yanwenGetOrderMethod:
			switch payload.WaybillNumber {
			case "YE-CANCEL-CONFIRMED":
				_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-CANCEL-CONFIRMED","orderNumber":"ORDER-CANCEL-CONFIRMED","referenceNumber":"","yanwenOrderNumber":"YW-CONFIRMED","status":"5","isPrint":"0"}}`))
			case "YE-CANCEL-PENDING":
				_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-CANCEL-PENDING","orderNumber":"ORDER-CANCEL-PENDING","referenceNumber":"","yanwenOrderNumber":"YW-PENDING","status":"3","isPrint":"0"}}`))
			default:
				t.Fatalf("unexpected waybill number %q", payload.WaybillNumber)
			}
		default:
			t.Fatalf("unexpected Yanwen method %q", method)
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: userID, APIToken: apiToken, Enabled: true}))

	confirmedWaybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 21, OrderNumber: "ORDER-CANCEL-CONFIRMED", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-CANCEL-CONFIRMED", YanwenOrderNumber: "YW-CREATE-1",
		Status: shipping.YanwenWaybillStatusCreated,
	}
	pendingWaybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: 22, OrderNumber: "ORDER-CANCEL-PENDING", ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-CANCEL-PENDING", YanwenOrderNumber: "YW-CREATE-2",
		Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(confirmedWaybill))
	require.NoError(t, waybillRepo.CreateYanwenWaybill(pendingWaybill))

	confirmedResult, err := apiService.CancelYanwenWaybill(context.Background(), confirmedWaybill.ID, "duplicate order")
	require.NoError(t, err)
	require.True(t, confirmedResult.CancellationAccepted)
	require.True(t, confirmedResult.OfficialStatusSynced)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusCancelled, confirmedResult.Waybill.OfficialStatus)
	require.Contains(t, confirmedResult.Message, "官方状态已同步为已取消")
	confirmedStored, err := waybillRepo.FindYanwenWaybillByID(confirmedWaybill.ID)
	require.NoError(t, err)
	require.Equal(t, shipping.YanwenWaybillStatusCreated, confirmedStored.Status, "local lifecycle status must not be fabricated by cancellation")

	pendingResult, err := apiService.CancelYanwenWaybill(context.Background(), pendingWaybill.ID, "operator review")
	require.NoError(t, err)
	require.True(t, pendingResult.CancellationAccepted)
	require.True(t, pendingResult.OfficialStatusSynced)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusInTransit, pendingResult.Waybill.OfficialStatus)
	require.Contains(t, pendingResult.Message, "尚未同步为已取消")
}

func TestYanwenAPIServiceBatchCancellationKeepsPerWaybillOutcomesAndEnvironmentIsolation(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenAPIConfig{}, &shipping.YanwenWaybill{}))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload struct {
			WaybillNumber string `json:"waybillNumber"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		switch request.URL.Query().Get("method") {
		case yanwenCancelOrderMethod:
			if payload.WaybillNumber == "YE-BATCH-CANCEL-FAILED" {
				_, _ = writer.Write([]byte(`{"success":false,"code":"CANCEL_400","message":"waybill cannot be cancelled","data":null}`))
				return
			}
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"操作成功","data":null}`))
		case yanwenGetOrderMethod:
			status := "3"
			if payload.WaybillNumber == "YE-BATCH-CANCEL-OK" {
				status = "5"
			}
			_, _ = writer.Write([]byte(fmt.Sprintf(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":%q,"orderNumber":"ORDER-BATCH-CANCEL","referenceNumber":"","yanwenOrderNumber":"YW-BATCH-CANCEL","status":%q,"isPrint":"0"}}`, payload.WaybillNumber, status)))
		default:
			t.Fatalf("unexpected Yanwen method %q", request.URL.Query().Get("method"))
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, nil)
	apiService.ConfigureYanwenWaybillRepositories(nil, waybillRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: "batch-cancel-user", APIToken: "batch-cancel-token", Enabled: true}))

	createWaybill := func(environment, number string, orderID uint) *shipping.YanwenWaybill {
		waybill := &shipping.YanwenWaybill{
			Environment: environment, OrderID: orderID, OrderNumber: "ORDER-BATCH-CANCEL", ProductCode: "481", ChannelName: "燕文普货",
			WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Spoke",
			TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: number, Status: shipping.YanwenWaybillStatusCreated,
		}
		require.NoError(t, waybillRepo.CreateYanwenWaybill(waybill))
		return waybill
	}
	okWaybill := createWaybill("fat", "YE-BATCH-CANCEL-OK", 41)
	failedWaybill := createWaybill("fat", "YE-BATCH-CANCEL-FAILED", 42)
	pendingWaybill := createWaybill("fat", "YE-BATCH-CANCEL-PENDING", 43)
	productionWaybill := createWaybill("production", "YE-BATCH-CANCEL-PRODUCTION", 44)

	result, err := apiService.CancelYanwenWaybills(context.Background(), []uint{okWaybill.ID, failedWaybill.ID, pendingWaybill.ID}, "operator batch review")
	require.NoError(t, err)
	require.Equal(t, 2, result.Succeeded)
	require.Equal(t, 1, result.Failed)
	require.Len(t, result.Items, 3)
	require.True(t, result.Items[0].CancellationAccepted)
	require.True(t, result.Items[0].OfficialStatusSynced)
	require.Equal(t, "YE-BATCH-CANCEL-FAILED", result.Items[1].WaybillNumber)
	require.Contains(t, result.Items[1].Error, "waybill cannot be cancelled")
	require.False(t, result.Items[1].CancellationAccepted)
	require.True(t, result.Items[2].CancellationAccepted)
	require.True(t, result.Items[2].OfficialStatusSynced)

	storedOK, err := waybillRepo.FindYanwenWaybillByID(okWaybill.ID)
	require.NoError(t, err)
	require.Equal(t, shipping.YanwenOfficialWaybillStatusCancelled, storedOK.OfficialStatus)
	storedFailed, err := waybillRepo.FindYanwenWaybillByID(failedWaybill.ID)
	require.NoError(t, err)
	require.Zero(t, storedFailed.OfficialStatus)

	_, err = apiService.CancelYanwenWaybills(context.Background(), []uint{okWaybill.ID, productionWaybill.ID}, "mixed environments")
	require.ErrorContains(t, err, "one environment")
}

func TestYanwenAPIServiceBatchCreationContinuesAfterOneOfficialFailureAndKeepsExistingWaybillIdempotent(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&order.Order{}, &order.OrderItem{}, &shipping.YanwenAPIConfig{},
		&shipping.YanwenProductCatalogEntry{}, &shipping.YanwenCountryCatalogEntry{},
		&shipping.YanwenWarehouseCatalogEntry{}, &shipping.YanwenPublishedChannel{},
		&shipping.YanwenWaybill{},
	))

	createRequestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		require.NoError(t, readErr)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		orderNumber, _ := payload["orderNumber"].(string)
		switch request.URL.Query().Get("method") {
		case yanwenVerifyUnitedStatesAddressMethod:
			_, _ = writer.Write([]byte(`{"success":true,"code":"0","message":"address ok","data":{"receiverInfo":{"address":"1 Main Street","city":"Austin","state":"TX","zipCode4":"7870","zipCode5":"78701"}}}`))
		case yanwenCreateOrderMethod:
			createRequestCount++
			if orderNumber == "ORDER-BATCH-FAIL" {
				_, _ = writer.Write([]byte(`{"success":false,"code":"CREATE_400","message":"official create rejected","data":null}`))
				return
			}
			_, _ = writer.Write([]byte(fmt.Sprintf(`{"success":true,"code":"0","message":"ok","data":{"waybillNumber":"YE-%s","orderNumber":%q,"yanwenOrderNumber":"YW-%s"}}`, strings.TrimPrefix(orderNumber, "ORDER-"), orderNumber, strings.TrimPrefix(orderNumber, "ORDER-"))))
		default:
			t.Fatalf("unexpected Yanwen method %q", request.URL.Query().Get("method"))
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)

	configRepo := repository.NewYanwenAPIConfigRepository(db)
	productRepo := repository.NewYanwenProductCatalogRepository(db)
	countryRepo := repository.NewYanwenCountryCatalogRepository(db)
	warehouseRepo := repository.NewYanwenWarehouseCatalogRepository(db)
	publishedChannelRepo := repository.NewYanwenPublishedChannelRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	waybillRepo := repository.NewYanwenWaybillRepository(db)
	apiService := NewYanwenAPIService(configRepo, productRepo)
	apiService.ConfigureYanwenCountryCatalogRepository(countryRepo)
	apiService.ConfigureYanwenWarehouseCatalogRepository(warehouseRepo)
	apiService.ConfigureYanwenWaybillRepositories(orderRepo, waybillRepo)
	apiService.ConfigureYanwenPublishedChannelRepository(publishedChannelRepo)
	apiService.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, apiService.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: "batch-create-user", APIToken: "batch-create-token", Enabled: true}))
	now := time.Now().UTC()
	_, err = productRepo.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{{ProductID: "481", NameChinese: "普货", NameEnglish: "Tracked"}}, now)
	require.NoError(t, err)
	_, err = countryRepo.UpsertYanwenCountryCatalogEntries("fat", []shipping.YanwenCountryCatalogEntry{{CountryID: "840", CountryCode: "US", NameEnglish: "United States"}}, now)
	require.NoError(t, err)
	_, err = warehouseRepo.UpsertYanwenWarehouseCatalogEntries("fat", []shipping.YanwenWarehouseCatalogEntry{{WarehouseCode: "WH01", Name: "华东仓"}}, now)
	require.NoError(t, err)
	channel := shipping.YanwenPublishedChannel{Environment: "fat", ProductCode: "481", DisplayName: "燕文普货", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000}
	require.NoError(t, publishedChannelRepo.CreateYanwenPublishedChannel(&channel))

	declaredValue := int64(1200)
	variantID := uint(1)
	createOrder := func(orderNumber string) *order.Order {
		orderRecord := &order.Order{
			OrderNumber: orderNumber, Status: "processing", PaymentStatus: "paid", Currency: "USD", PaymentCurrency: "USD", PaymentAmountMinor: 1200,
			ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "US", Phone: "123", Email: "rider@example.com"},
			Items:           []order.OrderItem{{ProductID: 1, VariantID: &variantID, ProductName: "Spoke", CustomsDescription: "Bicycle spoke", Quantity: 1, Currency: "USD", PriceMinor: 1200, SubtotalMinor: 1200, TotalMinor: 1200, WeightGrams: 100, DeclaredValueMinor: &declaredValue, DeclaredValueConfirmed: true}},
		}
		require.NoError(t, orderRepo.Create(orderRecord))
		return orderRecord
	}
	firstOrder := createOrder("ORDER-BATCH-FIRST")
	failingOrder := createOrder("ORDER-BATCH-FAIL")
	existingOrder := createOrder("ORDER-BATCH-EXISTING")
	existingWaybill := &shipping.YanwenWaybill{
		Environment: "fat", OrderID: existingOrder.ID, OrderNumber: existingOrder.OrderNumber, ProductCode: "481", ChannelName: "燕文普货",
		WarehouseCode: "WH01", DestinationCountry: "US", ConsigneeName: "Rider Example", DeclaredDescription: "Bicycle spoke",
		TotalQuantity: 1, TotalWeightGrams: 100, WaybillNumber: "YE-EXISTING", YanwenOrderNumber: "YW-EXISTING", Status: shipping.YanwenWaybillStatusCreated,
	}
	require.NoError(t, waybillRepo.CreateYanwenWaybill(existingWaybill))

	hasBattery := false
	result, err := apiService.CreateYanwenWaybills(context.Background(), []YanwenCreateWaybillInput{
		{Environment: "fat", OrderID: firstOrder.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
		{Environment: "fat", OrderID: failingOrder.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
		{Environment: "fat", OrderID: existingOrder.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Succeeded)
	require.Equal(t, 1, result.Failed)
	require.Len(t, result.Items, 3)
	require.NotZero(t, result.Items[0].WaybillID)
	require.Contains(t, result.Items[1].Error, "official create rejected")
	require.Equal(t, existingWaybill.ID, result.Items[2].WaybillID)
	require.Equal(t, 2, createRequestCount, "the existing local waybill must not call express.order.create")

	_, err = apiService.CreateYanwenWaybills(context.Background(), []YanwenCreateWaybillInput{
		{Environment: "fat", OrderID: firstOrder.ID + 100, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
		{Environment: "production", OrderID: firstOrder.ID + 101, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
	})
	require.ErrorContains(t, err, "one environment")
	require.Equal(t, 2, createRequestCount, "mixed environments must be rejected before any official request")

	_, err = apiService.CreateYanwenWaybills(context.Background(), []YanwenCreateWaybillInput{
		{Environment: "fat", OrderID: firstOrder.ID + 200, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
		{Environment: "fat", OrderID: firstOrder.ID + 200, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery},
	})
	require.ErrorContains(t, err, "duplicate order_id")
	require.Equal(t, 2, createRequestCount, "duplicate order and channel must be rejected before any official request")
}

func TestBuildYanwenCreateOrderRequestRequiresConfirmedDeclaredValue(t *testing.T) {
	orderRecord := &order.Order{
		OrderNumber: "ORDER-1", Currency: "USD",
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street"},
		Items:           []order.OrderItem{{ProductName: "Spoke", Quantity: 1, WeightGrams: 100}},
	}
	_, _, _, _, err := buildYanwenCreateOrderRequest(orderRecord, "481", "WH01", "", false, "US")
	require.ErrorContains(t, err, "confirmed positive declared value")
}

func TestBuildYanwenCreateOrderRequestUsesCustomsDescriptionForEnglishDeclaration(t *testing.T) {
	declaredValue := int64(4500)
	orderRecord := &order.Order{
		OrderNumber: "ORDER-CUSTOMS-DESCRIPTION", Currency: "USD",
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street"},
		Items: []order.OrderItem{{
			ProductName:            "Tanzanite 极光 45mm 碟刹碳纤维公路轮组 前轮",
			CustomsDescription:     "  Bicycle wheelset  ",
			Quantity:               1,
			WeightGrams:            100,
			DeclaredValueMinor:     &declaredValue,
			DeclaredValueConfirmed: true,
		}},
	}

	request, description, _, _, err := buildYanwenCreateOrderRequest(orderRecord, "481", "WH01", "", false, "US")
	require.NoError(t, err)
	require.Len(t, request.ParcelInfo.ProductList, 1)
	require.Equal(t, orderRecord.Items[0].ProductName, request.ParcelInfo.ProductList[0].GoodsNameChinese)
	require.Equal(t, "Bicycle wheelset", request.ParcelInfo.ProductList[0].GoodsNameEnglish)
	require.Equal(t, "Bicycle wheelset", description)
}

func TestBuildYanwenCreateOrderRequestMapsYanwenCustomsIdentifiersToOfficialFields(t *testing.T) {
	declaredValue := int64(4500)
	orderRecord := &order.Order{
		OrderNumber: "ORDER-YANWEN-CUSTOMS-FIELDS", Currency: "EUR",
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street"},
		Items: []order.OrderItem{{
			ProductName: "车轮辐条", CustomsDescription: "Bicycle spoke", Quantity: 1, WeightGrams: 100,
			DeclaredValueMinor: &declaredValue, DeclaredValueConfirmed: true,
		}},
	}

	request, _, _, _, err := buildYanwenCreateOrderRequestWithCustomsDeclaration(
		orderRecord,
		"481",
		"WH01",
		"",
		false,
		"DE",
		YanwenCustomsDeclarationInput{
			ReceiverTaxNumber: "  DE123456789  ",
			IOSS:              " IM1234567890 ",
			EORI:              " DE987654321000 ",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "DE123456789", request.ReceiverInfo.TaxNumber)
	require.Equal(t, "IM1234567890", request.ParcelInfo.IOSS)
	require.NotNil(t, request.ImportCustomsInfo)
	require.Equal(t, "DE987654321000", request.ImportCustomsInfo.EORI)

	encodedRequest, err := json.Marshal(request)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"channelId":"481",
		"orderSource":"tanzanite-theme",
		"orderNumber":"ORDER-YANWEN-CUSTOMS-FIELDS",
		"companyCode":"WH01",
		"receiverInfo":{"name":"Rider Example","country":"DE","address":"1 Main Street","taxNumber":"DE123456789"},
		"parcelInfo":{"hasBattery":0,"currency":"EUR","totalQuantity":1,"totalWeight":100,"ioss":"IM1234567890","productList":[{"goodsNameCh":"车轮辐条","goodsNameEn":"Bicycle spoke","price":45.00,"priceExport":45.00,"quantity":1,"weight":100}]},
		"importCustomsInfo":{"eori":"DE987654321000"},
		"salesPlatform":"tanzanite-theme"
	}`, string(encodedRequest))
}

func TestBuildYanwenCreateOrderRequestRejectsCustomsIdentifiersBeyondOfficialLimits(t *testing.T) {
	declaredValue := int64(4500)
	orderRecord := &order.Order{
		OrderNumber: "ORDER-YANWEN-CUSTOMS-LIMIT", Currency: "EUR",
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street"},
		Items: []order.OrderItem{{
			ProductName: "车轮辐条", CustomsDescription: "Bicycle spoke", Quantity: 1, WeightGrams: 100,
			DeclaredValueMinor: &declaredValue, DeclaredValueConfirmed: true,
		}},
	}
	tooLongEORI := strings.Repeat("E", maximumYanwenEORILength+1)
	_, _, _, _, err := buildYanwenCreateOrderRequestWithCustomsDeclaration(
		orderRecord, "481", "WH01", "", false, "DE", YanwenCustomsDeclarationInput{EORI: tooLongEORI},
	)
	require.ErrorContains(t, err, "EORI cannot exceed 64 characters")
}

func TestBuildYanwenCreateOrderRequestRejectsMissingCustomsDescription(t *testing.T) {
	declaredValue := int64(1200)
	orderRecord := &order.Order{
		OrderNumber: "ORDER-CUSTOMS-REQUIRED", Currency: "USD",
		ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Main Street"},
		Items: []order.OrderItem{{
			ProductName:            "Tanzanite 营销商品名",
			CustomsDescription:     "   ",
			Quantity:               1,
			WeightGrams:            100,
			DeclaredValueMinor:     &declaredValue,
			DeclaredValueConfirmed: true,
		}},
	}

	_, _, _, _, err := buildYanwenCreateOrderRequest(orderRecord, "481", "WH01", "", false, "US")
	require.ErrorContains(t, err, "customs description is required before Yanwen declaration")
}

func newYanwenCreateOrderTestServer(t *testing.T, userID, apiToken, responseBody string) *httptest.Server {
	return httpServerFunc(t, func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		raw := userID + string(body) + "json" + yanwenCreateOrderMethod + request.URL.Query().Get("timestamp") + yanwenAPIVersion
		require.Equal(t, signYanwenRequest(apiToken, raw), request.URL.Query().Get("sign"))
		_, _ = writer.Write([]byte(responseBody))
	})
}

func httpServerFunc(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}
