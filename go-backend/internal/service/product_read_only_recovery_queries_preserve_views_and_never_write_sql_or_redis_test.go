package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductReadOnlyRecoveryQueriesPreserveViewsAndNeverAttemptSQLOrRedisWrites(t *testing.T) {
	db, _ := newTestProductService(t)
	item := product.Product{Name: "Read-only recovery product", Slug: "read-only-recovery-product", Locale: "en", Status: "active", Currency: "USD", PriceMinor: 12345, ViewCount: 17}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&product.ProductVariant{ProductID: item.ID, SKU: "READ-ONLY-RECOVERY-SKU", Currency: "USD", PriceMinor: 12345, Stock: 7, IsActive: true, IsDefault: true}).Error)
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer client.Close()
	readOnly := NewProductServiceForReadOnlyRecoveryVerification(repository.NewProductRepository(db))
	// Even an accidental telemetry configuration cannot enable view writes.
	readOnly.ConfigureProductViewCountBuffer(client, 10)
	updates := 0
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register("count_read_only_recovery_update_attempts", func(*gorm.DB) { updates++ }))
	_, err := readOnly.GetByID(item.ID)
	require.NoError(t, err)
	_, err = readOnly.GetBySlug(item.Slug, item.Locale)
	require.NoError(t, err)
	_, err = readOnly.GetPublicByID(item.ID)
	require.NoError(t, err)
	resolved, routes, err := readOnly.GetPublicBySlugWithRoutesContext(context.Background(), item.Slug, item.Locale)
	require.NoError(t, err)
	require.Equal(t, item.ID, resolved.ID)
	require.NotEmpty(t, routes)
	items, total, err := readOnly.ListPublic("en", false, 1, 24)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)
	items, total, err = readOnly.SearchPublic(ProductSearchInput{Locale: "en", Keyword: "recovery", Page: 1, PageSize: 24})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 1, total)
	var stored product.Product
	require.NoError(t, db.First(&stored, item.ID).Error)
	require.Equal(t, item.ViewCount, stored.ViewCount)
	require.Zero(t, updates)
	require.Empty(t, server.Keys())
}
