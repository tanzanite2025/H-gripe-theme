package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductBrandServiceUpdateQueuesDependentCacheAndMerchantEvents(t *testing.T) {
	db := newProductBrandServiceTestDB(t)
	brand := product.ProductBrand{
		Name:      "DT Swiss",
		Slug:      "dt-swiss",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&brand).Error)

	activeBrandID := brand.ID
	require.NoError(t, db.Create(&product.Product{
		BrandID: &activeBrandID,
		SKU:     "DT-ACTIVE",
		Name:    "Active Rim",
		Slug:    "active-rim",
		Status:  "active",
	}).Error)
	require.NoError(t, db.Create(&product.Product{
		BrandID: &activeBrandID,
		SKU:     "DT-INACTIVE",
		Name:    "Inactive Rim",
		Slug:    "inactive-rim",
		Status:  "inactive",
	}).Error)
	require.NoError(t, db.Create(&product.Product{
		BrandID: &activeBrandID,
		SKU:     "DT-OOS",
		Name:    "Out of Stock Rim",
		Slug:    "out-of-stock-rim",
		Status:  "out_of_stock",
	}).Error)

	cacheEvents := &recordingProductCacheEventPublisher{}
	merchantEvents := &recordingMerchantProductEventPublisher{}
	service := NewProductBrandService(repository.NewProductBrandRepository(db))
	service.ConfigureProductDependencies(
		repository.NewProductRepository(db),
		cacheEvents,
		merchantEvents,
	)

	updated, err := service.Update(brand.ID, ProductBrandInput{
		Name:        "DT Swiss Updated",
		Slug:        "dt-swiss-updated",
		Description: "Updated brand description",
		IsEnabled:   true,
	})

	require.NoError(t, err)
	require.Equal(t, "DT Swiss Updated", updated.Name)
	require.Equal(t, []uint{brand.ID}, cacheEvents.brandIDs)
	require.Equal(t, []string{"product_brand_changed"}, cacheEvents.reasons)
	require.Equal(t, []merchantProductEvent{
		{productID: 1, eventType: "upsert", reason: "product_brand_changed"},
		{productID: 2, eventType: "withdraw", reason: "product_brand_changed"},
		{productID: 3, eventType: "withdraw", reason: "product_brand_changed"},
	}, merchantEvents.events)
}

func TestProductBrandServiceUpdateSortOrderDoesNotInvalidateProductPayload(t *testing.T) {
	db := newProductBrandServiceTestDB(t)
	brand := product.ProductBrand{
		Name:      "Zipp",
		Slug:      "zipp",
		IsEnabled: true,
		SortOrder: 10,
	}
	require.NoError(t, db.Create(&brand).Error)

	cacheEvents := &recordingProductCacheEventPublisher{}
	merchantEvents := &recordingMerchantProductEventPublisher{}
	service := NewProductBrandService(repository.NewProductBrandRepository(db))
	service.ConfigureProductDependencies(
		repository.NewProductRepository(db),
		cacheEvents,
		merchantEvents,
	)

	updated, err := service.Update(brand.ID, ProductBrandInput{
		Name:      brand.Name,
		Slug:      brand.Slug,
		IsEnabled: brand.IsEnabled,
		SortOrder: 20,
	})

	require.NoError(t, err)
	require.Equal(t, 20, updated.SortOrder)
	require.Empty(t, cacheEvents.brandIDs)
	require.Empty(t, merchantEvents.events)
}

type recordingProductCacheEventPublisher struct {
	brandIDs []uint
	reasons  []string
}

func (p *recordingProductCacheEventPublisher) EnqueueProductCacheInvalidateByIDs([]uint, string) error {
	return nil
}

func (p *recordingProductCacheEventPublisher) EnqueueProductCacheInvalidateByProductSpecificationTemplateID(uint, string) error {
	return nil
}

func (p *recordingProductCacheEventPublisher) EnqueueProductCacheInvalidateByBrandID(brandID uint, reason string) error {
	p.brandIDs = append(p.brandIDs, brandID)
	p.reasons = append(p.reasons, reason)
	return nil
}

func (p *recordingProductCacheEventPublisher) EnqueueProductCacheInvalidateByInformationTemplateID(uint, string) error {
	return nil
}

type merchantProductEvent struct {
	productID uint
	eventType string
	reason    string
}

type recordingMerchantProductEventPublisher struct {
	events []merchantProductEvent
}

func (p *recordingMerchantProductEventPublisher) EnqueueProductUpsert(productID uint, reason string) error {
	p.events = append(p.events, merchantProductEvent{
		productID: productID,
		eventType: "upsert",
		reason:    reason,
	})
	return nil
}

func (p *recordingMerchantProductEventPublisher) EnqueueProductWithdraw(productID uint, reason string) error {
	p.events = append(p.events, merchantProductEvent{
		productID: productID,
		eventType: "withdraw",
		reason:    reason,
	})
	return nil
}

func newProductBrandServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

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

	require.NoError(t, db.AutoMigrate(&product.ProductBrand{}, &product.Product{}))
	return db
}
