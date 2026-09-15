package service

import (
	"context"
	"fmt"
	"strconv"

	"commerce-platform/internal/repository"
	"github.com/redis/go-redis/v9"
)

const productViewCountRedisKey = "product:view_counts"

// ProductViewCountBuffer records product views in Redis and periodically flushes
// deltas to PostgreSQL in batches. Redis remains available across API instances.
type ProductViewCountBuffer struct {
	client    redis.UniversalClient
	repo      *repository.ProductRepository
	batchSize int
}

func NewProductViewCountBuffer(client redis.UniversalClient, repo *repository.ProductRepository, batchSize int) *ProductViewCountBuffer {
	if batchSize <= 0 {
		batchSize = 500
	}
	return &ProductViewCountBuffer{client: client, repo: repo, batchSize: batchSize}
}

func (b *ProductViewCountBuffer) Increment(ctx context.Context, id uint) error {
	if b == nil || b.client == nil || id == 0 {
		return fmt.Errorf("product view count buffer is not configured")
	}
	return b.client.HIncrBy(ctx, productViewCountRedisKey, strconv.FormatUint(uint64(id), 10), 1).Err()
}

// Flush atomically detaches the active hash before writing to the database.
// New requests continue accumulating in the active hash while this runs.
func (b *ProductViewCountBuffer) Flush(ctx context.Context) (int, error) {
	if b == nil || b.client == nil || b.repo == nil {
		return 0, nil
	}
	processingKey := productViewCountRedisKey + ":processing"
	processingExists, err := b.client.Exists(ctx, processingKey).Result()
	if err != nil {
		return 0, err
	}
	if processingExists == 0 {
		exists, err := b.client.Exists(ctx, productViewCountRedisKey).Result()
		if err != nil || exists == 0 {
			return 0, err
		}
		if err := b.client.Rename(ctx, productViewCountRedisKey, processingKey).Err(); err != nil {
			return 0, err
		}
	}
	values, err := b.client.HGetAll(ctx, processingKey).Result()
	if err != nil {
		_ = b.merge(ctx, processingKey, nil)
		return 0, err
	}
	counts := make(map[uint]int64, minViewInt(len(values), b.batchSize))
	for rawID, rawDelta := range values {
		if len(counts) >= b.batchSize {
			break
		}
		id, idErr := strconv.ParseUint(rawID, 10, 64)
		delta, deltaErr := strconv.ParseInt(rawDelta, 10, 64)
		if idErr != nil || deltaErr != nil || id == 0 || delta <= 0 {
			continue
		}
		counts[uint(id)] = delta
	}
	if err := b.repo.IncrementViewCounts(ctx, counts); err != nil {
		_ = b.merge(ctx, processingKey, nil)
		return 0, err
	}
	for rawID := range values {
		if _, ok := counts[parseID(rawID)]; ok {
			_ = b.client.HDel(ctx, processingKey, rawID).Err()
		}
	}
	if remaining, _ := b.client.HLen(ctx, processingKey).Result(); remaining == 0 {
		_ = b.client.Del(ctx, processingKey).Err()
	}
	return len(counts), nil
}

func (b *ProductViewCountBuffer) merge(ctx context.Context, processingKey string, _ map[string]string) error {
	values, err := b.client.HGetAll(ctx, processingKey).Result()
	if err != nil {
		return err
	}
	for field, value := range values {
		delta, e := strconv.ParseInt(value, 10, 64)
		if e == nil && delta > 0 {
			_ = b.client.HIncrBy(ctx, productViewCountRedisKey, field, delta).Err()
		}
	}
	return b.client.Del(ctx, processingKey).Err()
}

func parseID(raw string) uint { id, _ := strconv.ParseUint(raw, 10, 64); return uint(id) }
func minViewInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
