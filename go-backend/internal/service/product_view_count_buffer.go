package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const productViewCountRedisKey = "product:view_counts"
const productViewCountProcessingRedisKey = productViewCountRedisKey + ":processing"

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
	if batchSize > 5000 {
		batchSize = 5000
	}
	return &ProductViewCountBuffer{client: client, repo: repo, batchSize: batchSize}
}

func (b *ProductViewCountBuffer) Increment(ctx context.Context, id uint) error {
	if b == nil || b.client == nil || id == 0 || uint64(id) > math.MaxInt64 {
		return fmt.Errorf("product view count buffer is not configured")
	}
	return b.client.HIncrBy(ctx, productViewCountRedisKey, strconv.FormatUint(uint64(id), 10), 1).Err()
}

// Flush keeps a stable processing batch across SQL/Redis failures. SQL claims
// and increments commit together; acknowledgement checks the original batch
// and exact fields. A scheduler lease is an optimization, not correctness.
func (b *ProductViewCountBuffer) Flush(ctx context.Context) (int, error) {
	if b == nil || b.client == nil || b.repo == nil {
		return 0, nil
	}
	claimed, err := productViewCountClaimImmutableProcessingBatchScript.Run(ctx, b.client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), uuid.NewString(), b.batchSize).Slice()
	if err != nil {
		return 0, err
	}
	if len(claimed) != 2 {
		return 0, fmt.Errorf("product view processing claim has an invalid shape")
	}
	batchID, ok := claimed[0].(string)
	if !ok {
		return 0, fmt.Errorf("product view processing batch identity is invalid")
	}
	if batchID == "" {
		return 0, nil
	}
	counts, err := parseCanonicalProductViewCountRedisHashEntries(claimed[1])
	if err != nil {
		return 0, err
	}
	if len(counts) == 0 {
		return 0, fmt.Errorf("product view processing scan returned no entries; retry without clearing state")
	}
	ids := make([]uint, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if len(ids) > b.batchSize {
		for _, id := range ids[b.batchSize:] {
			delete(counts, id)
		}
		ids = ids[:b.batchSize]
	}
	applied, err := b.repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(ctx, batchID, counts)
	if err != nil {
		return 0, err
	}
	arguments := []any{batchID}
	for _, id := range ids {
		arguments = append(arguments, strconv.FormatUint(uint64(id), 10), strconv.FormatInt(counts[id], 10))
	}
	if err := productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Run(ctx, b.client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), arguments...).Err(); err != nil {
		return applied, fmt.Errorf("product view SQL batch committed; Redis acknowledgement requires retry: %w", err)
	}
	return applied, nil
}

func parseCanonicalProductViewCountRedisHashEntries(raw any) (map[uint]int64, error) {
	entries, ok := raw.([]interface{})
	if !ok || len(entries)%2 != 0 {
		return nil, fmt.Errorf("product view processing hash entries have an invalid shape")
	}
	counts := make(map[uint]int64, len(entries)/2)
	for index := 0; index < len(entries); index += 2 {
		rawID, idOK := entries[index].(string)
		rawDelta, deltaOK := entries[index+1].(string)
		id, idError := strconv.ParseUint(rawID, 10, 64)
		delta, deltaError := strconv.ParseInt(rawDelta, 10, 64)
		if !idOK || !deltaOK || idError != nil || deltaError != nil || id == 0 || id > math.MaxInt64 || uint64(uint(id)) != id || delta <= 0 || strconv.FormatUint(id, 10) != rawID || strconv.FormatInt(delta, 10) != rawDelta {
			return nil, fmt.Errorf("product view batch has noncanonical or invalid product IDs/deltas; preserve it for reconciliation")
		}
		if _, duplicate := counts[uint(id)]; duplicate {
			return nil, fmt.Errorf("product view batch has duplicate product identities")
		}
		counts[uint(id)] = delta
	}
	return counts, nil
}
