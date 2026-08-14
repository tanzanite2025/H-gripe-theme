package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductCacheOutboxPublisherCreatesInvalidateEvent(t *testing.T) {
	db := newProductCacheOutboxTestDB(t)
	publisher := NewProductCacheOutboxPublisher(repository.NewOutboxRepository(db))

	require.NoError(t, publisher.EnqueueProductCacheInvalidateByIDs([]uint{3, 3, 1}, "stock changed"))

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_type = ?", outboxdomain.EventTypeProductCacheInvalidate).First(&event).Error)
	assert.Equal(t, outboxdomain.AggregateTypeProductCache, event.AggregateType)
	assert.Contains(t, event.EventKey, "product.cache_invalidate")
	assert.Contains(t, event.EventKey, "stock_changed")

	var payload outboxdomain.ProductCacheInvalidatePayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, []uint{3, 1}, payload.ProductIDs)
	assert.Equal(t, "stock changed", payload.Reason)
}

func TestProductCacheOutboxHandlerInvalidatesProductIDs(t *testing.T) {
	executor := &recordingProductCacheInvalidationExecutor{}
	handler := NewProductCacheOutboxHandler(executor)
	payload, err := json.Marshal(outboxdomain.ProductCacheInvalidatePayload{
		ProductIDs: []uint{7, 8},
		Reason:     "stock changed",
	})
	require.NoError(t, err)

	err = handler.Handle(context.Background(), outboxdomain.Event{
		EventType: outboxdomain.EventTypeProductCacheInvalidate,
		Payload:   payload,
	})

	require.NoError(t, err)
	assert.Equal(t, []uint{7, 8}, executor.productIDs)
	assert.Equal(t, []string{productCacheInvalidationSourceOutbox}, executor.sources)
}

func TestProductCacheOutboxHandlerReturnsInvalidationErrorForRetry(t *testing.T) {
	executor := &recordingProductCacheInvalidationExecutor{err: errors.New("redis unavailable")}
	handler := NewProductCacheOutboxHandler(executor)
	payload, err := json.Marshal(outboxdomain.ProductCacheInvalidatePayload{ProductIDs: []uint{7}})
	require.NoError(t, err)

	err = handler.Handle(context.Background(), outboxdomain.Event{
		EventType: outboxdomain.EventTypeProductCacheInvalidate,
		Payload:   payload,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis unavailable")
}

func TestProductCacheOutboxHandlerProcessesThroughOutboxService(t *testing.T) {
	db := newProductCacheOutboxTestDB(t)
	productRecord := product.Product{SKU: "CACHE-OUTBOX-1", Slug: "cache-outbox-1", Locale: "en", Name: "Cache Outbox", Status: "active"}
	require.NoError(t, db.Create(&productRecord).Error)
	repo := &fakeProductCacheIdentityRepository{
		productsByID: map[uint]product.Product{
			productRecord.ID: {ID: productRecord.ID, Slug: productRecord.Slug, Locale: productRecord.Locale},
		},
	}
	cache := &recordingProductDetailCache{}
	handler := NewProductCacheOutboxHandler(NewProductDetailCacheInvalidator(repo, cache))
	service := NewOutboxService(repository.NewOutboxRepository(db))
	service.RegisterHandler(outboxdomain.EventTypeProductCacheInvalidate, handler.Handle)
	publisher := NewProductCacheOutboxPublisher(repository.NewOutboxRepository(db))
	require.NoError(t, publisher.EnqueueProductCacheInvalidateByIDs([]uint{productRecord.ID}, "stock changed"))

	result, err := service.ProcessPending(context.Background(), time.Now().UTC().Add(time.Second), 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Processed)
	assert.Contains(t, cache.deletedKeys, productIDCacheKey(productRecord.ID))
	assert.Contains(t, cache.deletedKeys, productSlugCacheKey(productRecord.Slug, "en"))
}

type recordingProductCacheInvalidationExecutor struct {
	productIDs                    []uint
	productTypeIDs                []uint
	productInformationTemplateIDs []uint
	sources                       []string
	err                           error
}

func (e *recordingProductCacheInvalidationExecutor) InvalidateProductCacheByIDsWithSource(ids []uint, source string) (ProductCacheInvalidationResult, error) {
	e.productIDs = append(e.productIDs, ids...)
	e.sources = append(e.sources, source)
	return ProductCacheInvalidationResult{Products: len(ids), Keys: len(ids)}, e.err
}

func (e *recordingProductCacheInvalidationExecutor) InvalidateProductCacheByProductTypeIDWithSource(productTypeID uint, source string) (ProductCacheInvalidationResult, error) {
	e.productTypeIDs = append(e.productTypeIDs, productTypeID)
	e.sources = append(e.sources, source)
	return ProductCacheInvalidationResult{Products: 1, Keys: 1}, e.err
}

func (e *recordingProductCacheInvalidationExecutor) InvalidateProductCacheByInformationTemplateIDWithSource(templateID uint, source string) (ProductCacheInvalidationResult, error) {
	e.productInformationTemplateIDs = append(e.productInformationTemplateIDs, templateID)
	e.sources = append(e.sources, source)
	return ProductCacheInvalidationResult{Products: 1, Keys: 1}, e.err
}

func newProductCacheOutboxTestDB(t *testing.T) *gorm.DB {
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
	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}, &product.Product{}))
	return db
}
