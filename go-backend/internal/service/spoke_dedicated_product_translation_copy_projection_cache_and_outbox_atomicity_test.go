package service

import (
	"encoding/json"
	"testing"

	outboxdomain "commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSpokeProductTranslationCopiesAndSynchronizesRootFactsAndInvalidatesTranslationCaches(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	sourceSpecification := &product.SpokesProductSpecification{
		SchemaVersion:      1,
		ModelName:          "CX-Ray",
		ManufacturerFamily: "Aero spokes",
	}
	source, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: sourceSpecification,
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-translation-regression",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)
	configureSpokeProductTranslationTransactionalOutboxForTest(t, db, productService)

	translated, group, err := productService.CopyAdminProductTranslation(source.ID, "fr")
	require.NoError(t, err)
	require.NotNil(t, translated.ParentID)
	assert.Equal(t, source.ID, *translated.ParentID)
	assert.Equal(t, product.ProductTypeSpoke, translated.ProductType)
	assert.Equal(t, sourceSpecification, translated.SpokeSpecification)
	assert.Equal(t, source.ProductCategoryID, translated.ProductCategoryID)
	assert.Len(t, group.Translations, 2)

	cacheIdentityRepository := &fakeProductCacheIdentityRepository{
		productsByID: map[uint]product.Product{
			source.ID:     {ID: source.ID, Slug: source.Slug, Locale: source.Locale},
			translated.ID: {ID: translated.ID, Slug: translated.Slug, Locale: translated.Locale},
		},
	}
	productDetailCache := &recordingProductDetailCache{}
	productService.productCacheInvalidator = NewProductDetailCacheInvalidator(cacheIdentityRepository, productDetailCache)

	updatedSpecification := *sourceSpecification
	updatedSpecification.ManufacturerFamily = "Updated SAPIM aero family"
	_, err = productService.UpdateAdminProduct(source.ID, ProductUpdateInput{
		UpdateSpokeSpecification: true,
		SpokeSpecification:       &updatedSpecification,
	})
	require.NoError(t, err)

	translatedAfterSourceUpdate, err := productService.GetAdminProduct(translated.ID)
	require.NoError(t, err)
	assert.Equal(t, &updatedSpecification, translatedAfterSourceUpdate.SpokeSpecification)

	var persistedTranslatedSpecificationJSON string
	require.NoError(t, db.Raw("SELECT spoke_specification FROM products WHERE id = ?", translated.ID).Scan(&persistedTranslatedSpecificationJSON).Error)
	var persistedTranslatedSpecification product.SpokesProductSpecification
	require.NoError(t, json.Unmarshal([]byte(persistedTranslatedSpecificationJSON), &persistedTranslatedSpecification))
	assert.Equal(t, updatedSpecification, persistedTranslatedSpecification)

	assert.Contains(t, cacheIdentityRepository.requestedIDs, translated.ID)
	assert.Contains(t, productDetailCache.deletedKeys, productSlugCacheKey(translated.Slug, translated.Locale))

	var cacheEvents []outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeProductCacheInvalidate).Find(&cacheEvents).Error)
	var foundTranslationGroupCacheEvent bool
	for _, event := range cacheEvents {
		var payload outboxdomain.ProductCacheInvalidatePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.Reason != "admin product update" {
			continue
		}
		require.ElementsMatch(t, []uint{source.ID, translated.ID}, payload.ProductIDs)
		foundTranslationGroupCacheEvent = true
	}
	assert.True(t, foundTranslationGroupCacheEvent, "source engineering fact updates must durably invalidate all translation product caches")

	var merchantEvents []outboxdomain.Event
	require.NoError(t, db.Where("event_type IN ?", []string{
		outboxdomain.EventTypeMerchantProductUpsert,
		outboxdomain.EventTypeMerchantProductWithdraw,
	}).Find(&merchantEvents).Error)
	var spokeSpecificationMerchantProductIDs []uint
	for _, event := range merchantEvents {
		var payload outboxdomain.MerchantProductSyncPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.Reason != "product_spoke_specification_changed" {
			continue
		}
		assert.Equal(t, outboxdomain.EventTypeMerchantProductWithdraw, event.EventType, "inactive translations must be withdrawn from Merchant")
		spokeSpecificationMerchantProductIDs = append(spokeSpecificationMerchantProductIDs, payload.ProductID)
	}
	assert.ElementsMatch(t, []uint{source.ID, translated.ID}, spokeSpecificationMerchantProductIDs,
		"shared spoke facts must notify Merchant for every product in the translation group")

	independentTranslatedSpecification := *translatedAfterSourceUpdate.SpokeSpecification
	independentTranslatedSpecification.ModelName = "Incorrect translation-specific spoke model"
	_, err = productService.UpdateAdminProduct(translated.ID, ProductUpdateInput{
		UpdateSpokeSpecification: true,
		SpokeSpecification:       &independentTranslatedSpecification,
	})
	require.ErrorIs(t, err, ErrProductSpecInvalid)
}

func TestSpokeTranslationCopyRequiresTransactionalOutboxBeforeCreatingProduct(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	source, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray"},
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-translation-requires-outbox",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)

	_, _, err = productService.CopyAdminProductTranslation(source.ID, "fr")
	require.ErrorIs(t, err, ErrTransactionalOutboxNotConfigured)

	var translatedProductCount int64
	require.NoError(t, db.Model(&product.Product{}).Where("parent_id = ? AND locale = ?", source.ID, "fr").Count(&translatedProductCount).Error)
	assert.Zero(t, translatedProductCount, "a spoke translation must not be created without atomic Merchant outbox support")
}

func TestSpokeTranslationRepositoryUsesCurrentRootFactsInsteadOfStaleSourceSnapshot(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	staleSourceSnapshot, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray", ManufacturerFamily: "Original family"},
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-stale-translation-snapshot",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)
	configureSpokeProductTranslationTransactionalOutboxForTest(t, db, productService)
	currentSpecification := &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray", ManufacturerFamily: "Updated family"}
	_, err = productService.UpdateAdminProduct(staleSourceSnapshot.ID, ProductUpdateInput{
		UpdateSpokeSpecification: true,
		SpokeSpecification:       currentSpecification,
	})
	require.NoError(t, err)

	rootID := staleSourceSnapshot.ID
	translated := &product.Product{
		ProductType:        product.ProductTypeSpoke,
		SpokeSpecification: staleSourceSnapshot.SpokeSpecification,
		ProductCategoryID:  staleSourceSnapshot.ProductCategoryID,
		ParentID:           &rootID,
		Name:               "SAPIM CX-Ray Spoke French",
		Slug:               "sapim-cx-ray-spoke-french-stale-snapshot",
		Status:             "inactive",
		Locale:             "fr",
		Currency:           staleSourceSnapshot.Currency,
	}
	transactionalProductRepository := repository.NewProductRepository(db)
	require.NoError(t, transactionalProductRepository.CreateTranslatedCopy(staleSourceSnapshot, translated, nil))

	var persistedSpecificationJSON string
	require.NoError(t, db.Raw("SELECT spoke_specification FROM products WHERE id = ?", translated.ID).Scan(&persistedSpecificationJSON).Error)
	var persistedSpecification product.SpokesProductSpecification
	require.NoError(t, json.Unmarshal([]byte(persistedSpecificationJSON), &persistedSpecification))
	assert.Equal(t, *currentSpecification, persistedSpecification,
		"translation persistence must re-read canonical root facts inside its write transaction")
}

func TestSpokeSpecificationUpdateRequiresTransactionalOutboxBeforeChangingProduct(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	originalSpecification := &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray"}
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: originalSpecification,
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-update-requires-outbox",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)
	updatedSpecification := &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray changed"}

	_, err = productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateSpokeSpecification: true,
		SpokeSpecification:       updatedSpecification,
	})
	require.ErrorIs(t, err, ErrTransactionalOutboxNotConfigured)

	unchanged, err := productService.GetAdminProduct(created.ID)
	require.NoError(t, err)
	assert.Equal(t, originalSpecification, unchanged.SpokeSpecification)
}

func TestSpokeTranslationGroupMerchantChangesUseEachProductsPublicationStatus(t *testing.T) {
	publisher := &recordingMerchantProductEventPublisher{}
	translations := []product.ProductTranslation{
		{ID: 21, Status: "active"},
		{ID: 22, Status: "inactive"},
	}

	require.NoError(t, enqueueMerchantProductTranslationGroupChangesWithPublisher(
		publisher,
		translations,
		"product_spoke_specification_changed",
	))
	assert.Equal(t, []merchantProductEvent{
		{productID: 21, eventType: "upsert", reason: "product_spoke_specification_changed"},
		{productID: 22, eventType: "withdraw", reason: "product_spoke_specification_changed"},
	}, publisher.events)
}

func TestSpokeProductTranslationCopyRollsBackWhenWithdrawalOutboxWriteFails(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	source, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray"},
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-outbox-rollback",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)
	configureSpokeProductTranslationTransactionalOutboxForTest(t, db, productService)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER reject_spoke_product_translation_withdrawal_event
		BEFORE INSERT ON outbox_events
		WHEN NEW.event_type = 'merchant.product_withdraw'
		BEGIN
			SELECT RAISE(ABORT, 'simulated spoke translation outbox failure');
		END
	`).Error)

	_, _, err = productService.CopyAdminProductTranslation(source.ID, "fr")
	require.ErrorContains(t, err, "simulated spoke translation outbox failure")

	var translatedProductCount int64
	require.NoError(t, db.Model(&product.Product{}).Where("parent_id = ? AND locale = ?", source.ID, "fr").Count(&translatedProductCount).Error)
	assert.Zero(t, translatedProductCount)
	var outboxEventCount int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).Count(&outboxEventCount).Error)
	assert.Zero(t, outboxEventCount)
}

func TestSpokeRootSpecificationUpdateRollsBackAllTranslationsWhenMerchantOutboxWriteFails(t *testing.T) {
	db, productService := newTestProductService(t)
	spokesCategory := createSpokesProductTranslationRegressionCategory(t, db)
	spokeProductType := product.ProductTypeSpoke
	originalSpecification := &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray", ManufacturerFamily: "Original family"}
	source, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductType:        &spokeProductType,
		SpokeSpecification: originalSpecification,
		ProductCategoryID:  &spokesCategory.ID,
		Name:               "SAPIM CX-Ray Spoke",
		Slug:               "sapim-cx-ray-spoke-update-outbox-rollback",
		Status:             "inactive",
		Locale:             "en",
	})
	require.NoError(t, err)
	configureSpokeProductTranslationTransactionalOutboxForTest(t, db, productService)
	translated, _, err := productService.CopyAdminProductTranslation(source.ID, "fr")
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER reject_spoke_product_specification_merchant_event
		BEFORE INSERT ON outbox_events
		WHEN NEW.event_type = 'merchant.product_withdraw'
		BEGIN
			SELECT RAISE(ABORT, 'simulated spoke specification Merchant outbox failure');
		END
	`).Error)

	updatedSpecification := &product.SpokesProductSpecification{SchemaVersion: 1, ModelName: "CX-Ray", ManufacturerFamily: "Updated family"}
	_, err = productService.UpdateAdminProduct(source.ID, ProductUpdateInput{
		UpdateSpokeSpecification: true,
		SpokeSpecification:       updatedSpecification,
	})
	require.ErrorContains(t, err, "simulated spoke specification Merchant outbox failure")

	unchangedSource, err := productService.GetAdminProduct(source.ID)
	require.NoError(t, err)
	unchangedTranslation, err := productService.GetAdminProduct(translated.ID)
	require.NoError(t, err)
	assert.Equal(t, originalSpecification, unchangedSource.SpokeSpecification)
	assert.Equal(t, originalSpecification, unchangedTranslation.SpokeSpecification)
	var specificationChangeEvents int64
	require.NoError(t, db.Model(&outboxdomain.Event{}).
		Where("event_type = ? AND payload LIKE ?", outboxdomain.EventTypeProductCacheInvalidate, "%admin product update%").
		Count(&specificationChangeEvents).Error)
	assert.Zero(t, specificationChangeEvents, "cache invalidation outbox writes must roll back with failed Merchant notifications")
}

func createSpokesProductTranslationRegressionCategory(t *testing.T, db *gorm.DB) product.ProductCategory {
	t.Helper()
	wheelComponents := seedProductCategoryForProductServiceTest(t, db, "Wheel Components", SystemProductCategoryWheelComponentsSlug, nil)
	return seedProductCategoryForProductServiceTest(t, db, "Spokes", SystemProductCategorySpokesSlug, &wheelComponents.ID)
}

func configureSpokeProductTranslationTransactionalOutboxForTest(t *testing.T, db *gorm.DB, productService *ProductService) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))
	productRepository := repository.NewProductRepository(db)
	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		productRepository,
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigureOutboxRepository(repository.NewOutboxRepository(db))
	productService.ConfigureTxManager(txManager)
}
