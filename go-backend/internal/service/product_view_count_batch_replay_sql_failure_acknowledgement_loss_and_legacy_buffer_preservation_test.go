package service

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type rejectProductViewRedisAcknowledgementAfterSQLCommitHook struct{ enabled atomic.Bool }

func (hook *rejectProductViewRedisAcknowledgementAfterSQLCommitHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return next(ctx, network, address)
	}
}
func (hook *rejectProductViewRedisAcknowledgementAfterSQLCommitHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (hook *rejectProductViewRedisAcknowledgementAfterSQLCommitHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, command redis.Cmder) error {
		arguments := command.Args()
		if hook.enabled.Load() && (command.Name() == "evalsha" || command.Name() == "eval") && len(arguments) > 1 && (arguments[1] == productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Hash() || arguments[1] == productViewCountAcknowledgeCommittedBatchLua) {
			return errors.New("synthetic Redis acknowledgement unavailable after SQL commit")
		}
		return next(ctx, command)
	}
}

func createProductViewCountBatchReplayTestEnvironment(t *testing.T) (*gorm.DB, *ProductViewCountBuffer, *miniredis.Miniredis, *redis.Client, product.Product) {
	t.Helper()
	db, _ := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&product.ProductViewCountFlushReceipt{}))
	record := product.Product{Name: "View batch replay", Slug: "view-batch-" + uuid.NewString(), Currency: "USD", PriceMinor: 100, Status: "active", Locale: "en"}
	require.NoError(t, db.Create(&record).Error)
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	return db, NewProductViewCountBuffer(client, repository.NewProductRepository(db), 2), mini, client, record
}

func requireProductViewSQLCountAndReplayReceipts(t *testing.T, db *gorm.DB, productID uint, expectedViews int, expectedReceipts int64) {
	t.Helper()
	var actualViews int
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", productID).Pluck("view_count", &actualViews).Error)
	require.Equal(t, expectedViews, actualViews)
	var receipts int64
	require.NoError(t, db.Model(&product.ProductViewCountFlushReceipt{}).Count(&receipts).Error)
	require.Equal(t, expectedReceipts, receipts)
}

func TestProductViewCountCommittedSQLBatchSurvivesLostRedisAcknowledgementAndBufferRestart(t *testing.T) {
	db, buffer, mini, client, record := createProductViewCountBatchReplayTestEnvironment(t)
	ctx := t.Context()
	hook := &rejectProductViewRedisAcknowledgementAfterSQLCommitHook{}
	client.AddHook(hook)
	hook.enabled.Store(true)
	require.NoError(t, buffer.Increment(ctx, record.ID))
	applied, err := buffer.Flush(ctx)
	require.Equal(t, 1, applied)
	require.ErrorContains(t, err, "SQL batch committed")
	originalBatch, err := client.Get(ctx, productViewCountProcessingBatchIdentityRedisKey).Result()
	require.NoError(t, err)
	requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 1, 1)
	require.NoError(t, buffer.Increment(ctx, record.ID))
	require.NoError(t, buffer.Increment(ctx, record.ID))
	report, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), report.ActiveDeltaTotal)
	require.Equal(t, int64(1), report.ProcessingDeltaTotal)
	require.Equal(t, 1, report.CommittedProcessingReceipts)
	require.Zero(t, report.PendingProcessingProducts)
	require.False(t, report.ReleaseAvailable)
	require.True(t, report.StoppedRevalidationRequired)
	repeated, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(ctx)
	require.NoError(t, err)
	require.Equal(t, report, repeated)
	require.Equal(t, "1", mini.HGet(productViewCountProcessingRedisKey, strconv.FormatUint(uint64(record.ID), 10)))
	currentBatch, err := client.Get(ctx, productViewCountProcessingBatchIdentityRedisKey).Result()
	require.NoError(t, err)
	require.Equal(t, originalBatch, currentBatch)
	hook.enabled.Store(false)
	restarted := NewProductViewCountBuffer(client, repository.NewProductRepository(db), 2)
	applied, err = restarted.Flush(ctx)
	require.NoError(t, err)
	require.Zero(t, applied, "the committed batch must be acknowledged without incrementing SQL again")
	require.False(t, mini.Exists(productViewCountProcessingBatchIdentityRedisKey))
	require.Equal(t, "2", mini.HGet(productViewCountRedisKey, strconv.FormatUint(uint64(record.ID), 10)))
	applied, err = restarted.Flush(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, applied)
	requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 3, 2)
	empty, err := restarted.InspectReadOnlyProductViewRecoveryBoundary(ctx)
	require.NoError(t, err)
	require.True(t, empty.RedisBuffersEmpty)
	require.False(t, empty.ReleaseAvailable, "empty Redis alone cannot prove stopped writers or approve a recovery")
}

func TestProductViewCountSQLFailureRollsBackReceiptAndPreservesOriginalBatchForRetry(t *testing.T) {
	db, buffer, mini, client, record := createProductViewCountBatchReplayTestEnvironment(t)
	callbackName := "reject_product_view_update_" + uuid.NewString()
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "products" {
			tx.AddError(errors.New("synthetic product update failed"))
		}
	}))
	defer db.Callback().Update().Remove(callbackName)
	require.NoError(t, buffer.Increment(t.Context(), record.ID))
	_, err := buffer.Flush(t.Context())
	require.ErrorContains(t, err, "synthetic product update failed")
	originalBatch, err := client.Get(t.Context(), productViewCountProcessingBatchIdentityRedisKey).Result()
	require.NoError(t, err)
	requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 0, 0)
	require.False(t, mini.Exists(productViewCountRedisKey), "failed processing must not merge into active under a new identity")
	report, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, report.PendingProcessingProducts)
	require.NoError(t, db.Callback().Update().Remove(callbackName))
	_, err = buffer.Flush(t.Context())
	require.NoError(t, err)
	requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 1, 1)
	var receipt product.ProductViewCountFlushReceipt
	require.NoError(t, db.Take(&receipt).Error)
	require.Equal(t, originalBatch, receipt.BatchID)
}

func TestProductViewCountLegacyProcessingAndMalformedRedisStateArePreservedForReconciliation(t *testing.T) {
	for _, scenario := range []string{"legacy_processing", "orphan_identity", "active_wrong_type", "active_expires", "processing_expires", "noncanonical_id", "invalid_delta", "malformed_batch_id"} {
		t.Run(scenario, func(t *testing.T) {
			db, buffer, mini, client, record := createProductViewCountBatchReplayTestEnvironment(t)
			id := strconv.FormatUint(uint64(record.ID), 10)
			mini.HSet(productViewCountRedisKey, id, "2")
			switch scenario {
			case "legacy_processing":
				mini.HSet(productViewCountProcessingRedisKey, id, "3")
			case "orphan_identity":
				mini.Set(productViewCountProcessingBatchIdentityRedisKey, uuid.NewString())
			case "active_wrong_type":
				mini.Del(productViewCountRedisKey)
				mini.Set(productViewCountRedisKey, "preserve unexpected object")
			case "active_expires":
				mini.SetTTL(productViewCountRedisKey, time.Hour)
			case "processing_expires":
				mini.HSet(productViewCountProcessingRedisKey, id, "3")
				mini.Set(productViewCountProcessingBatchIdentityRedisKey, uuid.NewString())
				mini.SetTTL(productViewCountProcessingRedisKey, time.Hour)
			case "noncanonical_id":
				mini.HSet(productViewCountRedisKey, "01", "4")
			case "invalid_delta":
				mini.HSet(productViewCountRedisKey, id, "-2")
			case "malformed_batch_id":
				mini.HSet(productViewCountProcessingRedisKey, id, "3")
				mini.Set(productViewCountProcessingBatchIdentityRedisKey, "not-a-batch-uuid")
			}
			_, err := buffer.Flush(t.Context())
			require.Error(t, err)
			requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 0, 0)
			if scenario == "noncanonical_id" || scenario == "invalid_delta" {
				require.True(t, mini.Exists(productViewCountProcessingRedisKey))
			} else {
				require.True(t, mini.Exists(productViewCountRedisKey))
			}
			if scenario == "legacy_processing" {
				require.False(t, mini.Exists(productViewCountProcessingBatchIdentityRedisKey))
				_, err = buffer.InspectReadOnlyProductViewRecoveryBoundary(t.Context())
				require.ErrorContains(t, err, "legacy_product_view_processing_batch")
			}
			_ = client
		})
	}
}

func TestProductViewCountStaleAcknowledgementCannotDeleteANewerProcessingBatch(t *testing.T) {
	db, buffer, mini, client, record := createProductViewCountBatchReplayTestEnvironment(t)
	ctx := t.Context()
	id := strconv.FormatUint(uint64(record.ID), 10)
	require.NoError(t, buffer.Increment(ctx, record.ID))
	first, err := productViewCountClaimImmutableProcessingBatchScript.Run(ctx, client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), uuid.NewString(), 2).Slice()
	require.NoError(t, err)
	counts := map[uint]int64{record.ID: 1}
	_, err = buffer.repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(ctx, first[0].(string), counts)
	require.NoError(t, err)
	require.NoError(t, productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Run(ctx, client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), first[0], id, "1").Err())
	require.NoError(t, buffer.Increment(ctx, record.ID))
	second, err := productViewCountClaimImmutableProcessingBatchScript.Run(ctx, client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), uuid.NewString(), 2).Slice()
	require.NoError(t, err)
	require.NotEqual(t, first[0], second[0])
	acknowledged, err := productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Run(ctx, client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), first[0], id, "1").Int()
	require.NoError(t, err)
	require.Zero(t, acknowledged)
	require.Equal(t, "1", mini.HGet(productViewCountProcessingRedisKey, id))
	_, err = buffer.Flush(ctx)
	require.NoError(t, err)
	requireProductViewSQLCountAndReplayReceipts(t, db, record.ID, 2, 2)
}

func TestProductViewCountAcknowledgementChecksAllDeltasBeforeDeletingAnyField(t *testing.T) {
	_, _, mini, client, _ := createProductViewCountBatchReplayTestEnvironment(t)
	batchID := uuid.NewString()
	mini.HSet(productViewCountProcessingRedisKey, "1", "2", "2", "3")
	mini.Set(productViewCountProcessingBatchIdentityRedisKey, batchID)
	err := productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Run(t.Context(), client, productViewCountActiveProcessingAndBatchIdentityRedisKeys(), batchID, "1", "2", "2", "4").Err()
	require.ErrorContains(t, err, "payload_changed")
	require.Equal(t, "2", mini.HGet(productViewCountProcessingRedisKey, "1"))
	require.Equal(t, "3", mini.HGet(productViewCountProcessingRedisKey, "2"))
}
