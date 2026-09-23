package service

import (
	"context"
	"strconv"
	"testing"

	"commerce-platform/internal/domain/product"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestProductViewCountBufferFlushesRedisDeltas(t *testing.T) {
	db, productService := newTestProductService(t)
	record := product.Product{SKU: "VIEW-BUFFER-001", Name: "Buffered", Slug: "buffered", Currency: "USD", PriceMinor: 1000, Status: "active", Locale: "en"}
	require.NoError(t, db.Create(&record).Error)
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	productService.ConfigureProductViewCountBuffer(client, 100)

	_, err := productService.GetPublicByIDContext(context.Background(), record.ID)
	require.NoError(t, err)
	var before int
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Pluck("view_count", &before).Error)
	require.Equal(t, 0, before)
	require.Equal(t, "1", mini.HGet(productViewCountRedisKey, strconv.FormatUint(uint64(record.ID), 10)))

	_, err = productService.FlushProductViewCounts(context.Background())
	require.NoError(t, err)
	var after int
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", record.ID).Pluck("view_count", &after).Error)
	require.Equal(t, 1, after)
}
