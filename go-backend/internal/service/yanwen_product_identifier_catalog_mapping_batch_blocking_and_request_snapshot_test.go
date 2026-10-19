package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenProductIdentifiersReadCatalogBlockIncompleteBatchItemsAndPersistActualRequest(t *testing.T) {
	t.Setenv(YanwenAPIMasterKeyEnv, "test-yanwen-master-key")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&product.Product{}, &order.Order{}, &order.OrderItem{},
		&shipping.YanwenAPIConfig{}, &shipping.YanwenProductCatalogEntry{}, &shipping.YanwenCountryCatalogEntry{},
		&shipping.YanwenWarehouseCatalogEntry{}, &shipping.YanwenPublishedChannel{}, &shipping.YanwenWaybill{}))

	var submitted []YanwenCreateOrderRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, yanwenCreateOrderMethod, r.URL.Query().Get("method"))
		var request YanwenCreateOrderRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		submitted = append(submitted, request)
		fmt.Fprintf(w, `{"success":true,"code":"0","data":{"waybillNumber":"YE-%s","orderNumber":"%s"}}`, request.OrderNumber, request.OrderNumber)
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	products := repository.NewYanwenProductCatalogRepository(db)
	countries := repository.NewYanwenCountryCatalogRepository(db)
	warehouses := repository.NewYanwenWarehouseCatalogRepository(db)
	channels := repository.NewYanwenPublishedChannelRepository(db)
	orders := repository.NewOrderRepository(db)
	waybills := repository.NewYanwenWaybillRepository(db)
	svc := NewYanwenAPIService(repository.NewYanwenAPIConfigRepository(db), products)
	svc.ConfigureYanwenCountryCatalogRepository(countries)
	svc.ConfigureYanwenWarehouseCatalogRepository(warehouses)
	svc.ConfigureYanwenWaybillRepositories(orders, waybills)
	svc.ConfigureYanwenPublishedChannelRepository(channels)
	svc.gateway.httpClient = &http.Client{Transport: yanwenTestRoundTripper{target: target}}
	require.NoError(t, svc.SaveYanwenAPIConfiguration(YanwenAPIConfigInput{Environment: "fat", Endpoint: yanwenFATEndpoint, UserID: "user", APIToken: "token", Enabled: true}))
	now := time.Now().UTC()
	_, err = products.UpsertYanwenProductCatalogEntries("fat", []shipping.YanwenProductCatalogEntry{{ProductID: "481", NameEnglish: "Tracked"}}, now)
	require.NoError(t, err)
	_, err = countries.UpsertYanwenCountryCatalogEntries("fat", []shipping.YanwenCountryCatalogEntry{{CountryID: "276", CountryCode: "DE", NameEnglish: "Germany"}}, now)
	require.NoError(t, err)
	_, err = warehouses.UpsertYanwenWarehouseCatalogEntries("fat", []shipping.YanwenWarehouseCatalogEntry{{WarehouseCode: "WH01", Name: "Warehouse"}}, now)
	require.NoError(t, err)
	channel := shipping.YanwenPublishedChannel{Environment: "fat", ProductCode: "481", DisplayName: "EU", Countries: `["DE"]`, Enabled: true, VolumetricDivisor: 8000}
	require.NoError(t, channels.CreateYanwenPublishedChannel(&channel))
	first := product.Product{Name: "Spoke", Slug: "spoke", MerchantProductIdentifier: " SHOP-SPOKE ", ManufacturerProductIdentifier: " FACTORY-SPOKE ", StandardProductIdentifier: "00012345678905"}
	second := product.Product{Name: "Nipple", Slug: "nipple", MerchantProductIdentifier: "SHOP-NIPPLE", ManufacturerProductIdentifier: "FACTORY-NIPPLE"}
	incomplete := product.Product{Name: "Tool", Slug: "tool", MerchantProductIdentifier: "SHOP-TOOL"}
	for _, record := range []*product.Product{&first, &second, &incomplete} {
		require.NoError(t, db.Create(record).Error)
	}
	createOrderWithCatalogProducts := func(number string, catalogProducts ...product.Product) *order.Order {
		declared := int64(1200)
		variantID := uint(1)
		record := &order.Order{OrderNumber: number, Status: "processing", PaymentStatus: "paid", Currency: "USD", PaymentCurrency: "USD", PaymentAmountMinor: 1200,
			ShippingAddress: order.Address{FirstName: "Rider", LastName: "Example", Address1: "1 Street", Country: "DE"}}
		for _, catalogProduct := range catalogProducts {
			record.Items = append(record.Items, order.OrderItem{ProductID: catalogProduct.ID, VariantID: &variantID, ProductName: catalogProduct.Name, SKU: "SKU-" + catalogProduct.Slug,
				CustomsDescription: "Bicycle parts", Quantity: 1, Currency: "USD", PriceMinor: 1200, SubtotalMinor: 1200, TotalMinor: 1200, WeightGrams: 100, DeclaredValueMinor: &declared, DeclaredValueConfirmed: true})
		}
		require.NoError(t, orders.Create(record))
		return record
	}
	validOrder := createOrderWithCatalogProducts("VALID", second, first)
	invalidOrder := createOrderWithCatalogProducts("INCOMPLETE", first, incomplete)
	hasBattery := false
	createInputForOrder := func(record *order.Order) YanwenCreateWaybillInput {
		return YanwenCreateWaybillInput{Environment: "fat", OrderID: record.ID, ChannelID: channel.ID, WarehouseCode: "WH01", HasBattery: &hasBattery}
	}
	_, err = svc.CreateYanwenWaybill(context.Background(), createInputForOrder(invalidOrder))
	require.ErrorContains(t, err, "NSPID")
	require.ErrorContains(t, err, "SKU-tool")
	require.Empty(t, submitted)
	batch, err := svc.CreateYanwenWaybills(context.Background(), []YanwenCreateWaybillInput{createInputForOrder(invalidOrder), createInputForOrder(validOrder)})
	require.NoError(t, err)
	require.Equal(t, 1, batch.Succeeded)
	require.Equal(t, 1, batch.Failed)
	require.Zero(t, batch.Items[0].WaybillID)
	require.Contains(t, batch.Items[0].Error, "NSPID")
	require.Len(t, submitted, 1)
	require.Equal(t, "SHOP-NIPPLE", submitted[0].ParcelInfo.ProductList[0].MerchantProductIdentifier)
	require.Equal(t, "FACTORY-NIPPLE", submitted[0].ParcelInfo.ProductList[0].ManufacturerProductIdentifier)
	require.Empty(t, submitted[0].ParcelInfo.ProductList[0].StandardProductIdentifier)
	require.Equal(t, "SHOP-SPOKE", submitted[0].ParcelInfo.ProductList[1].MerchantProductIdentifier)
	require.Equal(t, "00012345678905", submitted[0].ParcelInfo.ProductList[1].StandardProductIdentifier)
	var stored shipping.YanwenWaybill
	require.NoError(t, db.First(&stored, batch.Items[1].WaybillID).Error)
	originalSnapshot := string(stored.RequestData)
	require.Contains(t, originalSnapshot, `"mpid":"SHOP-NIPPLE"`)
	require.Contains(t, originalSnapshot, `"nspid":"FACTORY-SPOKE"`)
	var persistedCount int64
	require.NoError(t, db.Model(&shipping.YanwenWaybill{}).Count(&persistedCount).Error)
	require.EqualValues(t, 1, persistedCount)

	// Correct catalog data, then retry the previously blocked order.
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", incomplete.ID).Update("manufacturer_product_identifier", "FACTORY-TOOL").Error)
	_, err = svc.CreateYanwenWaybill(context.Background(), createInputForOrder(invalidOrder))
	require.NoError(t, err)
	require.Len(t, submitted, 2)
	require.Equal(t, "FACTORY-TOOL", submitted[1].ParcelInfo.ProductList[1].ManufacturerProductIdentifier)
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", first.ID).Update("merchant_product_identifier", "UPDATED-SHOP").Error)
	repeated, err := svc.CreateYanwenWaybill(context.Background(), createInputForOrder(validOrder))
	require.NoError(t, err)
	require.Len(t, submitted, 2)
	require.Equal(t, originalSnapshot, string(repeated.RequestData))

	// A removed catalog product cannot silently supply an obsolete identifier.
	removedOrder := createOrderWithCatalogProducts("REMOVED", second)
	require.NoError(t, db.Delete(&second).Error)
	_, err = svc.CreateYanwenWaybill(context.Background(), createInputForOrder(removedOrder))
	require.ErrorContains(t, err, "MPID")
	require.Len(t, submitted, 2)
}
